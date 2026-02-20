package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nishiki/backend-go/app/config"
	"github.com/nishiki/backend-go/app/container"
	"github.com/nishiki/backend-go/app/http/routes"
	mcpserver "github.com/nishiki/backend-go/app/mcp"
)

func main() {
	mcpMode := flag.Bool("mcp", false, "Run in MCP server mode (reads NISHIKI_TOKEN env var for auth)")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize dependency container
	appContainer, err := container.NewContainer(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize container: %v", err)
	}
	defer appContainer.Close()

	if *mcpMode {
		// Redirect logger to stderr so stdout stays clean for MCP JSON-RPC.
		stderrHandler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn})
		appContainer.SetLogger(slog.New(stderrHandler))

		// Resolve user from NISHIKI_TOKEN env var.
		token := os.Getenv("NISHIKI_TOKEN")
		if token == "" {
			log.Fatal("NISHIKI_TOKEN environment variable is required in MCP mode")
		}

		ctx := context.Background()
		claims, err := appContainer.AuthService.ValidateToken(ctx, token)
		if err != nil {
			log.Fatalf("Failed to validate NISHIKI_TOKEN: %v", err)
		}

		user, err := appContainer.AuthService.GetUserFromClaims(ctx, claims)
		if err != nil {
			user, err = appContainer.AuthService.CreateUserFromClaims(ctx, claims)
			if err != nil {
				log.Fatalf("Failed to resolve user from token claims: %v", err)
			}
		}

		mctx := &mcpserver.MCPContext{
			Container: appContainer,
			User:      user,
			Token:     token,
			Notifier:  mcpserver.NewMCPNotifier(),
		}

		mcpserver.RunMCPServer(mctx)
		return
	}

	logger := appContainer.GetLogger()

	// Setup routes
	handler := routes.Setup(appContainer)

	// Determine if HTTPS should be used
	useHTTPS := cfg.Server.TLS.Enabled && fileExists(cfg.Server.TLS.CertFile) && fileExists(cfg.Server.TLS.KeyFile)

	// Create server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: handler,
	}

	if useHTTPS {
		// Load TLS configuration
		cert, err := tls.LoadX509KeyPair(cfg.Server.TLS.CertFile, cfg.Server.TLS.KeyFile)
		if err != nil {
			logger.Error("Failed to load TLS certificates", slog.Any("error", err))
			os.Exit(1)
		}

		server.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}

		logger.Info("Starting HTTPS server", slog.Int("port", cfg.Server.Port))
	} else {
		logger.Info("Starting HTTP server", slog.Int("port", cfg.Server.Port))
	}

	// Start server in goroutine
	go func() {
		var err error
		if useHTTPS {
			err = server.ListenAndServeTLS("", "")
		} else {
			err = server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	logger.Info("Server started successfully")

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", slog.Any("error", err))
	} else {
		logger.Info("Server shutdown complete")
	}
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}
