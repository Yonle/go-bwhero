package main

import (
	"context"
	"net/http"
	"time"
)

var hc = http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		DisableCompression: true,
	},
}

var ua string

func proxy(
	ctx context.Context,
	r *http.Request,
	origin_url string,
) (
	resp *http.Response,
	err error,
) {
	req, err := http.NewRequestWithContext(ctx, r.Method, origin_url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "image/*")
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Via", "2.0 go-bwhero")

	// if it has a referrer, set it
	if ref := r.Referer(); len(ref) > 0 {
		req.Header.Set("Referer", ref)
	}

	copyClientHeaders(req.Header, r.Header)

	return hc.Do(req)
}
