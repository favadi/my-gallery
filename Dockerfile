FROM golang:1.15-alpine AS builder
WORKDIR /build
COPY . /build
RUN \
  apk update \
  && apk add --no-cache git ca-certificates
RUN \
  CGO_ENABLED=0 \
  go build \
    -ldflags "-w -s" \
    -o /my-gallery

FROM scratch
COPY --from=builder /my-gallery /app/
COPY ./assets/ /app/assets
WORKDIR /app
ENTRYPOINT ["./my-gallery"]
