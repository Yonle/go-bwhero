package main

import (
	"net/http"
	"strconv"

	"github.com/cshum/vipsgen/vips"
)

var plainLoadOptions = &vips.LoadOptions{
	Access: vips.AccessSequential,
}
var cameraLoadOptions = &vips.LoadOptions{
	Access:     vips.AccessSequential,
	Autorotate: true,
}
var animatedLoadOptions = &vips.LoadOptions{
	Access: vips.AccessSequential,
	N:      -1,
}

type responseWriteCloser struct {
	http.ResponseWriter
}

func (rwc responseWriteCloser) Close() error {
	return nil
}

func process_image(w http.ResponseWriter, resp *http.Response, isAnimated bool, potentiallyCamera bool, quality, grayscale int) error {
	source := vips.NewSource(resp.Body)
	defer source.Close()
	defer resp.Body.Close()

	var img *vips.Image
	var err error

	if isAnimated {
		img, err = vips.NewImageFromSource(source, animatedLoadOptions)
	} else if potentiallyCamera {
		img, err = vips.NewImageFromSource(source, cameraLoadOptions)
	} else {
		img, err = vips.NewImageFromSource(source, plainLoadOptions)
	}

	if err != nil {
		return err
	}

	defer img.Close()

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
		Keep:           vips.KeepIcc,
	}

	if img.Pages() > 1 { // animated
		webpOpt.Kmin = 3
		webpOpt.Kmax = 30
	}

	h := w.Header()
	h.Set("Access-Control-Allow-Origin", "*")
	h.Set("Cross-Origin-Resource-Policy", "cross-origin")
	h.Set("Cross-Origin-Embedder-Policy", "unsafe-none")
	h.Set("Cache-Control", "public, max-age=604800, stale-while-revalidate=86400") // cache for a week while asking for revalidation after a day
	h.Set("Content-Encoding", "identity")
	h.Set("Content-Type", "image/webp")
	h.Set("Transfer-Encoding", "chunked")
	h.Set("X-Original-Size", strconv.FormatInt(resp.ContentLength, 10))

	// since we're going io to io, we can't count.
	//h.Set("Content-Length", strconv.FormatInt(procsize, 10))
	//h.Set("X-Bytes-Saved", strconv.FormatInt(imgsize-procsize, 10))

	w.WriteHeader(200)

	writ := responseWriteCloser{w}
	target := vips.NewTarget(writ)
	defer target.Close()

	return img.WebpsaveTarget(target, &webpOpt)
}
