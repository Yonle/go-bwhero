package main

import (
	"context"
	"io"
	"os"
	"os/exec"
	"strconv"
)

// apng is rather, a bit specific case here.
// to put it simply, the reason we can't basically use pipe for apng is,
// somehow, the way of how the metadata / header was being put in apng is really somewhat fragmented,
// that seeking is basically needed here. we can just invoke -seekable 0, but that would mean more memory.
func process_apng(
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
		"-c:v", "libwebp_anim",
		"-loop", "0",
		"-q:v", strconv.Itoa(quality),
		"-compression_level", "2",
		"-f", "webp", // Output format
		"pipe:1", // Output to stdout
	)

	cmd.Stdout = w
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
