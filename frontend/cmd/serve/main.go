package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	nishikiConfig "github.com/nishiki/frontend/config"
)

// spaHandler implements SPA fallback routing
type spaHandler struct {
	staticDir  string
	indexPath  string
	backendURL string
}

func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Serve docs.html with backend URL injected
	if r.URL.Path == "/docs" || r.URL.Path == "/docs/" {
		h.serveDocs(w, r)
		return
	}

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

func (h spaHandler) serveDocs(w http.ResponseWriter, r *http.Request) {
	docsPath := filepath.Join(h.staticDir, "docs.html")
	content, err := os.ReadFile(docsPath)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	// Inject the backend URL so the page can find the OpenAPI spec
	injected := strings.Replace(string(content),
		"window.__NISHIKI_BACKEND_URL__",
		fmt.Sprintf("'%s'", h.backendURL),
		1)
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(injected)) //nolint:errcheck
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
	slog.Info("Starting Nishiki Gio web server...")

	// Load frontend configuration to get port
	frontendConfig := nishikiConfig.LoadConfig()

	// Get current working directory (should be frontend root)
	cwd, err := os.Getwd()
	if err != nil {
		slog.Error("getting current working directory", "error", err)
		os.Exit(1)
	}

	// Path relative to frontend root - using gio-web directory
	webOutputDir := filepath.Join(cwd, "gio-web")

	// Verify we're in the right place and gio-web directory exists
	if _, err := os.Stat(webOutputDir); os.IsNotExist(err) {
		slog.Error("gio-web directory not found; build first with: go run cmd/gio-web/main.go",
			"cwd", cwd, "expected", webOutputDir)
		os.Exit(1)
	}

	indexPath := filepath.Join(webOutputDir, "index.html")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		slog.Error("index.html not found", "path", indexPath)
		os.Exit(1)
	}

	// Use port from frontend config, fallback to 3000 if not set
	port := frontendConfig.Port
	if port == "" {
		port = "3000"
	}

	// Create SPA handler
	spa := spaHandler{
		staticDir:  webOutputDir,
		indexPath:  indexPath,
		backendURL: frontendConfig.BackendURL,
	}

	addr := ":" + port
	slog.Info("Serving Gio app",
		"dir", webOutputDir,
		"url", "http://localhost:"+port,
		"note", "SPA routing enabled — press Ctrl+C to stop")

	if err := http.ListenAndServe(addr, spa); err != nil {
		log.Fatalf("Error serving web application: %v\n", err)
	}
}
