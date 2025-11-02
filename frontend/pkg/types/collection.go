package types

import "time"

// Collection represents a collection of objects
type Collection struct {
	ID          string      `json:"id"`
	UserID      string      `json:"user_id"`
	GroupID     *string     `json:"group_id,omitempty"`
	Name        string      `json:"name"`
	CategoryID  *string     `json:"category_id,omitempty"`
	ObjectType  string      `json:"object_type"` // food, book, videogame, music, boardgame, general
	Containers  []Container `json:"containers,omitempty"`
	Tags        []string    `json:"tags"`
	Location    string      `json:"location"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// CreateCollectionRequest represents the request to create a new collection
type CreateCollectionRequest struct {
	GroupID    *string  `json:"group_id,omitempty"`
	Name       string   `json:"name" binding:"required"`
	ObjectType string   `json:"object_type" binding:"required"`
	Tags       []string `json:"tags,omitempty"`
	Location   string   `json:"location,omitempty"`
}

// UpdateCollectionRequest represents the request to update a collection
type UpdateCollectionRequest struct {
	Name     string   `json:"name"`
	Tags     []string `json:"tags,omitempty"`
	Location string   `json:"location,omitempty"`
}

// BulkImportRequest represents a request to bulk import objects to a collection
type BulkImportRequest struct {
	CollectionID      string                   `json:"collection_id" binding:"required"`
	TargetContainerID *string                  `json:"target_container_id,omitempty"`
	DistributionMode  string                   `json:"distribution_mode,omitempty"` // "automatic", "manual", "target"
	Format            string                   `json:"format" binding:"required"`   // "csv" or "json"
	Data              []map[string]interface{} `json:"data" binding:"required"`
	DefaultTags       []string                 `json:"default_tags,omitempty"`
}

// BulkImportResponse represents the response from a bulk import operation
type BulkImportResponse struct {
	Imported         int                       `json:"imported"`
	Failed           int                       `json:"failed"`
	Total            int                       `json:"total"`
	Errors           []string                  `json:"errors,omitempty"`
	CapacityWarnings []CapacityWarning         `json:"capacity_warnings,omitempty"`
	Assignments      map[string]int            `json:"assignments,omitempty"` // containerID -> count
}

// CapacityWarning represents a warning about container capacity
type CapacityWarning struct {
	ContainerID   string  `json:"container_id"`
	ContainerName string  `json:"container_name"`
	UsedCapacity  float64 `json:"used_capacity"`
	TotalCapacity float64 `json:"total_capacity"`
	Utilization   float64 `json:"utilization"`
	Severity      string  `json:"severity"` // "warning" or "critical"
}

// ObjectTypeInfo holds display information for object types
type ObjectTypeInfo struct {
	Name  string
	Icon  string
	Color string
}

// ObjectTypes maps object type IDs to their display information
var ObjectTypes = map[string]ObjectTypeInfo{
	"food": {
		Name:  "Food",
		Icon:  "🍎",
		Color: "#6ab3ab",
	},
	"book": {
		Name:  "Book",
		Icon:  "📚",
		Color: "#fcd884",
	},
	"videogame": {
		Name:  "Video Game",
		Icon:  "🎮",
		Color: "#cd5a5a",
	},
	"music": {
		Name:  "Music",
		Icon:  "🎵",
		Color: "#95cec6",
	},
	"boardgame": {
		Name:  "Board Game",
		Icon:  "🎲",
		Color: "#f1c560",
	},
	"general": {
		Name:  "General",
		Icon:  "📦",
		Color: "#bdbdbd",
	},
}
