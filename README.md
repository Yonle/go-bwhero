# go-bwhero

Rewritten backend of [bandwidth-hero-proxy](https://github.com/Yonle/bandwidth-hero-proxy).

## Installation

**Requirements:**

- [Go](https://go.dev) installed
- [libvips](https://github.com/libvips/libvips) 8.18.0 installed
- [FFmpeg](https://ffmpeg.org) installed *(optional, only required for animations)*

**Install:**

```sh
go install -v github.com/Yonle/go-bwhero@latest
````

Or use the ready-to-use Docker image:

```sh
docker run -d -p 8080:8080 --name bwhero yonle/bwhero
```

Or with Docker Compose:

```sh
docker compose up --build
```

## Listening

```sh
go-bwhero \
  -listen localhost:8080
```

You can increase the libvips concurrency level with:

```sh
go-bwhero \
  -vipsConcurrencyLevel 4
```

## File Size Limits

Larger files take more time and CPU resources to process. By default, go-bwhero does not limit the file size.

You can set the following limits:

* `-imgSizeLimit` — Maximum image size in bytes
* `-animSizeLimit` — Maximum animation size in bytes
* `-videoSizeLimit` — Maximum video size in bytes

For example:

```sh
# Maximum image size: 50 MB
# Maximum animation size: 50 MB

go-bwhero \
  -imgSizeLimit 50000000 \
  -animSizeLimit 50000000
```

Set `-animSizeLimit` to `-2` to disable animation processing completely.

`-videoSizeLimit` works similarly, but only affects video thumbnailing.

FFmpeg is required for animation processing. You can run go-bwhero without FFmpeg by disabling animation processing.

## Go Memory Usage

You can adjust Go's memory and garbage collection behavior with the following environment variables:

* `GOMEMLIMIT` — Soft memory limit, e.g. `512MiB`
* `GOGC` — Garbage collection target percentage, e.g. `50` or `off`
* `GOMAXPROCS` — Maximum number of CPUs available to Go, e.g. `2`

Example:

```sh
GOMEMLIMIT=512MiB GOMAXPROCS=2 GOGC=30 ./go-bwhero
```

## User Agent

The default user agent is:

```text
Mozilla/5.0; go-bwhero [https://github.com/Yonle/bwhero]
```

You can change it with the `-userAgent` flag:

```sh
go-bwhero \
  -userAgent "My User Agent"
```

## Testing

By default, go-bwhero redirects to the original image URL when it detects that the user is opening the image in a new tab.

To prevent this behavior during testing, add the `nr` query parameter:

```text
http://localhost:8080/?url=....&nr=1
```
