package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var imagesizelimit int64

func init() {
	i, err := strconv.ParseInt(os.Getenv("IMAGESIZELIMIT"), 10, 64)

	if err != nil {
		return
	}

	imagesizelimit = i
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

	grayscale, err := strconv.Atoi(query.Get("bw"))
	if err != nil {
		grayscale = 1
	}

	resp, err := proxy(ctx, r, origin_url)
	if err != nil {
		http.Redirect(w, r, origin_url, http.StatusFound)
		return
	}

	var isImage bool = strings.HasPrefix(resp.Header.Get("Content-Type"), "image/")
	var isBig bool

	if imagesizelimit > 0 {
		isBig = resp.ContentLength > imagesizelimit
	}

	if resp.StatusCode >= 400 || !isImage || isBig {
		resp.Body.Close()
		http.Redirect(w, r, origin_url, http.StatusFound)
		return
	}

	log.Printf("Processing: %s", origin_url)

	if err := process_image(w, resp, grayscale); err != nil {
		log.Printf("Failed to process %s: %s", origin_url, err)
		http.Redirect(w, r, origin_url, http.StatusFound)
		return
	}
}
