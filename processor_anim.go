package main

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
)

var semaphore_anim = make(chan struct{}, runtime.NumCPU())

func wait(ctx context.Context) bool {
	select {
	case semaphore_anim <- struct{}{}:
		return true
	case <-ctx.Done():
		<-semaphore_anim
		return true
	}
}

func process_anim(ctx context.Context, w http.ResponseWriter, resp *http.Response, quality, grayscale int) error {
	if !wait(ctx) {
		return context.Canceled
	}

	defer func() { <-semaphore_anim }()

	filters := "null"

	if grayscale == 1 {
		filters = "hue=s=0"
	}

	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-loglevel", "error",
		"-i", "pipe:0", // Input from stdin
		"-vf", filters, // Video filters (Greyscale)
		"-c:v", "libwebp_anim", // WebP encoder
		"-loop", "0", // Infinite loop
		"-q:v", strconv.Itoa(quality),
		"-compression_level", "4",
		"-f", "webp", // Output format
		"pipe:1", // Output to stdout
	)

	cmd.Stdin = resp.Body
	cmd.Stdout = w
	cmd.Stderr = os.Stderr

	h := w.Header()
	h.Set("Access-Control-Allow-Origin", "*")
	h.Set("Cross-Origin-Resource-Policy", "cross-origin")
	h.Set("Cross-Origin-Embedder-Policy", "unsafe-none")
	h.Set("Cache-Control", "public, max-age=604800, stale-while-revalidate=86400") // cache for a week while asking for revalidation after a day
	h.Set("Content-Encoding", "identity")
	h.Set("Content-Type", "image/webp")
	h.Set("Transfer-Encoding", "chunked")
	h.Set("X-Original-Size", strconv.FormatInt(resp.ContentLength, 10))

	w.WriteHeader(200)

	return cmd.Run()
}
