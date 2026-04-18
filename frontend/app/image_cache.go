package app

import (
	"fmt"
	"image"
	// Register JPEG and PNG decoders for image.Decode.
	_ "image/jpeg"
	_ "image/png"
	"sync"
)

// imageEntry holds a cached decoded image and its loading state.
type imageEntry struct {
	img     image.Image
	loading bool
	err     error
}

// imageCache provides thread-safe in-memory caching for decoded images.
type imageCache struct {
	mu      sync.Mutex
	entries map[string]*imageEntry
}

func newImageCache() *imageCache {
	return &imageCache{
		entries: make(map[string]*imageEntry),
	}
}

// getOrLoad returns the cached image for the given URL. If it hasn't been fetched
// yet, it marks it as loading and returns nils — the caller should call loadImage
// in a goroutine. Returns (img, alreadyLoading, err).
func (ic *imageCache) getOrLoad(url string) (image.Image, bool, error) {
	ic.mu.Lock()
	defer ic.mu.Unlock()

	if e, ok := ic.entries[url]; ok {
		return e.img, e.loading, e.err
	}

	// Mark as loading so subsequent calls don't start duplicate fetches.
	ic.entries[url] = &imageEntry{loading: true}
	return nil, false, nil
}

// store saves a decoded image (or error) into the cache.
func (ic *imageCache) store(url string, img image.Image, err error) {
	ic.mu.Lock()
	defer ic.mu.Unlock()
	ic.entries[url] = &imageEntry{img: img, err: err}
}

// loadImage fetches and decodes an image from the backend via the API client.
// Call this from a goroutine; it stores the result in the cache and invalidates
// the window so the next frame picks it up.
func (ga *GioApp) loadImage(url string) {
	resp, err := ga.apiClient.Get(url)
	if err != nil {
		ga.imgCache.store(url, nil, fmt.Errorf("fetch: %w", err))
		ga.window.Invalidate()
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		ga.imgCache.store(url, nil, fmt.Errorf("HTTP %d", resp.StatusCode))
		ga.window.Invalidate()
		return
	}

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		ga.imgCache.store(url, nil, fmt.Errorf("decode: %w", err))
		ga.window.Invalidate()
		return
	}

	ga.imgCache.store(url, img, nil)
	ga.window.Invalidate()
}

// getImage returns a decoded image for the URL if cached and ready, or nil.
// Kicks off an async fetch on first access.
func (ga *GioApp) getImage(url string) image.Image {
	img, _ := ga.getImageStatus(url)
	return img
}

// getImageStatus returns the cached image (or nil) and the fetch/decode error
// (or nil) for the URL. Kicks off an async fetch on first access. Callers can
// use the error to render a broken-image placeholder instead of the default
// loading one.
func (ga *GioApp) getImageStatus(url string) (image.Image, error) {
	if url == "" {
		return nil, nil
	}
	img, alreadyLoading, err := ga.imgCache.getOrLoad(url)
	if img != nil || err != nil {
		return img, err
	}
	if !alreadyLoading {
		go ga.loadImage(url)
	}
	return nil, nil
}
