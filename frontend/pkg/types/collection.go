package types

import "time"

// Collection represents a collection of objects
type Collection struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ObjectType  string    `json:"object_type"` // food, book, videogame, music, boardgame, general
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Containers  []Container `json:"containers,omitempty"`
}

// CreateCollectionRequest represents the request to create a new collection
type CreateCollectionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ObjectType  string `json:"object_type"`
}

// UpdateCollectionRequest represents the request to update a collection
type UpdateCollectionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
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
