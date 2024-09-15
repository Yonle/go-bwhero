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
