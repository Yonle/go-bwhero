# go-bwhero
Rewritten backend of [bandwidth-hero-proxy](https://github.com/Yonle/bandwidth-hero-proxy)

## Installation
**Requirements:**
- Have [Go](https://go.dev) installed
- Have [libvips](https://github.com/libvips/libvips) 8.18.0 installed

**Install:**
```
go install -v github.com/Yonle/go-bwhero@latest
```

or via an ready-to-use docker image:
```
docker run -d -p 8080:8080 --name bwhero yonle/bwhero
```

or via docker compose:
```
docker compose up --build
```

## Listening
```
LISTEN=localhost:8080 go-bwhero
```

If you want to increase libvips ConcurrencyLevel, You could change it by setting `CONCURRENCY_LEVEL` environment variable.

By default, `go-bwhero` did not have a limit of original bytes during fetching. You could configure it to redirect user to original image instead if an image is too big to be processed by setting `IMAGESIZELIMIT`

```
# This will start a server with a limit of 50 MB
IMAGESIZELIMIT=50000000 go-bwhero
```

## Authentication

To make this proxy to be only accessible to you, You can set `AUTH_USERNAME` and `AUTH_PASSWORD` environment variable.
```
AUTH_USERNAME=lain AUTH_PASSWORD=letsalllovelain go-bwhero
```

---

## Setting up Limit

Due to Golang's garbage cleaner nature, You might want to adjust `GOMEMLIMIT`, `GOGC`, and `GOMAXPROCS` environment variable, where:
- `GOMEMLIMIT`: Heap Limit (example: `512MiB`)
- `GOGC`: Level of GC (example: `50` or `off`). The less, the more aggresive
- `GOMAXPROCS`: Maximum parallel allocations. (example: `2`)

Example:

```
GOMEMLIMIT=512MiB GOMAXPROCS=2 GOGC=30 ./bwhero
```
