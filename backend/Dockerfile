FROM golang:1.26-trixie AS builder

WORKDIR /app

COPY backend/ ./src

WORKDIR /app/src

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOEXPERIMENT=jsonv2 go build \
    #-ldflags='-w -s' \
    -o nishiki .

FROM debian:trixie-slim AS server

RUN apt update && apt install -y ca-certificates curl

RUN groupadd -g 1001 nishiki && \
    useradd -u 1001 -g nishiki -m nishiki

WORKDIR /app

COPY --from=builder /app/src/nishiki /app/src/app.toml .

RUN mkdir -p certs && chown -R nishiki:nishiki /app

USER nishiki

EXPOSE 3001 3002 3003

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:3001/health || exit 1

CMD ["./nishiki"]
