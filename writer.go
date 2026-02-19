package main

import (
	"net/http"
)

type clientWriter struct {
	w             http.ResponseWriter
	headerWritten bool
}

func (cw *clientWriter) writeHeader() {
	h := cw.w.Header()
	h.Set("Access-Control-Allow-Origin", "*")
	h.Set("Cross-Origin-Resource-Policy", "cross-origin")
	h.Set("Cross-Origin-Embedder-Policy", "unsafe-none")
	h.Set("Cache-Control", "public, max-age=604800, immutable")
	h.Set("Content-Type", "image/webp")

	cw.w.WriteHeader(http.StatusOK)
}

func (cw *clientWriter) Write(p []byte) (n int, err error) {
	if !cw.headerWritten {
		cw.writeHeader()
		cw.headerWritten = true
	}

	return cw.w.Write(p)
}

func (cw *clientWriter) Close() error {
	return nil
}
