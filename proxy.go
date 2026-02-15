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
	req.Header.Set("User-Agent", "go-bwhero [https://github.com/Yonle/bwhero]")
	req.Header.Set("Via", "2.0 go-bwhero")

	return hc.Do(req)
}
