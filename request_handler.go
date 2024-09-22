package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func request_handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query()

	origin_url := query.Get("url")
	if len(origin_url) < 7 {
		fmt.Fprintf(w, "bandwidth-hero-proxy")
		return
	}

	quality, err := strconv.Atoi(query.Get("l"))
	if err != nil || quality > 100 || quality < 1 {
		quality = 40
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

	log.Printf("Processing: %s", origin_url)

	if err := process_image(w, resp, quality, grayscale); err != nil {
		log.Printf("Failed to process %s: %s", origin_url, err)
		http.Redirect(w, r, origin_url, http.StatusFound)
		return
	}
}
