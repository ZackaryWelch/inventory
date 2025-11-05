package types

import "time"

// DialogState manages dialog visibility and content
type DialogState struct {
	ShowCreateGroup      bool
	ShowEditGroup        bool
	ShowDeleteGroup      bool
	ShowInviteUser       bool
	ShowCreateCollection bool
	ShowEditCollection   bool
	ShowDeleteCollection bool
	ShowCreateContainer  bool
	ShowEditContainer    bool
	ShowDeleteContainer  bool
	ShowCreateObject     bool
	ShowEditObject       bool
	ShowDeleteObject     bool

	// Current context
	CurrentGroupID      string
	CurrentCollectionID string
	CurrentContainerID  string
	CurrentObjectID     string
}

// SearchFilter represents search and filter criteria
type SearchFilter struct {
	Query         string
	Category      string
	Tags          []string
	ObjectType    string
	ContainerID   string
	CollectionID  string
	SortBy        string // name, created_at, updated_at, expires_at
	SortDirection string // asc, desc
	ExpiryRange   *DateRange
}

// DateRange represents a date range for filtering
type DateRange struct {
	Start *time.Time
	End   *time.Time
}

// SearchResult represents a search result item
type SearchResult struct {
	ObjectID       string
	ObjectName     string
	ContainerID    string
	ContainerName  string
	CollectionID   string
	CollectionName string
	Category       string
	Snippet        string
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// SuccessResponse represents a generic success response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalItems int         `json:"total_items"`
	TotalPages int         `json:"total_pages"`
}

// ViewMode represents the view mode for displaying items
type ViewMode string

const (
	ViewModeGrid ViewMode = "grid"
	ViewModeList ViewMode = "list"
)

// View constants for navigation
const (
	ViewLogin       = "login"
	ViewCallback    = "callback"
	ViewDashboard   = "dashboard"
	ViewGroups      = "groups"
	ViewCollections = "collections"
	ViewProfile     = "profile"
	ViewSearch      = "search"
)

// SortDirection constants
const (
	SortAsc  = "asc"
	SortDesc = "desc"
)

// SortField constants
const (
	SortByName      = "name"
	SortByCreatedAt = "created_at"
	SortByUpdatedAt = "updated_at"
	SortByExpiresAt = "expires_at"
)

// Note: BulkImportRequest and BulkImportCollectionRequest are re-exported in user.go from backend types

// Container type constants
const (
	ContainerTypeRoom      = "room"
	ContainerTypeBookshelf = "bookshelf"
	ContainerTypeShelf     = "shelf"
	ContainerTypeBinder    = "binder"
	ContainerTypeCabinet   = "cabinet"
	ContainerTypeGeneral   = "general"
)
