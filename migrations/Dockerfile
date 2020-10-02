FROM golang:1.15-alpine AS builder
WORKDIR /build/migrations
COPY . /build
RUN \
  apk update \
  && apk add --no-cache git ca-certificates
RUN \
  CGO_ENABLED=0 \
  go build \
    -ldflags "-w -s" \
    -o /migrate

FROM scratch
COPY --from=builder /migrate /app/
COPY ./migrations/sql /app/sql
WORKDIR /app
ENTRYPOINT ["./migrate", "up"]
