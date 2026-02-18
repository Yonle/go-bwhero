package main

import "net/http"

// client -> upstream
var clientHeaders = []string{
	// https://www.w3.org/TR/fetch-metadata/
	"Sec-Fetch-Dest",
	"Sec-Fetch-Mode",
	"Sec-Fetch-Site",
	"Sec-Fetch-User",
}

func copyClientHeaders(dst http.Header, src http.Header) {
	for _, k := range clientHeaders {
		if vv, ok := src[k]; ok {
			for _, v := range vv {
				dst.Add(k, v)
			}
		}
	}
}
