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

// CollectionListResponse represents the wrapped list response from the backend
type CollectionListResponse struct {
	Collections []Collection `json:"collections"`
	Total       int          `json:"total"`
}

// CreateCollectionRequest represents the request to create a new collection
type CreateCollectionRequest struct {
	UserID     string   `json:"user_id" binding:"required"`
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
