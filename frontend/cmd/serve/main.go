package main

import (
	"fmt"
	"os"
	"path/filepath"

	"cogentcore.org/core/cmd/config"
	"cogentcore.org/core/cmd/web"
	nishikiConfig "github.com/nishiki/frontend/config"
)

func main() {
	fmt.Println("Starting cogentcore web server...")

	// Load frontend configuration to get port
	frontendConfig := nishikiConfig.LoadConfig()

	// Get current working directory (should be frontend root)
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current working directory: %v\n", err)
		os.Exit(1)
	}

	// Path relative to frontend root
	webOutputDir := filepath.Join(cwd, "web")

	// Verify we're in the right place and web directory exists
	if _, err := os.Stat(webOutputDir); os.IsNotExist(err) {
		fmt.Printf("Error: web directory not found. Please run this command from the frontend root directory and build first.\n")
		fmt.Printf("Current directory: %s\n", cwd)
		fmt.Printf("Expected web directory: %s\n", webOutputDir)
		os.Exit(1)
	}

	// Use port from frontend config, fallback to 8080 if not set
	port := frontendConfig.Port
	if port == "" {
		port = "8080"
	}

	cfg := &config.Config{
		Build: config.Build{
			Output: webOutputDir,
		},
		Web: config.Web{
			Port: port,
			Gzip: false,
		},
	}

	fmt.Printf("Serving files from: %s\n", webOutputDir)
	fmt.Printf("Server will be available at: http://localhost:%s\n", cfg.Web.Port)

	if err := web.Serve(cfg); err != nil {
		fmt.Printf("Error serving web application: %v\n", err)
		os.Exit(1)
	}
}
