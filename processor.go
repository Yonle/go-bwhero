package main

import (
	"net/http"
	"strconv"

	"github.com/cshum/vipsgen/vips"
)

var loadOptions *vips.LoadOptions

func init() {
	loadOptions = &vips.LoadOptions{}
}

type responseWriteCloser struct {
	http.ResponseWriter
}

func (rwc responseWriteCloser) Close() error {
	return nil
}

func process_image(w http.ResponseWriter, resp *http.Response, quality, grayscale int) error {
	source := vips.NewSource(resp.Body)
	defer source.Close()
	defer resp.Body.Close()

	img, err := vips.NewImageFromSource(source, loadOptions)

	if err != nil {
		return err
	}

	defer img.Close()

	// rotate properly
	if img.Orientation() > 1 {
		if err := img.Autorot(nil); err != nil {
			return err
		}
	}

	img.RemoveExif()

	if grayscale == 1 {
		if err := img.Colourspace(vips.InterpretationBW, nil); err != nil {
			return err
		}
	}

	webpOpt := vips.WebpsaveTargetOptions{
		Effort:         4,
		NearLossless:   false,
		Lossless:       false,
		Mixed:          true,
		SmartSubsample: true,
		MinSize:        true,
		Q:              quality,
	}

	if img.Pages() > 1 { // animated
		webpOpt.Kmin = 9
		webpOpt.Kmax = 17
	}

	h := w.Header()
	h.Set("Access-Control-Allow-Origin", "*")
	h.Set("Cross-Origin-Resource-Policy", "cross-origin")
	h.Set("Cross-Origin-Embedder-Policy", "unsafe-none")
	h.Set("Cache-Control", "public, max-age=604800, stale-while-revalidate=86400") // cache for a week while asking for revalidation after a day
	h.Set("Content-Encoding", "identity")
	h.Set("Content-Type", "image/webp")
	h.Set("X-Original-Size", strconv.FormatInt(resp.ContentLength, 10))

	// since we're going io to io, we can't count.
	//h.Set("Content-Length", strconv.FormatInt(procsize, 10))
	//h.Set("X-Bytes-Saved", strconv.FormatInt(imgsize-procsize, 10))

	w.WriteHeader(200)

	writ := responseWriteCloser{w}
	target := vips.NewTarget(writ)
	defer target.Close()

	img.WebpsaveTarget(target, &webpOpt)

	return nil
}
