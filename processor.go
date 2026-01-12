package main

import (
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/davidbyttow/govips/v2/vips"
)

var importParams *vips.ImportParams

var bufPool = sync.Pool{
	New: func() any {
		b := make([]byte, 0, 256*1024)
		return &b
	},
}

func init() {
	importParams = vips.NewImportParams()
	animationDisabled := len(os.Getenv("NO_ANIMATE")) > 0

	if !animationDisabled {
		importParams.NumPages.Set(-1)
	}
}

func readAll(r io.ReadCloser, buf *[]byte) error {
	defer r.Close()

	b := (*buf)[:0]

	for {
		// grow capacity if needed
		if len(b) == cap(b) {
			newCap := cap(b) * 2
			if newCap == 0 {
				newCap = 32 * 1024
			}
			nb := make([]byte, len(b), newCap)
			copy(nb, b)
			b = nb
		}

		n, err := r.Read(b[len(b):cap(b)])
		b = b[:len(b)+n]

		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	*buf = b
	return nil
}

func process_image(w http.ResponseWriter, resp *http.Response, grayscale int) error {
	bufPtr := bufPool.Get().(*[]byte)
	defer bufPool.Put(bufPtr)

	var b []byte

	*bufPtr = (*bufPtr)[:0]
	b = *bufPtr

	if err := readAll(resp.Body, &b); err != nil {
		return err
	}

	img, err := vips.LoadImageFromBuffer(b, importParams)

	if err != nil {
		return err
	}

	defer img.Close()

	if grayscale == 1 {
		if err := img.ToColorSpace(vips.InterpretationBW); err != nil {
			return err
		}
	}

	params := vips.NewWebpExportParams()
	params.MinSize = true

	if img.Pages() > 1 { // animated
		params.MinKeyFrames = 9
		params.MaxKeyFrames = 17
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
	h.Set("Cache-Control", "public, max-age=604800, stale-while-revalidate=86400") // cache for a week while asking for revalidation after a day
	h.Set("Content-Encoding", "identity")
	h.Set("Content-Type", "image/webp")
	h.Set("Content-Length", strconv.FormatInt(procsize, 10))
	h.Set("X-Original-Size", strconv.FormatInt(imgsize, 10))
	h.Set("X-Bytes-Saved", strconv.FormatInt(imgsize-procsize, 10))

	w.WriteHeader(200)
	w.Write(webp)

	return nil
}
