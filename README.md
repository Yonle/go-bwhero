# go-bwhero
Rewritten backend of [bandwidth-hero-proxy](https://github.com/Yonle/bandwidth-hero-proxy)

## Installation
**Requirements:**
- Have [Go](https://go.dev) installed
- Have [libvips](https://github.com/libvips/libvips) 8.18.0 installed
- Have [ffmpeg](https://ffmpeg.org) installed *[optional. only for animation]*

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
env LISTEN=localhost:8080 go-bwhero
```

If you want to increase libvips ConcurrencyLevel, You could change it by setting `CONCURRENCY_LEVEL` environment variable.

## Setting up Limit

### File Size

The bigger the file size, The longer it takes to process and the more CPU cycles are being used. By default, go-bwhero did not have any limit on file size.

You can set one (or two) of the following in environment variable to adjust:
- `IMAGESIZELIMIT`
- `ANIMATIONSIZELIMIT` (by default, it follows IMAGESIZELIMIT. If it's "-1", Then it will completely disable animation)

```
# Max image size from upstream: 50 MB
# Max animation size from upstream: 50 MB
env IMAGESIZELIMIT=50000000 ANIMATIONSIZELIMIT=50000000 go-bwhero
```

You will need to have ffmpeg installed in your system in order for the animation conversion to work. You can still run go-bwhero without it by disabling animation via `ANIMATIONSIZELIMIT=-1`.

---

### Golang Memory Usage
Due to Golang's garbage cleaner nature, You might want to adjust `GOMEMLIMIT`, `GOGC`, and `GOMAXPROCS` environment variable, where:
- `GOMEMLIMIT`: Heap Limit (example: `512MiB`)
- `GOGC`: Level of GC (example: `50` or `off`). The less, the more aggresive
- `GOMAXPROCS`: Maximum parallel allocations. (example: `2`)

Example:

```
env GOMEMLIMIT=512MiB GOMAXPROCS=2 GOGC=30 ./bwhero
```

## Testing

By default, go-bwhero will redirect to the original image URL when it was found that user is opening the image in new tab. To make testing possible, You must add `nr` query tag in your request.

Your request should look like this:
```
http://localhost:8080/?url=....&nr=1
```
