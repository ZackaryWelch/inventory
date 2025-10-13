package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/nishiki/frontend/pkg/types"
)

// TokenFetcher is an interface for fetching authentication tokens
type TokenFetcher interface {
	GetAccessToken() (string, error)
	IsTokenValid() bool
}

// Client is the shared HTTP client for making API requests
type Client struct {
	BaseURL      string
	HTTPClient   *http.Client
	TokenFetcher TokenFetcher
}

// NewClient creates a new API client
func NewClient(baseURL string, tokenFetcher TokenFetcher) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		TokenFetcher: tokenFetcher,
	}
}

// Request makes an authenticated HTTP request
func (c *Client) Request(method, endpoint string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, c.BaseURL+endpoint, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Get access token from token fetcher
	accessToken, err := c.TokenFetcher.GetAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	return c.HTTPClient.Do(req)
}

// Get makes a GET request
func (c *Client) Get(endpoint string) (*http.Response, error) {
	return c.Request(http.MethodGet, endpoint, nil)
}

// Post makes a POST request
func (c *Client) Post(endpoint string, body interface{}) (*http.Response, error) {
	return c.Request(http.MethodPost, endpoint, body)
}

// Put makes a PUT request
func (c *Client) Put(endpoint string, body interface{}) (*http.Response, error) {
	return c.Request(http.MethodPut, endpoint, body)
}

// Delete makes a DELETE request
func (c *Client) Delete(endpoint string) (*http.Response, error) {
	return c.Request(http.MethodDelete, endpoint, nil)
}

// DecodeResponse decodes a JSON response into the provided type
func DecodeResponse[T any](resp *http.Response) (*T, error) {
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp types.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("API error: status %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("API error: %s (code: %d)", errResp.Message, resp.StatusCode)
	}

	var result T
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// DecodeResponseList decodes a JSON array response
func DecodeResponseList[T any](resp *http.Response) ([]T, error) {
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp types.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("API error: status %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("API error: %s (code: %d)", errResp.Message, resp.StatusCode)
	}

	var result []T
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// CheckResponse checks the response status and returns an error if not successful
func CheckResponse(resp *http.Response) error {
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp types.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return fmt.Errorf("API error: status %d", resp.StatusCode)
		}
		return fmt.Errorf("API error: %s (code: %d)", errResp.Message, resp.StatusCode)
	}

	return nil
}

// Result represents a Result type for error handling (similar to Rust's Result)
type Result[T any] struct {
	Value T
	Err   error
}

// Ok creates a successful Result
func Ok[T any](value T) Result[T] {
	return Result[T]{Value: value, Err: nil}
}

// Err creates an error Result
func Err[T any](err error) Result[T] {
	var zero T
	return Result[T]{Value: zero, Err: err}
}

// IsOk returns true if the result is successful
func (r Result[T]) IsOk() bool {
	return r.Err == nil
}

// IsErr returns true if the result is an error
func (r Result[T]) IsErr() bool {
	return r.Err != nil
}

// Unwrap returns the value or panics if there's an error
func (r Result[T]) Unwrap() T {
	if r.Err != nil {
		panic(r.Err)
	}
	return r.Value
}

// UnwrapOr returns the value or the provided default if there's an error
func (r Result[T]) UnwrapOr(defaultValue T) T {
	if r.Err != nil {
		return defaultValue
	}
	return r.Value
}
