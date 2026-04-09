package services

import (
	"context"

	"github.com/nishiki/backend/domain/entities"
)

// ImageSearchResult contains the result of an image search.
type ImageSearchResult struct {
	// SourceURL is the original URL of the image found online.
	SourceURL string
	// CachedPath is the local filesystem path where the image was cached.
	CachedPath string
	// ServingURL is the URL path the backend will serve the cached image at.
	ServingURL string
}

// ImageSearchService searches for object images and caches them locally.
type ImageSearchService interface {
	// SearchAndCache searches for an image matching the object, downloads it
	// to the local cache, and returns the serving URL.
	// Returns empty string and nil error if no image was found.
	SearchAndCache(ctx context.Context, name string, objectType entities.ObjectType, properties map[string]entities.TypedValue) (string, error)
}
