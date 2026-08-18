package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var imagesizelimit int64
var animationsizelimit int64
var videosizelimit int64
var workers chan Task

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

func (t *Task) canclRedir() {
	http.Redirect(t.W, t.R, t.URL, http.StatusFound)
}

func (frc fakeReadCloser) Close() error {
	// do absolutely nothing
	return nil
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

func request_handler(
	w http.ResponseWriter,
	r *http.Request,
) {
	ctx := r.Context()
	query := r.URL.Query()

	origin_url := query.Get("url")
	if len(origin_url) < 7 {
		fmt.Fprintf(w, "bandwidth-hero-proxy")
		return
	}

	if r.Header.Get("User-Agent") == "go-bwhero" || r.Header.Get("Via") == "2.0 go-bwhero" {
		http.Redirect(w, r, origin_url, http.StatusFound)
		return
	}

	// user opens in new tab
	if strings.HasPrefix(r.Header.Get("Accept"), "text/html") && !query.Has(("nr")) {
		http.Redirect(w, r, origin_url, http.StatusFound)
		return
	}

	grayscale, err := strconv.Atoi(query.Get("bw"))
	if err != nil {
		grayscale = 1
	}

	quality, err := strconv.Atoi(query.Get("l"))
	if err != nil || quality < 1 || quality > 100 {
		quality = 80
	}

	anim := true
	animP, err := strconv.Atoi(query.Get("a"))
	if err == nil && animP == 0 {
		anim = false
	}

	thumb := false
	thumbWidth, err := strconv.Atoi(query.Get("t"))
	if err == nil && thumbWidth > 0 {
		thumb = true
	}

	task := Task{
		W:   w,
		R:   r,
		Ctx: ctx,

		URL: origin_url,

		Grayscale:  grayscale,
		Quality:    quality,
		Anim:       anim,
		Thumb:      thumb,
		ThumbWidth: thumbWidth,

		Done: make(chan struct{}),
	}

	select {
	case workers <- task:
		<-task.Done

	case <-r.Context().Done():
		return
	}
}

func (t *Task) Process() {
	fetch_time := time.Now()

	resp, err := proxy(t.Ctx, t.R, t.URL)
	if err != nil {
		log.Printf("Failed to fetch %s. Redirecting", t.URL)
		http.Redirect(t.W, t.R, t.URL, http.StatusFound)
		return
	}

	defer resp.Body.Close()

	kind := resp.Header.Get("Content-Type")
	isImage := strings.HasPrefix(kind, "image/")
	isVideo := strings.HasPrefix(kind, "video/")

	// animation
	isGIF := strings.Contains(kind, "image/gif")
	isAPNG := strings.Contains(kind, "image/apng")
	isAnimated := (t.Anim && (isGIF || isAPNG))

	// camera / printer
	potentiallyCamera :=
		strings.Contains(kind, "jpeg") ||
			strings.Contains(kind, "heic") ||
			strings.Contains(kind, "heif") ||
			strings.Contains(kind, "tiff")

	var isBig bool
	var animformat string = "gif"

	if isGIF {
		animformat = "gif"
	} else if isAPNG {
		animformat = "apng"
	}

	limit := imagesizelimit

	if isAnimated && animationsizelimit > 0 {
		limit = animationsizelimit
	}

	if isVideo && videosizelimit > 0 {
		limit = videosizelimit
	}

	// if limit is set
	if limit > 0 {
		isBig = resp.ContentLength > limit
	}

	// if it's too big for animation OR we forced a downgrade
	if (isAnimated && animationsizelimit == -2) || (isAnimated && isBig) {
		isAnimated = false
		// Re-check size against image limit if we just downgraded from animation
		if imagesizelimit > 0 {
			isBig = resp.ContentLength > imagesizelimit
		} else {
			// we got no limit being set.
			isBig = false
		}
	}

	// if video thumbnailing is disabled, do not process at all
	if isVideo && videosizelimit == -2 {
		isBig = true
	}

	if resp.StatusCode >= 400 || (!isImage && !isVideo) || isBig {
		if resp.StatusCode >= 400 {
			log.Printf("Got status code %d on %s", resp.StatusCode, t.URL)
		} else {
			log.Printf("is an image: %v; is an video: %v; is big: %v; url: %s", isImage, isVideo, isBig, t.URL)
		}
		resp.Body.Close()
		t.canclRedir()
		return
	}

	ft := time.Since(fetch_time)

	processing_time := time.Now()
	var writ io.WriteCloser = &clientWriter{t.W, false}

	if isAnimated && isAPNG {
		// immediately close the body. We won't use it.
		resp.Body.Close()

		h := HeaderToFFmpegFormat(resp.Request.Header)
		if err := process_apng(t.Ctx, writ, t.URL, h, t.ThumbWidth, t.Quality, t.Grayscale); err != nil {
			log.Printf("Failed to convert apng %s: %s", t.URL, err)
			t.canclRedir()
			return
		}

		pt := time.Since(processing_time)
		tl := time.Since(fetch_time)

		log.Printf("apng->webp Took %.1fs | Fetch: %.1fs | Processing: %.1fs | URL: %s", tl.Seconds(), ft.Seconds(), pt.Seconds(), t.URL)

		return
	}

	if isVideo {
		// immediately close the body. We won't use it.
		resp.Body.Close()

		h := HeaderToFFmpegFormat(resp.Request.Header)
		if err := process_vidthumb(t.Ctx, writ, t.URL, h, t.ThumbWidth, t.Quality, t.Grayscale); err != nil {
			log.Printf("Failed to thumbnail the video %s: %s", t.URL, err)
			t.canclRedir()
			return
		}

		pt := time.Since(processing_time)
		tl := time.Since(fetch_time)

		log.Printf("Thumbnailing Took %.1fs | Fetch: %.1fs | Processing: %.1fs | URL: %s", tl.Seconds(), ft.Seconds(), pt.Seconds(), t.URL)

		return
	}

	body := io.LimitReader(resp.Body, resp.ContentLength)

	if isAnimated {
		if err := process_anim(t.Ctx, writ, body, animformat, t.ThumbWidth, t.Quality, t.Grayscale); err != nil {
			log.Printf("Failed to animate %s: %s", t.URL, err)
			t.canclRedir()
			return
		}
	} else if t.Thumb {
		if err := process_thumb(t.Ctx, writ, fakeReadCloser{body}, t.ThumbWidth, t.Quality, t.Grayscale); err != nil {
			log.Printf("Failed to process %s: %s", t.URL, err)
			t.canclRedir()
			return
		}
	} else {
		if err := process_image(t.Ctx, writ, fakeReadCloser{body}, potentiallyCamera, t.Quality, t.Grayscale); err != nil {
			log.Printf("Failed to process %s: %s", t.URL, err)
			t.canclRedir()
			return
		}
	}

	pt := time.Since(processing_time)
	tl := time.Since(fetch_time)

	log.Printf("Took %.1fs | Fetch: %.1fs | Processing: %.1fs | URL: %s", tl.Seconds(), ft.Seconds(), pt.Seconds(), t.URL)
}
