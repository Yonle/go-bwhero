package main

import (
	"context"
	"io"

	"github.com/cshum/vipsgen/vips"
)

var thumbPlainLoadOptions = &vips.ThumbnailSourceOptions{}

func process_thumb(
	ctx context.Context,
	w io.WriteCloser,
	body io.ReadCloser,
	width,
	quality,
	grayscale int,
) error {
	source := vips.NewSource(body)
	defer source.Close()

	var img *vips.Image
	var err error

	img, err = vips.NewThumbnailSource(source, width, thumbPlainLoadOptions)

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

	target := vips.NewTarget(w)
	defer target.Close()

	return img.WebpsaveTarget(target, &webpOpt)
}
