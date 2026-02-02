# -- build

FROM docker.io/alpine:20260127 AS builder

RUN apk add --no-cache go vips-dev vips-magick imagemagick pkgconf

WORKDIR /src/
COPY . .

RUN go build -v -trimpath -ldflags="-s -w -buildid=" -buildvcs=false -o /out/exec.bin .

# -- after build

FROM docker.io/alpine:20260127

LABEL org.opencontainers.image.source="https://github.com/Yonle/go-bwhero" \
      org.opencontainers.image.description="an image to run go-bwhero backend, an image compressor backend to be used with bandwidth hero addon." \
      org.opencontainers.image.licenses="BSD-3-Clause"

RUN apk add --no-cache vips vips-magick imagemagick
COPY --from=builder /out/exec.bin /bin/go-bwhero

ENV LISTEN=0.0.0.0:8080

CMD ["/bin/go-bwhero"]