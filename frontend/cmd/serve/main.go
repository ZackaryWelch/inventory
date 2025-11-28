package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	nishikiConfig "github.com/nishiki/frontend/config"
)

// spaHandler implements SPA fallback routing
type spaHandler struct {
	staticDir string
	indexPath string
}

func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get the requested path
	urlPath := r.URL.Path
	path := filepath.Join(h.staticDir, urlPath)

	// Check if the request is for a static asset (has file extension)
	ext := filepath.Ext(urlPath)
	isStaticAsset := ext != ""

	// Check if file exists
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		// If it's a static asset, try to find it in the root
		if isStaticAsset {
			// Extract just the filename
			filename := filepath.Base(urlPath)
			rootPath := filepath.Join(h.staticDir, filename)

			// Check if file exists in root
			if _, err := os.Stat(rootPath); err == nil {
				// File exists in root, serve it
				contentType := getContentType(rootPath)
				if contentType != "" {
					w.Header().Set("Content-Type", contentType)
				}
				http.ServeFile(w, r, rootPath)
				return
			}

			// Static asset not found anywhere, return 404
			http.NotFound(w, r)
			return
		}

		// Not a static asset, serve index.html for SPA routing
		http.ServeFile(w, r, h.indexPath)
		return
	} else if err != nil {
		// Other error, return internal server error
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// If it's a directory, check for index.html
	info, err := os.Stat(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if info.IsDir() {
		indexFile := filepath.Join(path, "index.html")
		_, err := os.Stat(indexFile)
		if os.IsNotExist(err) {
			// No index.html in directory, serve root index.html
			http.ServeFile(w, r, h.indexPath)
			return
		}
	}

	// File exists, serve it with proper content type
	contentType := getContentType(path)
	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	http.ServeFile(w, r, path)
}

func getContentType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".wasm":
		return "application/wasm"
	case ".js":
		return "application/javascript"
	case ".css":
		return "text/css"
	case ".html":
		return "text/html"
	case ".json":
		return "application/json"
	case ".png":
		return "image/png"
	case ".svg":
		return "image/svg+xml"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	default:
		return ""
	}
}

func main() {
	fmt.Println("Starting Nishiki Gio web server...")

	// Load frontend configuration to get port
	frontendConfig := nishikiConfig.LoadConfig()

	// Get current working directory (should be frontend root)
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current working directory: %v\n", err)
		os.Exit(1)
	}

	// Path relative to frontend root - using gio-web directory
	webOutputDir := filepath.Join(cwd, "gio-web")

	// Verify we're in the right place and gio-web directory exists
	if _, err := os.Stat(webOutputDir); os.IsNotExist(err) {
		fmt.Printf("Error: gio-web directory not found. Please build first with: go run cmd/gio-web/main.go\n")
		fmt.Printf("Current directory: %s\n", cwd)
		fmt.Printf("Expected gio-web directory: %s\n", webOutputDir)
		os.Exit(1)
	}

	indexPath := filepath.Join(webOutputDir, "index.html")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		fmt.Printf("Error: index.html not found at %s\n", indexPath)
		os.Exit(1)
	}

	// Use port from frontend config, fallback to 3000 if not set
	port := frontendConfig.Port
	if port == "" {
		port = "3000"
	}

	// Create SPA handler
	spa := spaHandler{
		staticDir: webOutputDir,
		indexPath: indexPath,
	}

	addr := fmt.Sprintf(":%s", port)
	fmt.Printf("Serving Gio app from: %s\n", webOutputDir)
	fmt.Printf("Server available at: http://localhost:%s\n", port)
	fmt.Println("SPA routing enabled - press Ctrl+C to stop")

	if err := http.ListenAndServe(addr, spa); err != nil {
		log.Fatalf("Error serving web application: %v\n", err)
	}
}
