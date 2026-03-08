package mcpserver

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// MCPNotifier monitors DB health and logs state transitions.
// In HTTP mode, notifications are logged rather than sent to a specific session.
type MCPNotifier struct {
	mu          sync.Mutex
	cancel      context.CancelFunc
	lastDBState string
}

// NewMCPNotifier creates a new MCPNotifier.
func NewMCPNotifier() *MCPNotifier {
	return &MCPNotifier{}
}

// StartConnectionMonitor polls DB health every 30s and logs state transitions.
// Should be called once at startup.
func (n *MCPNotifier) StartConnectionMonitor(ctx context.Context, mctx *MCPContext) {
	ctx, cancel := context.WithCancel(ctx)
	n.mu.Lock()
	n.cancel = cancel
	n.mu.Unlock()

	if n.pingDB(ctx, mctx) {
		n.lastDBState = "connected"
	} else {
		n.lastDBState = "disconnected"
	}

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				n.pollDB(ctx, mctx)
			}
		}
	}()
}

// Stop cancels the connection monitor goroutine.
func (n *MCPNotifier) Stop() {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.cancel != nil {
		n.cancel()
	}
}

// pingDB checks DB health by listing collections. Returns true if healthy.
func (n *MCPNotifier) pingDB(ctx context.Context, mctx *MCPContext) bool {
	_, err := mctx.Container.CollectionRepo.List(ctx, 1, 0)
	return err == nil
}

// pollDB checks DB health and logs on state change.
func (n *MCPNotifier) pollDB(ctx context.Context, mctx *MCPContext) {
	newState := "disconnected"
	if n.pingDB(ctx, mctx) {
		newState = "connected"
	}

	n.mu.Lock()
	old := n.lastDBState
	n.lastDBState = newState
	n.mu.Unlock()

	if old != newState {
		if newState == "disconnected" {
			slog.Warn("MCP: MongoDB connection state changed", "old", old, "new", newState)
		} else {
			slog.Info("MCP: MongoDB connection state changed", "old", old, "new", newState)
		}
	}
}
