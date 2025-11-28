package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	fmt.Println("Building Gio app for WebAssembly...")

	// Get current working directory (should be frontend root)
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current working directory: %v\n", err)
		os.Exit(1)
	}

	// Paths relative to frontend root
	webMainDir := filepath.Join(cwd, "cmd", "gio-webmain")
	webOutputDir := filepath.Join(cwd, "gio-web")
	wasmOutput := filepath.Join(webOutputDir, "app.wasm")

	// Verify we're in the right place
	if _, err := os.Stat(webMainDir); os.IsNotExist(err) {
		fmt.Printf("Error: cmd/gio-webmain directory not found. Please run this command from the frontend root directory.\n")
		fmt.Printf("Current directory: %s\n", cwd)
		os.Exit(1)
	}

	// Create output directory
	if err := os.MkdirAll(webOutputDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Build the WASM binary
	cmd := exec.Command("go", "build",
		"-o", wasmOutput,
		"./cmd/gio-webmain")

	cmd.Env = append(os.Environ(),
		"GOOS=js",
		"GOARCH=wasm",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("Running: GOOS=js GOARCH=wasm go build -o", wasmOutput, "./cmd/gio-webmain")

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error building WASM: %v\n", err)
		os.Exit(1)
	}

	// Copy wasm_exec.js from Go installation
	goRoot := os.Getenv("GOROOT")
	if goRoot == "" {
		// Try to get GOROOT from go env
		cmd := exec.Command("go", "env", "GOROOT")
		output, err := cmd.Output()
		if err != nil {
			fmt.Printf("Error getting GOROOT: %v\n", err)
			os.Exit(1)
		}
		goRoot = string(output[:len(output)-1]) // Remove trailing newline
	}

	// Copy wasm_exec.js from Go installation (Go 1.23+)
	wasmExecSrc := filepath.Join(goRoot, "lib", "wasm", "wasm_exec.js")
	wasmExecDst := filepath.Join(webOutputDir, "wasm_exec.js")

	input, err := os.ReadFile(wasmExecSrc)
	if err != nil {
		fmt.Printf("Error reading wasm_exec.js from %s: %v\n", wasmExecSrc, err)
		os.Exit(1)
	}

	if err := os.WriteFile(wasmExecDst, input, 0644); err != nil {
		fmt.Printf("Error writing wasm_exec.js: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ WASM build completed successfully!")
	fmt.Printf("✓ Output: %s\n", wasmOutput)
	fmt.Printf("✓ WASM size: ")

	if stat, err := os.Stat(wasmOutput); err == nil {
		fmt.Printf("%.2f MB\n", float64(stat.Size())/(1024*1024))
	}

	fmt.Println("\nNext steps:")
	fmt.Println("1. Create index.html in the gio-web directory")
	fmt.Println("2. Serve the application with: go run cmd/serve/main.go")
}
