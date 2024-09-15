package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	log.Println("bwhero, rewritten backend.")

	http.HandleFunc("/", request_handler)
	listen, ok := os.LookupEnv("LISTEN")
	if !ok {
		listen = "localhost:8080"
	}

	log.Printf("Listening on %s", listen)
	if err := http.ListenAndServe(listen, nil); err != nil {
		panic(err)
	}
}
