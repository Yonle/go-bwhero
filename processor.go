package main

import (
	"context"
	"io"
	"net/http"

	"github.com/cshum/vipsgen/vips"
)

var plainLoadOptions = &vips.LoadOptions{
	Access: vips.AccessSequential,
}
var cameraLoadOptions = &vips.LoadOptions{
	Access:     vips.AccessSequential,
	Autorotate: true,
}

type responseWriteCloser struct {
	http.ResponseWriter
}

func (rwc responseWriteCloser) Close() error {
	return nil
}

func process_image(
	ctx context.Context,
	w http.ResponseWriter,
	body io.ReadCloser,
	potentiallyCamera bool,
	quality,
	grayscale int,
) error {
	source := vips.NewSource(body)
	defer source.Close()

	var img *vips.Image
	var err error

	if potentiallyCamera {
		img, err = vips.NewImageFromSource(source, cameraLoadOptions)
	} else {
		img, err = vips.NewImageFromSource(source, plainLoadOptions)
	}

	if err != nil {
		return err
	}

	defer img.Close()

	if ce := ctx.Err(); ce != nil {
		return ce
	}

	if grayscale == 1 {
		if err := img.Colourspace(vips.InterpretationBW, nil); err != nil {
			return err
		}
	}

	webpOpt := vips.WebpsaveTargetOptions{
		Effort:         2,
		NearLossless:   false,
		Lossless:       false,
		Mixed:          false,
		SmartSubsample: false,
		MinSize:        false,
		Q:              quality,
		Keep:           vips.KeepNone,
	}

	h := w.Header()
	h.Set("Access-Control-Allow-Origin", "*")
	h.Set("Cross-Origin-Resource-Policy", "cross-origin")
	h.Set("Cross-Origin-Embedder-Policy", "unsafe-none")
	h.Set("Cache-Control", "public, max-age=604800, stale-while-revalidate=86400") // cache for a week while asking for revalidation after a day
	h.Set("Content-Encoding", "identity")
	h.Set("Content-Type", "image/webp")
	h.Set("Transfer-Encoding", "chunked")

	w.WriteHeader(200)

	writ := responseWriteCloser{w}
	target := vips.NewTarget(writ)
	defer target.Close()

	return img.WebpsaveTarget(target, &webpOpt)
}
