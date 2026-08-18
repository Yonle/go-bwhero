package main

import (
	"flag"
	"log"
	"net/http"
	"runtime"

	"github.com/cshum/vipsgen/vips"
)

var vips_config = &vips.Config{
	MaxCacheFiles: 0,
	MaxCacheMem:   0,
	MaxCacheSize:  0,
	ReportLeaks:   true,
}

func main() {
	log.Println("bwhero, rewritten backend.")

	listen := flag.String(
		"listen",
		"localhost:8080",
		"Listen address",
	)

	flag.IntVar(
		&vips_config.ConcurrencyLevel,
		"vipsConcurrencyLevel",
		0,
		"libvips concurrency level to use",
	)

	flag.Int64Var(
		&imagesizelimit,
		"imgSizeLimit",
		-1,
		"Original image size limit in bytes",
	)

	flag.Int64Var(
		&animationsizelimit,
		"animSizeLimit",
		-1,
		"Original animation size limit in bytes",
	)

	flag.Int64Var(
		&videosizelimit,
		"videoSizeLimit",
		-1,
		"Original video size limit in bytes",
	)

	flag.StringVar(
		&ua,
		"userAgent",
		"Mozilla/5.0; go-bwhero [https://github.com/Yonle/bwhero]",
		"User agent that go-bwhero should use.",
	)

	workerLen := flag.Int(
		"workers",
		runtime.NumCPU(),
		"Amount of workers to spawn",
	)

	flag.Parse()

	start_vips()
	startWorker(*workerLen)
	serve_http(*listen)
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
