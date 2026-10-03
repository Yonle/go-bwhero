package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	imagesizelimit     int64
	animationsizelimit int64
	videosizelimit     int64
	workers            chan Task
)

const checkerReadLimit = 64 * 1024

type fakeReadCloser struct {
	io.Reader
}

type Task struct {
	W   http.ResponseWriter
	R   *http.Request
	Ctx context.Context

	URL string

	Quality    int
	Grayscale  int
	Anim       bool
	Thumb      bool
	ThumbWidth int

	Done chan struct{}
}

type responseInfo struct {
	IsImage           bool
	IsVideo           bool
	IsAnimated        bool
	IsGIF             bool
	IsAPNG            bool
	IsAnimatedWEBP    bool
	IsBig             bool
	AnimFormat        string
	PotentiallyCamera bool
}

func (t *Task) canclRedir() {
	http.Redirect(t.W, t.R, t.URL, http.StatusFound)
}

func (frc fakeReadCloser) Close() error {
	// do absolutely nothing
	return nil
}

func getQV(query url.Values, key string, def int) int {
	value, err := strconv.Atoi(query.Get(key))
	if err != nil {
		return def
	}

	return value
}

func init() {
	imagesizelimit, _ = strconv.ParseInt(os.Getenv("IMAGESIZELIMIT"), 10, 64)
	animationsizelimit, _ = strconv.ParseInt(os.Getenv("ANIMATIONSIZELIMIT"), 10, 64)
	videosizelimit, _ = strconv.ParseInt(os.Getenv("VIDEOSIZELIMIT"), 10, 64)

	if _, ok := os.LookupEnv("ANIMATIONSIZELIMIT"); !ok {
		animationsizelimit = imagesizelimit
	}

	if _, ok := os.LookupEnv("VIDEOSIZELIMIT"); !ok {
		videosizelimit = animationsizelimit
	}
}

func startWorker(N int) {
	workers = make(chan Task, N)

	for i := 0; i < N; i++ {
		go func(wID int) {
			log.Printf("Worker %d started.", wID)

			for task := range workers {
				task.Process()
				close(task.Done)
			}
		}(i)
	}
}

func request_handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query()

	originURL := query.Get("url")
	if len(originURL) < 7 {
		fmt.Fprint(w, "bandwidth-hero-proxy")
		return
	}

	// user opens in new tab
	if strings.HasPrefix(r.Header.Get("Accept"), "text/html") && !query.Has("nr") {
		http.Redirect(w, r, originURL, http.StatusFound)
		return
	}

	grayscale := getQV(query, "bw", 1)

	quality := getQV(query, "l", 80)
	if quality < 1 || quality > 100 {
		quality = 80
	}

	anim := getQV(query, "a", 1) != 0

	thumbWidth := getQV(query, "t", 0)
	thumb := thumbWidth > 0

	task := Task{
		W:   w,
		R:   r,
		Ctx: ctx,

		URL: originURL,

		Quality:    quality,
		Grayscale:  grayscale,
		Anim:       anim,
		Thumb:      thumb,
		ThumbWidth: thumbWidth,

		Done: make(chan struct{}),
	}

	select {
	case workers <- task:
		<-task.Done

	case <-ctx.Done():
		return
	}
}

func inspectResponse(resp *http.Response, head []byte) responseInfo {
	kind := strings.ToLower(resp.Header.Get("Content-Type"))

	gif := isGIF(head)
	apng := strings.Contains(kind, "image/apng") || isAnimatedPNG(head)
	animatedWEBP := isAnimatedWEBP(head)

	info := responseInfo{
		IsImage:           strings.HasPrefix(kind, "image/"),
		IsVideo:           strings.HasPrefix(kind, "video/"),
		IsGIF:             gif,
		IsAPNG:            apng,
		IsAnimatedWEBP:    animatedWEBP,
		IsAnimated:        gif || apng || animatedWEBP,
		PotentiallyCamera: strings.Contains(kind, "jpeg") || strings.Contains(kind, "heic") || strings.Contains(kind, "heif") || strings.Contains(kind, "tiff"),
	}

	// The signature is more trustworthy than a shitty Content-Type.
	if info.IsGIF || info.IsAPNG || info.IsAnimatedWEBP {
		info.IsImage = true
	}

	switch {
	case info.IsAPNG:
		info.AnimFormat = "apng"
	case info.IsAnimatedWEBP:
		info.AnimFormat = "webp"
	default:
		info.AnimFormat = "gif"
	}

	limit := imagesizelimit

	if info.IsAnimated && animationsizelimit > 0 {
		limit = animationsizelimit
	}

	if info.IsVideo && videosizelimit > 0 {
		limit = videosizelimit
	}

	if limit > 0 && resp.ContentLength >= 0 {
		info.IsBig = resp.ContentLength > limit
	}

	// Too big for animation, or animation processing is explicitly disabled.
	if info.IsAnimated && (animationsizelimit == -2 || info.IsBig) {
		info.IsAnimated = false

		if imagesizelimit > 0 && resp.ContentLength >= 0 {
			info.IsBig = resp.ContentLength > imagesizelimit
		} else {
			info.IsBig = false
		}
	}

	// Video thumbnailing disabled.
	if info.IsVideo && videosizelimit == -2 {
		info.IsBig = true
	}

	return info
}

