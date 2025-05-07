# go-bwhero
Rewritten backend of [bandwidth-hero-proxy](https://github.com/Yonle/bandwidth-hero-proxy)

## Installation
**Requirements:**
- Have [Go](https://go.dev) installed
- Have [libvips](https://github.com/libvips/libvips) installed

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
