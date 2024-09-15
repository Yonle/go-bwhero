package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/h2non/bimg"
)

var hc = http.Client{
	Timeout: 10 * time.Second,
}

var blank = []byte{}

func redirect(w http.ResponseWriter, url string) {
	w.Header().Set("Location", url)
	w.WriteHeader(302)
	w.Write(blank)
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query()

	origin_url := query.Get("url")
	if len(origin_url) < 7 {
		fmt.Fprintf(w, "bandwidth-hero-proxy")
		return
	}

	quality, err := strconv.Atoi(query.Get("l"))
	if err != nil || quality > 100 || quality < 1 {
		quality = 40
	}

	grayscale, err := strconv.Atoi(query.Get("bw"))
	if err != nil {
		grayscale = 1
	}

	resp, err := proxy(ctx, r, origin_url)
	if err != nil {
		redirect(w, origin_url)
		return
	}

	log.Printf("Processing: %s", origin_url)

	if err := process_image(w, resp.Body, quality, grayscale); err != nil {
		redirect(w, origin_url)
		return
	}
}

func proxy(ctx context.Context, r *http.Request, origin_url string) (resp *http.Response, err error) {
	clientHeader := r.Header.Clone()
	clientHeader.Set("Range", r.Header.Get("Range"))
	clientHeader.Set("User-Agent", "go-bwhero")
	clientHeader.Set("Via", "2.0 go-bwhero")

	req, err := http.NewRequestWithContext(ctx, r.Method, origin_url, nil)
	if err != nil {
		return nil, err
	}

	req.Header = clientHeader
	for _, cookie := range r.Cookies() {
		req.AddCookie(cookie)
	}

	return hc.Do(req)
}

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

func main() {
	log.Println("bwhero, rewritten backend.")

	http.HandleFunc("/", handleRequest)
	listen, ok := os.LookupEnv("LISTEN")
	if !ok {
		listen = "localhost:8080"
	}

	log.Printf("Listening on %s", listen)
	if err := http.ListenAndServe(listen, nil); err != nil {
		panic(err)
	}
}