func (t *Task) Process() {
	fetchTime := time.Now()

	resp, err := proxy(t.Ctx, t.R, t.URL)
	if err != nil {
		log.Printf("Failed to fetch %s. Redirecting", t.URL)
		t.canclRedir()
		return
	}

	defer resp.Body.Close()

	// Sniff the response, then put the bytes back so decoders still get the original stream.
	head, err := io.ReadAll(io.LimitReader(resp.Body, checkerReadLimit))
	if err != nil {
		log.Printf("Failed to inspect %s. Redirecting", t.URL)
		t.canclRedir()
		return
	}

	resp.Body = io.NopCloser(io.MultiReader(bytes.NewReader(head), resp.Body))

	kind := strings.ToLower(resp.Header.Get("Content-Type"))

	// SVG gets yeeted regardless of whether the server correctly identifies it.
	if strings.Contains(kind, "image/svg+xml") || isSVG(head) {
		log.Printf("SVG detected on %s. Redirecting", t.URL)
		t.canclRedir()
		return
	}

	info := inspectResponse(resp, head)

	if resp.StatusCode >= 400 || (!info.IsImage && !info.IsVideo) || info.IsBig {
		if resp.StatusCode >= 400 {
			log.Printf("Got status code %d on %s", resp.StatusCode, t.URL)
		} else {
			log.Printf("is an image: %v; is a video: %v; is animated: %v; is big: %v; url: %s", info.IsImage, info.IsVideo, info.IsAnimated, info.IsBig, t.URL)
		}

		t.canclRedir()
		return
	}

	fetchDuration := time.Since(fetchTime)
	processingTime := time.Now()

	var writ io.WriteCloser = &clientWriter{t.W, false}

	if info.IsAnimated && info.IsAPNG {
		h := HeaderToFFmpegFormat(resp.Request.Header)

		if err := process_apng(t.Ctx, writ, t.URL, h, t.ThumbWidth, t.Quality, t.Grayscale); err != nil {
			log.Printf("Failed to convert apng %s: %s", t.URL, err)
			t.canclRedir()
			return
		}

		processingDuration := time.Since(processingTime)
		totalDuration := time.Since(fetchTime)

		log.Printf("apng->webp Took %.1fs | Fetch: %.1fs | Processing: %.1fs | URL: %s", totalDuration.Seconds(), fetchDuration.Seconds(), processingDuration.Seconds(), t.URL)
		return
	}

	if info.IsVideo {
		h := HeaderToFFmpegFormat(resp.Request.Header)

		if err := process_vidthumb(t.Ctx, writ, t.URL, h, t.ThumbWidth, t.Quality, t.Grayscale); err != nil {
			log.Printf("Failed to thumbnail the video %s: %s", t.URL, err)
			t.canclRedir()
			return
		}

		processingDuration := time.Since(processingTime)
		totalDuration := time.Since(fetchTime)

		log.Printf("Thumbnailing Took %.1fs | Fetch: %.1fs | Processing: %.1fs | URL: %s", totalDuration.Seconds(), fetchDuration.Seconds(), processingDuration.Seconds(), t.URL)
		return
	}

	var body io.Reader

	if resp.ContentLength >= 0 {
		body = io.LimitReader(resp.Body, resp.ContentLength)
	} else {
		body = resp.Body
	}

	switch {
	case info.IsAnimated:
		if err := process_anim(t.Ctx, writ, body, info.AnimFormat, t.ThumbWidth, t.Quality, t.Grayscale); err != nil {
			log.Printf("Failed to animate %s: %s", t.URL, err)
			t.canclRedir()
			return
		}

	case t.Thumb:
		if err := process_thumb(t.Ctx, writ, fakeReadCloser{body}, t.ThumbWidth, t.Quality, t.Grayscale); err != nil {
			log.Printf("Failed to process %s: %s", t.URL, err)
			t.canclRedir()
			return
		}

	default:
		if err := process_image(t.Ctx, writ, fakeReadCloser{body}, info.PotentiallyCamera, t.Quality, t.Grayscale); err != nil {
			log.Printf("Failed to process %s: %s", t.URL, err)
			t.canclRedir()
			return
		}
	}

	processingDuration := time.Since(processingTime)
	totalDuration := time.Since(fetchTime)

	log.Printf("Took %.1fs | Fetch: %.1fs | Processing: %.1fs | URL: %s", totalDuration.Seconds(), fetchDuration.Seconds(), processingDuration.Seconds(), t.URL)
}