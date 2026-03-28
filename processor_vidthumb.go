package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func process_vidthumb(
	ctx context.Context,
	w io.Writer,
	url,
	headers string,
	thumbWidth,
	quality,
	grayscale int,
) error {
	if !wait(ctx) {
		return context.Canceled
	}

	defer func() { <-semaphore_anim }()

	filters := ffmpegFilterBuilder(grayscale, thumbWidth)

	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-loglevel", "warning",
		"-user_agent", ua,
		"-rw_timeout", strconv.Itoa(int(hc.Timeout)),
		"-headers", headers,
		"-an",
		"-i", url,
		"-vf", filters,
		"-pix_fmt", "yuv420p",
		"-frames:v", "1",
		"-c:v", "libwebp",
		"-q:v", strconv.Itoa(quality),
		"-compression_level", "2",
		"-f", "webp", // Output format
		"pipe:1", // Output to stdout
	)

	cmd.Stdout = w
	cmd.Stderr = os.Stderr

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
