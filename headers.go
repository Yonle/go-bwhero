package main

import "net/http"

// client -> upstream
var clientHeaders = []string{
	"If-None-Match",
	"If-Modified-Since",
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
