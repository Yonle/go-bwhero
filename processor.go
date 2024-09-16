package main

import (
	"net/http"
	"strconv"

	"github.com/davidbyttow/govips/v2/vips"
)

func process_image(w http.ResponseWriter, resp *http.Response, quality int, grayscale int) error {
	defer resp.Body.Close()
	img, err := vips.NewImageFromReader(resp.Body)
	if err != nil {
		return err
	}

	params := vips.NewWebpExportParams()
	params.Quality = quality

	if grayscale == 1 {
		if err := img.ToColorSpace(vips.InterpretationBW); err != nil {
			return err
		}
	}

	webp, _, err := img.ExportWebp(params)

	if err != nil {
		return err
	}

	imgsize := resp.ContentLength
	procsize := int64(len(webp))

	h := w.Header()
	h.Set("Access-Control-Allow-Origin", "*")
	h.Set("Cross-Origin-Resource-Policy", "cross-origin")
	h.Set("Cross-Origin-Embedder-Policy", "unsafe-none")
	h.Set("Content-Encoding", "identity")
	h.Set("Content-Type", "image/webp")
	h.Set("Content-Length", strconv.FormatInt(procsize, 10))
	h.Set("X-Original-Size", strconv.FormatInt(imgsize, 10))
	h.Set("X-Bytes-Saved", strconv.FormatInt(imgsize-procsize, 10))

	w.WriteHeader(200)
	w.Write(webp)

	return nil
}
