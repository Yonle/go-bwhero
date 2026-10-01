package main

import (
	"context"
	"net"
	"net/http"
	"time"
)

var timeout = 10 * time.Second

var hc = http.Client{
	Transport: &http.Transport{
		DisableCompression: true,
		DialContext: (&net.Dialer{
			Timeout: timeout,
		}).DialContext,

		TLSHandshakeTimeout:   timeout,
		ResponseHeaderTimeout: timeout,
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

	// if it has a referrer, set it
	if ref := r.Referer(); len(ref) > 0 {
		req.Header.Set("Referer", ref)
	}

	copyClientHeaders(req.Header, r.Header)

	return hc.Do(req)
}
