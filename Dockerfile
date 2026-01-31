FROM docker.io/alpine:edge

LABEL org.opencontainers.image.source="https://github.com/Yonle/go-bwhero" \
      org.opencontainers.image.description="basically a container for bkerler/edl program, because you hate waiting for its dependencies to get compiled." \
      org.opencontainers.image.licenses="MIT"

WORKDIR /a
COPY . .

RUN apk add --no-cache vips vips-dev vips-magick imagemagick go \
    && go build -trimpath -ldflags="-s -w -buildid=" -buildvcs=false -o /a/b . \
    && apk del go vips-dev 

ENV LISTEN=0.0.0.0:8080

CMD ["/a/b"]
