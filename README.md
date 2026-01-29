# go-bwhero
Rewritten backend of [bandwidth-hero-proxy](https://github.com/Yonle/bandwidth-hero-proxy)

## Installation
**Requirements:**
- Have [Go](https://go.dev) installed
- Have [libvips](https://github.com/libvips/libvips) 8.18.0 installed

**Install:**
```
go install -v codeberg.org/Yonle/go-bwhero@latest
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

---

**NOTICE: THE FOLLOWING OPTIONS IS TEMPORARILY UNAVAILABLE**

By default, `go-bwhero` has no semaphore level. You might need to configure it to avoid OOM (out of memory) on intensive environment:
```
lv: 10
SEM=10 go-bwhero
```

By default, `go-bwhero` render animated image. This could cause high memory usage on intensive environment. To disable it, set `NO_ANIMATE` to `1`:
```
NO_ANIMATE=1 go-bwhero
```

---

## Setting up Limit

Due to Golang's garbage cleaner nature, You might want to adjust `GOMEMLIMIT`, `GOGC`, and `GOMAXPROCS` environment variable, where:
- `GOMEMLIMIT`: Heap Limit (example: `512MiB`)
- `GOGC`: Level of GC (example: `50` or `off`). The less, the more aggresive
- `GOMAXPROCS`: Maximum parallel allocations. (example: `2`)
