package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"strings"
	"time"
)

var imagesizelimit int64
var animationsizelimit int64
var videosizelimit int64

type fakeReadCloser struct {
	io.Reader
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

	fetch_time := time.Now()

	resp, err := proxy(ctx, r, origin_url)
	if err != nil {
		log.Printf("Failed to fetch %s. Redirecting", origin_url)
		http.Redirect(w, r, origin_url, http.StatusFound)
		return
	}

	defer resp.Body.Close()

	kind := resp.Header.Get("Content-Type")
	isImage := strings.HasPrefix(kind, "image/")
	isVideo := strings.HasPrefix(kind, "video/")

	// animation
	isGIF := strings.Contains(kind, "image/gif")
	//isAPNG := strings.Contains(kind, "image/apng")
	isAnimated := isGIF //|| isAPNG

	// camera / printer
	potentiallyCamera :=
		strings.Contains(kind, "jpeg") ||
			strings.Contains(kind, "heic") ||
			strings.Contains(kind, "heif") ||
			strings.Contains(kind, "tiff")

	var isBig bool
	var animformat string = "gif"

	/*if isGIF {
		animformat = "gif"
	} else if isAPNG {
		animformat = "apng"
	}*/

	limit := imagesizelimit

	if isAnimated && animationsizelimit != -1 {
		limit = animationsizelimit
	}

	if limit > 0 {
		isBig = resp.ContentLength > limit
	}

	// if it's too big for animation OR we forced a downgrade
	if (isAnimated && animationsizelimit == -1) || (isAnimated && isBig) {
		isAnimated = false
		// Re-check size against image limit if we just downgraded from animation
		if imagesizelimit > 0 {
			isBig = resp.ContentLength > imagesizelimit
		} else {
			// we got no limit being set.
			isBig = false
		}
	}

	if resp.StatusCode >= 400 || (!isImage && !isVideo) || isBig {
		if resp.StatusCode >= 400 {
			log.Printf("Got status code %d on %s", resp.StatusCode, origin_url)
		} else {
			log.Printf("is an image: %v; is big: %v; url: %s", isImage, isBig, origin_url)
		}
		resp.Body.Close()
		http.Redirect(w, r, origin_url, http.StatusFound)
		return
	}

	ft := time.Since(fetch_time)

	processing_time := time.Now()
	var writ io.WriteCloser = &clientWriter{w, false}

	if isVideo {
		// immediately close the body. We won't use it.
		resp.Body.Close()

		h := HeaderToFFmpegFormat(resp.Request.Header)
		if err := process_vidthumb(r.Context(), writ, origin_url, h, quality, grayscale); err != nil {
			log.Printf("Failed to thumbnail the video %s: %s", origin_url, err)
			http.Redirect(w, r, origin_url, http.StatusFound)
			return
		}

		pt := time.Since(processing_time)
		tl := time.Since(fetch_time)

		log.Printf("Thumbnailing Took %.1fs | Fetch: %.1fs | Processing: %.1fs | URL: %s", tl.Seconds(), ft.Seconds(), pt.Seconds(), origin_url)

		return
	}

	body := io.LimitReader(resp.Body, resp.ContentLength)

	if isAnimated {
		if err := process_anim(r.Context(), writ, body, animformat, quality, grayscale); err != nil {
			log.Printf("Failed to animate %s: %s", origin_url, err)
			http.Redirect(w, r, origin_url, http.StatusFound)
			return
		}
	} else {
		if err := process_image(r.Context(), writ, fakeReadCloser{body}, potentiallyCamera, quality, grayscale); err != nil {
			log.Printf("Failed to process %s: %s", origin_url, err)
			http.Redirect(w, r, origin_url, http.StatusFound)
			return
		}
	}

	pt := time.Since(processing_time)
	tl := time.Since(fetch_time)

	log.Printf("Took %.1fs | Fetch: %.1fs | Processing: %.1fs | URL: %s", tl.Seconds(), ft.Seconds(), pt.Seconds(), origin_url)
}
