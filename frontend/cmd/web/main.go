package main

import (
	"context"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	slog.Info("Building Gio app for WebAssembly...")

	// Get current working directory (should be frontend root)
	cwd, err := os.Getwd()
	if err != nil {
		slog.Error("getting current working directory", "error", err)
		os.Exit(1)
	}

	// Paths relative to frontend root
	webMainDir := filepath.Join(cwd, "cmd", "gio-webmain")
	webOutputDir := filepath.Join(cwd, "gio-web")
	wasmOutput := filepath.Join(webOutputDir, "app.wasm")

	// Verify we're in the right place
	if _, err := os.Stat(webMainDir); os.IsNotExist(err) {
		slog.Error("cmd/gio-webmain directory not found; run from the frontend root", "cwd", cwd)
		os.Exit(1)
	}

	// Create output directory
	if err := os.MkdirAll(webOutputDir, 0755); err != nil {
		slog.Error("creating output directory", "error", err)
		os.Exit(1)
	}

	// Build the WASM binary
	cmd := exec.CommandContext(context.Background(), "go", "build",
		"-o", wasmOutput,
		"./cmd/gio-webmain")

	cmd.Env = append(os.Environ(),
		"GOOS=js",
		"GOARCH=wasm",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	slog.Info("Running go build", "cmd", "GOOS=js GOARCH=wasm go build -o "+wasmOutput+" ./cmd/gio-webmain")

	if err := cmd.Run(); err != nil {
		slog.Error("building WASM", "error", err)
		os.Exit(1)
	}

	// Copy wasm_exec.js from Go installation
	goRoot := os.Getenv("GOROOT")
	if goRoot == "" {
		// Try to get GOROOT from go env
		cmd := exec.CommandContext(context.Background(), "go", "env", "GOROOT")
		output, err := cmd.Output()
		if err != nil {
			slog.Error("getting GOROOT", "error", err)
			os.Exit(1)
		}
		goRoot = string(output[:len(output)-1]) // Remove trailing newline
	}

	// Copy wasm_exec.js from Go installation (Go 1.23+)
	wasmExecSrc := filepath.Join(goRoot, "lib", "wasm", "wasm_exec.js")
	wasmExecDst := filepath.Join(webOutputDir, "wasm_exec.js")

	input, err := os.ReadFile(wasmExecSrc)
	if err != nil {
		slog.Error("reading wasm_exec.js", "src", wasmExecSrc, "error", err)
		os.Exit(1)
	}

	if err := os.WriteFile(wasmExecDst, input, 0644); err != nil {
		slog.Error("writing wasm_exec.js", "error", err)
		os.Exit(1)
	}

	// Update vendor assets (redoc) if missing
	vendorDir := filepath.Join(webOutputDir, "vendor")
	redocPath := filepath.Join(vendorDir, "redoc.standalone.js")
	if _, err := os.Stat(redocPath); os.IsNotExist(err) {
		slog.Info("Vendor assets missing, running update-vendor.sh...")
		vendorCmd := exec.CommandContext(context.Background(), "bash", filepath.Join(webOutputDir, "update-vendor.sh"))
		vendorCmd.Stdout = os.Stdout
		vendorCmd.Stderr = os.Stderr
		if err := vendorCmd.Run(); err != nil {
			slog.Warn("failed to update vendor assets", "error", err)
		}
	} else {
		slog.Info("Vendor assets up to date")
	}

	slog.Info("WASM build completed successfully", "output", wasmOutput)

	if stat, err := os.Stat(wasmOutput); err == nil {
		slog.Info("WASM size", "mb", float64(stat.Size())/(1024*1024))
	}

	slog.Info("Next steps: 1) create index.html in gio-web; 2) serve with `go run cmd/serve/main.go`")
}
