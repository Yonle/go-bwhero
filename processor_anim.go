package main

import (
	"context"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strconv"
)

var semaphore_anim = make(chan struct{}, runtime.NumCPU())

func wait(
	ctx context.Context,
) bool {
	select {
	case semaphore_anim <- struct{}{}:
		return true
	case <-ctx.Done():
		<-semaphore_anim
		return true
	}
}

func process_anim(
	ctx context.Context,
	w io.Writer,
	body io.Reader,
	format string,
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

	filters += ",format=yuva420p"

	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-loglevel", "warning",
		"-f", format,
		"-i", "pipe:0", // Input from stdin
		"-vf", filters, // Video filters (Greyscale)
		"-c:v", "libwebp_anim", // WebP encoder
		"-loop", "0", // Infinite loop
		"-q:v", strconv.Itoa(quality),
		"-compression_level", "2",
		"-f", "webp", // Output format
		"pipe:1", // Output to stdout
	)

	cmd.Stdin = body
	cmd.Stdout = w
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
