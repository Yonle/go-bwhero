FROM docker.io/alpine:edge

LABEL org.opencontainers.image.source="https://github.com/Yonle/go-bwhero" \
      org.opencontainers.image.description="an image to run go-bwhero backend, an image compressor backend to be used with bandwidth hero addon." \
      org.opencontainers.image.licenses="BSD-3-Clause"

WORKDIR /a
COPY . .

RUN apk add --no-cache vips vips-dev vips-magick imagemagick go \
    && go build -trimpath -ldflags="-s -w -buildid=" -buildvcs=false -o /a/b . \
    && apk del go vips-dev 

ENV LISTEN=0.0.0.0:8080

CMD ["/a/b"]
