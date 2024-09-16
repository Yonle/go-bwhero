package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/davidbyttow/govips/v2/vips"
)

var vips_config = &vips.Config{
	MaxCacheFiles: 0,
	MaxCacheMem:   0,
	MaxCacheSize:  0,
}

func main() {
	log.Println("bwhero, rewritten backend.")

	listen, ok := os.LookupEnv("LISTEN")
	if !ok {
		listen = "localhost:8080"
	}

	concurrency_level, ok := os.LookupEnv("CONCURRENCY_LEVEL")
	if ok {
		cl, err := strconv.Atoi(concurrency_level)
		if err != nil {
			panic(err)
		}

		vips_config.ConcurrencyLevel = cl
	}

	start_vips()
	serve_http(listen)
}

func start_vips() {
	vips.Startup(vips_config)
}

func serve_http(listen string) {
	http.HandleFunc("/", request_handler)

	log.Printf("Listening on %s", listen)

	if err := http.ListenAndServe(listen, nil); err != nil {
		panic(err)
	}
}
