#!/bin/bash

# Build script for WebAssembly deployment
# This compiles the Cogent Core app to run in web browsers

set -e

echo "Building Nishiki Frontend for WebAssembly..."

# Set environment for WASM build
export GOOS=js
export GOARCH=wasm

# Build the WASM binary
echo "Compiling Go to WebAssembly..."
go build -o web/nishiki-frontend.wasm main.go

# Copy the WASM support files
echo "Copying WASM support files..."
GOROOT=$(go env GOROOT)
cp "$GOROOT/misc/wasm/wasm_exec.js" web/

echo "WebAssembly build complete!"
echo "Files generated:"
echo "  - web/nishiki-frontend.wasm (main application)"
echo "  - web/wasm_exec.js (Go WASM runtime)"
echo ""
echo "To serve the application:"
echo "  cd web && python3 -m http.server 8080"
echo "  or use any web server to serve the web/ directory"