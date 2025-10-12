package main

import (
	"fmt"
	"os"
	"path/filepath"

	"cogentcore.org/core/cmd/config"
	"cogentcore.org/core/cmd/web"
)

func main() {
	fmt.Println("Building for web with cogentcore...")

	// Get current working directory (should be frontend root)
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current working directory: %v\n", err)
		os.Exit(1)
	}

	// Paths relative to frontend root
	webMainDir := filepath.Join(cwd, "cmd", "webmain")
	webOutputDir := filepath.Join(cwd, "web")

	// Verify we're in the right place
	if _, err := os.Stat(webMainDir); os.IsNotExist(err) {
		fmt.Printf("Error: cmd/webmain directory not found. Please run this command from the frontend root directory.\n")
		fmt.Printf("Current directory: %s\n", cwd)
		os.Exit(1)
	}

	// Change to the webmain directory before building
	if err := os.Chdir(webMainDir); err != nil {
		fmt.Printf("Error changing to webmain directory: %v\n", err)
		os.Exit(1)
	}

	cfg := &config.Config{
		Build: config.Build{
			Output: webOutputDir,
		},
		Web: config.Web{
			Gzip: false,
		},
	}

	if err := web.Build(cfg); err != nil {
		fmt.Printf("Error building for web: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Web build completed successfully!")
	fmt.Println("Built files are in the 'web' directory.")
	fmt.Println("To serve the application, run: go run cmd/serve/main.go")
}
