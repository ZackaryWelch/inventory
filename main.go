package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nishiki/backend-go/app/config"
	"github.com/nishiki/backend-go/app/container"
	"github.com/nishiki/backend-go/app/http/routes"
)

func main() {
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

	logger := appContainer.GetLogger()

	// Setup Gin mode
	if cfg.Server.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	router := gin.New()
	router.Use(gin.Recovery())

	// Setup routes
	routes.Setup(router, appContainer)

	// Determine if HTTPS should be used
	useHTTPS := cfg.Server.TLS.Enabled && fileExists(cfg.Server.TLS.CertFile) && fileExists(cfg.Server.TLS.KeyFile)

	// Create server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: router,
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
