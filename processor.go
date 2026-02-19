package main

import (
	"context"
	"io"

	"github.com/cshum/vipsgen/vips"
)

var plainLoadOptions = &vips.LoadOptions{
	Access: vips.AccessSequential,
}
var cameraLoadOptions = &vips.LoadOptions{
	Access:     vips.AccessSequential,
	Autorotate: true,
}

func process_image(
	ctx context.Context,
	w io.WriteCloser,
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

	target := vips.NewTarget(w)
	defer target.Close()

	return img.WebpsaveTarget(target, &webpOpt)
}
