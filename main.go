package main

import (
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
	w.Header().Add("Location", url)
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

	quality, err1 := strconv.Atoi(query.Get("l"))
	if err1 != nil || quality > 100 || quality < 1 {
		quality = 40
	}

	grayscale, err2 := strconv.Atoi(query.Get("bw"))
	if err2 != nil {
		grayscale = 1
	}


	// cookie, dnt, referer, range, user-agent, x-forwarded-for
	clientHeader := r.Header.Clone()
	clientHeader.Add("Range", r.Header.Get("Range"))
	clientHeader.Add("User-Agent", "go-bwhero")
	clientHeader.Add("Via", "2.0 go-bwhero")

	req, err3 := http.NewRequestWithContext(ctx, r.Method, origin_url, nil)
	if err3 != nil {
		http.Error(w, "Something was wrong.", http.StatusBadGateway)
		return
	}

	req.Header = clientHeader
	for _, cookie := range r.Cookies() {
		req.AddCookie(cookie)
	}

	resp, err4 := hc.Do(req)
	if err4 != nil {
		redirect(w, origin_url)
		return
	}

	if resp.ContentLength > 200000000 {
		redirect(w, origin_url)
		resp.Body.Close()
		return
	}

	imgbytes, err5 := io.ReadAll(resp.Body)
	if err5 != nil {
		redirect(w, origin_url)
		resp.Body.Close()
		return
	}
	resp.Body.Close()

	log.Printf("Processing: %s", origin_url)

	opt := bimg.Options{
		Quality: quality,
		Type:    bimg.WEBP,
	}

	if grayscale == 1 {
		opt.Interpretation = bimg.InterpretationBW
	}

	processed, err6 := bimg.NewImage(imgbytes).Process(opt)
	if err6 != nil {
		redirect(w, origin_url)
		return
	}

	imgsize := len(imgbytes)
	procsize := len(processed)

	h := w.Header()
	h.Add("content-type", "image/webp")
	h.Add("content-length", strconv.Itoa(procsize))
	h.Add("x-original-size", strconv.Itoa(imgsize))
	h.Add("x-bytes-saved", strconv.Itoa(imgsize-procsize))

	w.WriteHeader(200)
	w.Write(processed)

	return
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
