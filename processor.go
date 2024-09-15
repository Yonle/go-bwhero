package main

import (
	"io"
	"net/http"
	"strconv"

	"github.com/h2non/bimg"
)

func process_image(w http.ResponseWriter, rc io.ReadCloser, quality int, grayscale int) error {
	defer rc.Close()
	imgbytes, err := io.ReadAll(rc)
	if err != nil {
		return err
	}

	opt := bimg.Options{
		Quality: quality,
		Type:    bimg.WEBP,
	}

	if grayscale == 1 {
		opt.Interpretation = bimg.InterpretationBW
	}

	processed, err := bimg.NewImage(imgbytes).Process(opt)
	if err != nil {
		return err
	}

	imgsize := len(imgbytes)
	procsize := len(processed)

	h := w.Header()
	h.Set("content-type", "image/webp")
	h.Set("content-length", strconv.Itoa(procsize))
	h.Set("x-original-size", strconv.Itoa(imgsize))
	h.Set("x-bytes-saved", strconv.Itoa(imgsize-procsize))

	w.WriteHeader(200)
	w.Write(processed)

	return nil
}
