package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nishiki/backend/app/config"
	"github.com/nishiki/backend/domain/entities"
)

type GoogleImageSearchService struct {
	apiKey         string
	searchEngineID string
	cacheDir       string
	client         *http.Client
	logger         *slog.Logger
}

func NewGoogleImageSearchService(cfg config.ImagesConfig, logger *slog.Logger) (*GoogleImageSearchService, error) {
	if cfg.GoogleAPIKey == "" || cfg.GoogleSearchEngineID == "" {
		return nil, fmt.Errorf("google_api_key and google_search_engine_id are required when images are enabled")
	}

	// Ensure cache directory exists
	if err := os.MkdirAll(cfg.CacheDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create image cache directory %s: %w", cfg.CacheDir, err)
	}

	return &GoogleImageSearchService{
		apiKey:         cfg.GoogleAPIKey,
		searchEngineID: cfg.GoogleSearchEngineID,
		cacheDir:       cfg.CacheDir,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}, nil
}

type googleSearchResponse struct {
	Items []struct {
		Link string `json:"link"`
	} `json:"items"`
}

func (s *GoogleImageSearchService) SearchAndCache(ctx context.Context, name string, objectType entities.ObjectType, properties map[string]entities.TypedValue) (string, error) {
	query := buildSearchQuery(name, objectType, properties)

	imageURL, err := s.searchImage(ctx, query)
	if err != nil {
		return "", fmt.Errorf("image search failed: %w", err)
	}
	if imageURL == "" {
		return "", nil
	}

	filename, err := s.downloadToCache(ctx, imageURL)
	if err != nil {
		s.logger.Warn("Failed to cache image, storing source URL",
			slog.String("object", name),
			slog.String("url", imageURL),
			slog.Any("error", err))
		return "", nil
	}

	servingURL := "/images/" + filename
	return servingURL, nil
}

func (s *GoogleImageSearchService) searchImage(ctx context.Context, query string) (string, error) {
	u := fmt.Sprintf(
		"https://www.googleapis.com/customsearch/v1?key=%s&cx=%s&searchType=image&num=1&q=%s",
		url.QueryEscape(s.apiKey),
		url.QueryEscape(s.searchEngineID),
		url.QueryEscape(query),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return "", fmt.Errorf("google search returned status %d: %s", resp.StatusCode, string(body))
	}

	var result googleSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode search response: %w", err)
	}

	if len(result.Items) == 0 {
		return "", nil
	}

	return result.Items[0].Link, nil
}

func (s *GoogleImageSearchService) downloadToCache(ctx context.Context, imageURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("image download returned status %d", resp.StatusCode)
	}

	// Determine file extension from Content-Type
	ext := extensionFromContentType(resp.Header.Get("Content-Type"))
	if ext == "" {
		// Try from URL
		ext = extensionFromURL(imageURL)
	}
	if ext == "" {
		ext = ".jpg"
	}

	// Generate filename from URL hash
	hash := sha256.Sum256([]byte(imageURL))
	filename := hex.EncodeToString(hash[:16]) + ext

	destPath := filepath.Join(s.cacheDir, filename)

	// Skip if already cached
	if _, err := os.Stat(destPath); err == nil {
		return filename, nil
	}

	// Download to temp file then rename (atomic)
	tmpFile, err := os.CreateTemp(s.cacheDir, "img-*")
	if err != nil {
		return "", err
	}
	tmpPath := tmpFile.Name()

	// Limit download to 10MB
	_, err = io.Copy(tmpFile, io.LimitReader(resp.Body, 10*1024*1024))
	tmpFile.Close()
	if err != nil {
		os.Remove(tmpPath)
		return "", err
	}

	if err := os.Rename(tmpPath, destPath); err != nil {
		os.Remove(tmpPath)
		return "", err
	}

	return filename, nil
}

func buildSearchQuery(name string, objectType entities.ObjectType, properties map[string]entities.TypedValue) string {
	parts := []string{name}

	// Add type-specific context to improve search results
	switch objectType {
	case entities.ObjectTypeBook:
		if v, ok := properties["author"]; ok {
			parts = append(parts, fmt.Sprintf("%v", v.Val))
		}
		parts = append(parts, "book cover")
	case entities.ObjectTypeVideoGame:
		if v, ok := properties["platform"]; ok {
			parts = append(parts, fmt.Sprintf("%v", v.Val))
		}
		parts = append(parts, "video game cover")
	case entities.ObjectTypeMusic:
		if v, ok := properties["artist"]; ok {
			parts = append(parts, fmt.Sprintf("%v", v.Val))
		}
		parts = append(parts, "album cover")
	case entities.ObjectTypeBoardGame:
		parts = append(parts, "board game box")
	case entities.ObjectTypeFood:
		if v, ok := properties["brand"]; ok {
			parts = append(parts, fmt.Sprintf("%v", v.Val))
		}
		parts = append(parts, "product")
	default:
		parts = append(parts, "product")
	}

	return strings.Join(parts, " ")
}

func extensionFromContentType(ct string) string {
	switch {
	case strings.Contains(ct, "image/jpeg"):
		return ".jpg"
	case strings.Contains(ct, "image/png"):
		return ".png"
	case strings.Contains(ct, "image/gif"):
		return ".gif"
	case strings.Contains(ct, "image/webp"):
		return ".webp"
	default:
		return ""
	}
}

func extensionFromURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	ext := filepath.Ext(u.Path)
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		return ext
	default:
		return ""
	}
}
