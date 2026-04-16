package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nishiki/backend/app/config"
	"github.com/nishiki/backend/app/container"
	"github.com/nishiki/backend/app/http/routes"
	mcpserver "github.com/nishiki/backend/app/mcp"
	"github.com/nishiki/backend/domain/entities"
	"github.com/nishiki/backend/domain/services"
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

	// --- REST server ---
	restHandler := routes.Setup(appContainer)
	useHTTPS := cfg.Server.TLS.Enabled && fileExists(cfg.Server.TLS.CertFile) && fileExists(cfg.Server.TLS.KeyFile)

	restServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: restHandler,
	}

	if useHTTPS {
		cert, err := tls.LoadX509KeyPair(cfg.Server.TLS.CertFile, cfg.Server.TLS.KeyFile)
		if err != nil {
			logger.Error("Failed to load TLS certificates", slog.Any("error", err))
			os.Exit(1)
		}
		restServer.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
	}

	// --- MCP server ---
	mctx := &mcpserver.MCPContext{
		Container: appContainer,
		Notifier:  mcpserver.NewMCPNotifier(),
	}
	mcpSrv := mcpserver.NewMCPServer(mctx)

	// Auth factory: validates Bearer token and injects user into context.
	factory := func(r *http.Request) *mcp.Server {
		token := bearerToken(r)
		if token == "" {
			return nil
		}
		claims, err := appContainer.AuthService.ValidateToken(r.Context(), token)
		if err != nil {
			return nil
		}
		user, err := resolveOrCreateUser(r.Context(), appContainer, claims)
		if err != nil {
			return nil
		}
		*r = *r.WithContext(mcpserver.WithMCPUser(r.Context(), user, token))
		return mcpSrv
	}

	mcpHTTPServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.MCPPort),
		Handler: mcp.NewStreamableHTTPHandler(factory, &mcp.StreamableHTTPOptions{Stateless: true}),
	}
	mcpSSEServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.MCPSSEPort),
		Handler: mcp.NewSSEHandler(factory, nil),
	}

	// Start DB connection monitor
	mctx.Notifier.StartConnectionMonitor(context.Background(), mctx)

	// --- Start all servers ---
	go func() {
		var err error
		if useHTTPS {
			logger.Info("Starting HTTPS server", slog.Int("port", cfg.Server.Port))
			err = restServer.ListenAndServeTLS("", "")
		} else {
			logger.Info("Starting HTTP server", slog.Int("port", cfg.Server.Port))
			err = restServer.ListenAndServe()
		}
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("REST server failed", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	go func() {
		logger.Info("Starting MCP Streamable HTTP server", slog.Int("port", cfg.Server.MCPPort))
		if err := mcpHTTPServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("MCP HTTP server failed", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	go func() {
		logger.Info("Starting MCP SSE server", slog.Int("port", cfg.Server.MCPSSEPort))
		if err := mcpSSEServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("MCP SSE server failed", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	logger.Info("All servers started successfully")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down servers...")
	mctx.Notifier.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shut down all three servers
	var shutdownErr error
	if err := restServer.Shutdown(ctx); err != nil {
		shutdownErr = err
	}
	if err := mcpHTTPServer.Shutdown(ctx); err != nil {
		shutdownErr = err
	}
	if err := mcpSSEServer.Shutdown(ctx); err != nil {
		shutdownErr = err
	}

	if shutdownErr != nil {
		logger.Error("Server forced to shutdown", slog.Any("error", shutdownErr))
	} else {
		logger.Info("All servers shutdown complete")
	}
}

// bearerToken extracts the token from the Authorization header or ?token= query param.
func bearerToken(r *http.Request) string {
	if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	if t := r.URL.Query().Get("token"); t != "" {
		return t
	}
	return ""
}

// resolveOrCreateUser looks up or creates a user from JWT claims.
func resolveOrCreateUser(ctx context.Context, c *container.Container, claims *services.AuthClaims) (*entities.User, error) {
	user, err := c.AuthService.GetUserFromClaims(ctx, claims)
	if err != nil {
		user, err = c.AuthService.CreateUserFromClaims(ctx, claims)
		if err != nil {
			return nil, fmt.Errorf("resolve user from claims: %w", err)
		}
	}
	return user, nil
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}
