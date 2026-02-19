package main

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func process_vidthumb(
	ctx context.Context,
	w http.ResponseWriter,
	url,
	headers string,
	quality,
	grayscale int,
) error {
	if !wait(ctx) {
		return context.Canceled
	}

	defer func() { <-semaphore_anim }()

	filters := "null"

	if grayscale == 1 {
		filters = "hue=s=0"
	}

	filters += ",thumbnail"

	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-loglevel", "warning",
		"-user_agent", ua,
		"-headers", headers,
		"-i", url,
		"-vf", filters,
		"-frames:v", "1",
		"-c:v", "libwebp",
		"-q:v", strconv.Itoa(quality),
		"-compression_level", "2",
		"-f", "webp", // Output format
		"pipe:1", // Output to stdout
	)

	cmd.Stdout = w
	cmd.Stderr = os.Stderr

	h := w.Header()
	h.Set("Access-Control-Allow-Origin", "*")
	h.Set("Cross-Origin-Resource-Policy", "cross-origin")
	h.Set("Cross-Origin-Embedder-Policy", "unsafe-none")
	h.Set("Cache-Control", "public, max-age=604800, immutable")
	h.Set("Content-Encoding", "identity")
	h.Set("Content-Type", "image/webp")
	h.Set("Transfer-Encoding", "chunked")

	w.WriteHeader(200)

	return cmd.Run()
}

func HeaderToFFmpegFormat(h http.Header) string {
	var b strings.Builder

	for key, values := range h {
		for _, v := range values {
			b.WriteString(key)
			b.WriteString(": ")
			b.WriteString(v)
			b.WriteString("\r\n")
		}
	}

	return b.String()
}
