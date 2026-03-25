FROM golang:1.26-trixie AS builder

WORKDIR /app

COPY frontend/ ./

# Build WASM binary
RUN GOOS=js GOARCH=wasm go build -o gio-web/app.wasm ./cmd/gio-webmain

# Copy wasm_exec.js from Go installation
RUN cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" gio-web/

# Build the serve binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o nishiki-frontend ./cmd/serve

FROM debian:trixie-slim

RUN groupadd -g 1001 nishiki && \
    useradd -u 1001 -g nishiki -m nishiki

WORKDIR /app

COPY --from=builder /app/gio-web/ ./gio-web/
COPY --from=builder /app/nishiki-frontend .
COPY --from=builder /app/config.toml .

RUN chown -R nishiki:nishiki /app

USER nishiki

EXPOSE 3000

CMD ["./nishiki-frontend"]
