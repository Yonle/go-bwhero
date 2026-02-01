package main

import (
	"encoding/base64"
	"net/http"
	"os"
	"strings"
)

var username, password string
var doAuth bool
var auth_realm = `Basic realm="go-bwhero compressor service"`

func init() {
	username = os.Getenv("AUTH_USERNAME")
	password = os.Getenv("AUTH_PASSWORD")
	doAuth = len(username) != 0 && len(password) != 0
}

func basicAuth(w http.ResponseWriter, r *http.Request) bool {
	if !doAuth {
		return true
	}

	auth := r.Header.Get("Authorization")
	if auth == "" {
		w.Header().Set("WWW-Authenticate", auth_realm)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return false
	}

	const prefix = "Basic "
	if !strings.HasPrefix(auth, prefix) {
		w.Header().Set("WWW-Authenticate", auth_realm)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return false
	}

	decoded, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		w.Header().Set("WWW-Authenticate", auth_realm)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return false
	}

	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 || parts[0] != username || parts[1] != password {
		w.Header().Set("WWW-Authenticate", auth_realm)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return false
	}

	return true
}
