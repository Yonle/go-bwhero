package main

import (
	"context"
	"net/http"
	"time"
)

var hc = http.Client{
	Timeout: 10 * time.Second,
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
