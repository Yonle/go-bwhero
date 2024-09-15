package main

import (
	"net/http"
)

func redirect(w http.ResponseWriter, url string) {
	w.Header().Set("Location", url)
	w.WriteHeader(302)
	w.Write([]byte{})
}
