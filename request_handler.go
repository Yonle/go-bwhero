package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var imagesizelimit int64
var animationsizelimit int64

func init() {
	imagesizelimit, _ = strconv.ParseInt(os.Getenv("IMAGESIZELIMIT"), 10, 64)
	animationsizelimit, _ = strconv.ParseInt(os.Getenv("ANIMATIONSIZELIMIT"), 10, 64)

	if _, ok := os.LookupEnv("ANIMATIONSIZELIMIT"); !ok {
		animationsizelimit = imagesizelimit
	}
}

func request_handler(w http.ResponseWriter, r *http.Request) {
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
		http.Redirect(w, r, origin_url, http.StatusFound)
		return
	}

	kind := resp.Header.Get("Content-Type")
	isImage := strings.HasPrefix(kind, "image/")
	isAnimated :=
		strings.Contains(kind, "gif") ||
			strings.Contains(kind, "webp") ||
			strings.Contains(kind, "avif")

	var isBig bool

	switch true {
	case isImage:
		if imagesizelimit > 0 {
			isBig = resp.ContentLength > imagesizelimit
		}
	case isAnimated:
		if animationsizelimit == -1 {
			isBig = resp.ContentLength > imagesizelimit
			isAnimated = false // process a frame instead.
			break
		}
		if animationsizelimit > 0 {
			isBig = resp.ContentLength > animationsizelimit
		}
	}

	if resp.StatusCode >= 400 || !isImage || isBig {
		resp.Body.Close()
		http.Redirect(w, r, origin_url, http.StatusFound)
		return
	}

	ft := time.Since(fetch_time)

	processing_time := time.Now()

	if err := process_image(w, resp, isAnimated, quality, grayscale); err != nil {
		log.Printf("Failed to process %s: %s", origin_url, err)
		http.Redirect(w, r, origin_url, http.StatusFound)
		return
	}

	pt := time.Since(processing_time)
	tl := time.Since(fetch_time)

	log.Printf("Took %.1fs | Fetch: %.1fs | Processing: %.1fs | URL: %s", tl.Seconds(), ft.Seconds(), pt.Seconds(), origin_url)
}
