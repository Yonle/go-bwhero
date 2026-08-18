# -- build

FROM docker.io/alpine:20260805 AS builder

RUN apk add --no-cache go vips-dev vips-magick imagemagick pkgconf

WORKDIR /src/
COPY . .

RUN go build -v -trimpath -ldflags="-s -w -buildid=" -buildvcs=false -o /out/exec.bin .

# -- after build

FROM docker.io/alpine:20260805

LABEL org.opencontainers.image.source="https://github.com/Yonle/go-bwhero" \
      org.opencontainers.image.description="an image to run go-bwhero backend, an image compressor backend to be used with bandwidth hero addon." \
      org.opencontainers.image.licenses="BSD-3-Clause"

RUN apk add --no-cache ffmpeg vips vips-magick imagemagick bash
COPY --from=builder /out/exec.bin /bin/go-bwhero
COPY backward-compatibility.sh /bin/backward-compatibility.sh

ENV LISTEN=0.0.0.0:8080

CMD ["/bin/backward-compatibility.sh"]
