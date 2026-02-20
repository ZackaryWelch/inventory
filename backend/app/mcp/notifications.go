package mcpserver

import (
	"context"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nishiki/backend-go/domain/usecases"
)

// MCPNotifier sends notifications to the MCP client via the server session.
type MCPNotifier struct {
	mu          sync.Mutex
	session     *mcp.ServerSession
	cancel      context.CancelFunc
	lastDBState string
}

// NewMCPNotifier creates a new MCPNotifier.
func NewMCPNotifier() *MCPNotifier {
	return &MCPNotifier{}
}

// SetSession stores the server session for sending notifications.
func (n *MCPNotifier) SetSession(ss *mcp.ServerSession) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.session = ss
}

// SendNotification sends a structured log notification to the client.
// Gracefully no-ops if no session is set.
func (n *MCPNotifier) SendNotification(ctx context.Context, level mcp.LoggingLevel, event string, data map[string]any) {
	n.mu.Lock()
	ss := n.session
	n.mu.Unlock()
	if ss == nil {
		return
	}

	payload := map[string]any{"event": event}
	for k, v := range data {
		payload[k] = v
	}

	_ = ss.Log(ctx, &mcp.LoggingMessageParams{
		Level:  level,
		Logger: "nishiki",
		Data:   payload,
	})
}

// StartConnectionMonitor polls DB health every 30s and fires
// notifications on state transitions. Stops when ctx is cancelled.
func (n *MCPNotifier) StartConnectionMonitor(ctx context.Context, mctx *MCPContext) {
	ctx, cancel := context.WithCancel(ctx)
	n.mu.Lock()
	n.cancel = cancel
	n.mu.Unlock()

	// Initialize cached state.
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

// pingDB checks DB health by fetching the user's collections. Returns true if healthy.
func (n *MCPNotifier) pingDB(ctx context.Context, mctx *MCPContext) bool {
	_, err := mctx.Container.CollectionRepo.GetByUserID(ctx, mctx.User.ID())
	return err == nil
}

// pollDB checks DB health and fires a notification on state change.
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
		level := mcp.LoggingLevel("info")
		if newState == "disconnected" {
			level = "warning"
		}
		n.SendNotification(ctx, level, "connection/changed", map[string]any{
			"service":   "mongodb",
			"old_state": old,
			"new_state": newState,
		})
	}
}

// NotifyExpiringItems scans all collections for objects expiring within 7 days
// and sends a notification if any are found. Called on-demand.
func (n *MCPNotifier) NotifyExpiringItems(ctx context.Context, mctx *MCPContext) {
	resp, err := mctx.getCollectionsUC().Execute(ctx, usecases.GetCollectionsRequest{
		UserID:    mctx.userID(),
		UserToken: mctx.Token,
	})
	if err != nil {
		return
	}

	threshold := time.Now().Add(7 * 24 * time.Hour)
	var expiring []map[string]any

	for _, collection := range resp.Collections {
		for _, obj := range collection.GetAllObjects() {
			if obj.ExpiresAt() != nil && obj.ExpiresAt().Before(threshold) {
				expiring = append(expiring, map[string]any{
					"id":            obj.ID().String(),
					"name":          obj.Name().String(),
					"expires_at":    obj.ExpiresAt(),
					"collection_id": collection.ID().String(),
				})
			}
		}
	}

	if len(expiring) > 0 {
		n.SendNotification(ctx, "warning", "items/expiring_soon", map[string]any{
			"items": expiring,
			"count": len(expiring),
		})
	}
}
