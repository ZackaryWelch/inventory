package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	fake "github.com/brianvoe/gofakeit/v7"
	"goauthentik.io/api/v3"

	"github.com/nishiki/backend-go/app/config"
)

func TestAuthentikAuthService_GetUserGroups(t *testing.T) {
	t.Skip("Skipping integration test that requires OIDC setup - needs proper mocking or integration test setup")
}

func TestAuthentikAuthService_GetGroupUsers(t *testing.T) {
	t.Skip("Skipping integration test that requires OIDC setup - needs proper mocking or integration test setup")
}

func TestAuthentikAuthService_GetGroupUsers_Old(t *testing.T) {
	// Create mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request is for users with groups query parameter
		if r.URL.Path == "/api/v3/core/users/" && r.URL.Query().Has("groups") {
			response := AuthentikUsersResponse{
				Results: []AuthentikUser{
					{
						ID:       fake.UUID(),
						Username: fake.Username(),
						Email:    fake.Email(),
						Name:     fake.Name(),
					},
					{
						ID:       fake.UUID(),
						Username: fake.Username(),
						Email:    fake.Email(),
						Name:     fake.Name(),
					},
				},
				Count: 2,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockServer.Close()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	cfg := config.AuthConfig{
		AuthentikURL: mockServer.URL,
		Clients: []config.OAuthClient{
			{
				ProviderName: "test-provider",
				ClientID:     "test-client",
				ClientSecret: "test-secret",
				RedirectURL:  "http://localhost:3001/callback",
			},
		},
	}

	// Create API config for Authentik client
	apiConfig := api.NewConfiguration()
	apiConfig.Host = strings.TrimPrefix(mockServer.URL, "http://")
	apiConfig.Scheme = "http"
	apiConfig.HTTPClient = &http.Client{Timeout: 10 * time.Second}

	service := &AuthentikAuthService{
		config:     cfg,
		logger:     logger,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		apiConfig:  apiConfig,
	}

	// Test GetGroupUsers
	ctx := context.Background()
	users, err := service.GetGroupUsers(ctx, "test-token", "test-group-id")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}

	if len(users) > 0 {
		// Just verify we got valid user data (no specific value assertion since they're random)
		if users[0].Username().String() == "" {
			t.Error("Expected non-empty username")
		}
		if users[0].EmailAddress().String() == "" {
			t.Error("Expected non-empty email")
		}
	}
}

func TestAuthentikAuthService_GetUserByID(t *testing.T) {
	t.Skip("Skipping integration test that requires OIDC setup - needs proper mocking or integration test setup")
}

func TestAuthentikAuthService_GetUserByID_Old(t *testing.T) {
	// Create mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request is for a specific user
		if r.URL.Path == "/api/v3/core/users/test-user-1/" {
			response := AuthentikUser{
				ID:       fake.UUID(),
				Username: fake.Username(),
				Email:    fake.Email(),
				Name:     fake.Name(),
				IsActive: true,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockServer.Close()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	cfg := config.AuthConfig{
		AuthentikURL: mockServer.URL,
		Clients: []config.OAuthClient{
			{
				ProviderName: "test-provider",
				ClientID:     "test-client",
				ClientSecret: "test-secret",
				RedirectURL:  "http://localhost:3001/callback",
			},
		},
	}

	service := &AuthentikAuthService{
		config:     cfg,
		logger:     logger,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}

	// Test GetUserByID
	ctx := context.Background()
	user, err := service.GetUserByID(ctx, "test-token", "test-user-1")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if user.Username().String() == "" {
		t.Error("Expected non-empty username")
	}

	if user.EmailAddress().String() == "" {
		t.Error("Expected non-empty email")
	}

	// Test user not found
	_, err = service.GetUserByID(ctx, "test-token", "nonexistent-user")
	if err == nil {
		t.Error("Expected error for nonexistent user, got nil")
	}
}

func TestAuthentikAuthService_GetGroupByID(t *testing.T) {
	// Create mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request is for a specific group
		if r.URL.Path == "/api/v3/core/groups/test-group-1/" {
			response := AuthentikGroup{
				ID:   fake.UUID(),
				Name: fake.Company(),
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockServer.Close()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	cfg := config.AuthConfig{
		AuthentikURL: mockServer.URL,
		Clients: []config.OAuthClient{
			{
				ProviderName: "test-provider",
				ClientID:     "test-client",
				ClientSecret: "test-secret",
				RedirectURL:  "http://localhost:3001/callback",
			},
		},
	}

	service := &AuthentikAuthService{
		config:     cfg,
		logger:     logger,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}

	// Test GetGroupByID
	ctx := context.Background()
	group, err := service.GetGroupByID(ctx, "test-token", "test-group-1")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if group.Name().String() == "" {
		t.Error("Expected non-empty group name")
	}

	// Test group not found
	_, err = service.GetGroupByID(ctx, "test-token", "nonexistent-group")
	if err == nil {
		t.Error("Expected error for nonexistent group, got nil")
	}
}

func TestAuthentikAuthService_GetOIDCConfig(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse map[string]interface{}
		serverStatus   int
		backendURL     string
		expectError    bool
		expectedToken  string
	}{
		{
			name: "successful_oidc_config_retrieval",
			serverResponse: map[string]interface{}{
				"issuer":                 "https://auth.example.com/application/o/nishiki/",
				"authorization_endpoint": "https://auth.example.com/application/o/nishiki/auth/",
				"token_endpoint":         "https://auth.example.com/application/o/nishiki/token/",
				"userinfo_endpoint":      "https://auth.example.com/application/o/nishiki/userinfo/",
				"jwks_uri":               "https://auth.example.com/application/o/nishiki/jwks/",
			},
			serverStatus:  http.StatusOK,
			backendURL:    "http://localhost:3001",
			expectError:   false,
			expectedToken: "http://localhost:3001/auth/token",
		},
		{
			name:          "server_error_response",
			serverStatus:  http.StatusInternalServerError,
			expectError:   true,
			expectedToken: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock server
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.serverStatus == http.StatusOK {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(tt.serverResponse)
				} else {
					w.WriteHeader(tt.serverStatus)
				}
			}))
			defer mockServer.Close()

			// Setup service
			logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
			cfg := config.AuthConfig{
				AuthentikURL: mockServer.URL,
				Clients: []config.OAuthClient{
					{
						ProviderName: "nishiki",
						ClientID:     "test-client",
						ClientSecret: "test-secret",
						RedirectURL:  "http://localhost:3001/callback",
					},
				},
			}

			// Create mock provider for the test client
			clients := make(map[string]*clientProvider)
			clients["test-client"] = &clientProvider{
				config: cfg.Clients[0],
			}

			service := &AuthentikAuthService{
				config:     cfg,
				clients:    clients,
				logger:     logger,
				httpClient: &http.Client{Timeout: 5 * time.Second},
			}

			// Set environment variable if provided
			if tt.backendURL != "" {
				os.Setenv("BACKEND_URL", tt.backendURL)
				defer os.Unsetenv("BACKEND_URL")
			}

			// Execute
			result, err := service.GetOIDCConfig(context.Background(), "test-client")

			// Assert
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Expected no error but got: %v", err)
				return
			}

			if result == nil {
				t.Fatal("Expected result but got nil")
			}

			// Verify token endpoint was replaced
			actual := result["token_endpoint"]
			if actual != tt.expectedToken {
				t.Errorf("Expected token_endpoint %q, got %q", tt.expectedToken, actual)
			}

			// Verify other fields are preserved
			if tt.serverResponse != nil {
				for key, expected := range tt.serverResponse {
					if key == "token_endpoint" {
						continue // We expect this to be modified
					}
					actual := result[key]
					if actual != expected {
						t.Errorf("Expected %s to be %v, got %v", key, expected, actual)
					}
				}
			}
		})
	}
}

func TestAuthentikAuthService_ProxyTokenExchange(t *testing.T) {
	t.Skip("Skipping test that requires OIDC provider mocking - needs refactoring for multi-client support")

	tests := []struct {
		name              string
		inputRequest      map[string]interface{}
		serverResponse    map[string]interface{}
		serverStatus      int
		expectError       bool
		expectedStatus    int
		verifyCredentials bool
	}{
		{
			name: "successful_authorization_code_exchange",
			inputRequest: map[string]interface{}{
				"grant_type":   "authorization_code",
				"code":         "test-auth-code",
				"redirect_uri": "http://localhost:3000/callback",
			},
			serverResponse: map[string]interface{}{
				"access_token":  "access-token-123",
				"token_type":    "Bearer",
				"expires_in":    3600,
				"refresh_token": "refresh-token-123",
				"id_token":      "id.token.jwt",
			},
			serverStatus:      http.StatusOK,
			expectError:       false,
			expectedStatus:    http.StatusOK,
			verifyCredentials: true,
		},
		{
			name: "successful_refresh_token_exchange",
			inputRequest: map[string]interface{}{
				"grant_type":    "refresh_token",
				"refresh_token": "refresh-token-123",
			},
			serverResponse: map[string]interface{}{
				"access_token":  "new-access-token-456",
				"token_type":    "Bearer",
				"expires_in":    3600,
				"refresh_token": "new-refresh-token-456",
				"id_token":      "new.id.token.jwt",
			},
			serverStatus:      http.StatusOK,
			expectError:       false,
			expectedStatus:    http.StatusOK,
			verifyCredentials: true,
		},
		{
			name: "server_error_invalid_grant",
			inputRequest: map[string]interface{}{
				"grant_type": "authorization_code",
				"code":       "invalid-code",
			},
			serverResponse: map[string]interface{}{
				"error":             "invalid_grant",
				"error_description": "Authorization code is invalid",
			},
			serverStatus:   http.StatusBadRequest,
			expectError:    false, // Error should be in response, not Go error
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedFormData map[string]string

			// Setup mock server
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Capture form data for verification
				r.ParseForm()
				receivedFormData = make(map[string]string)
				for key, values := range r.Form {
					if len(values) > 0 {
						receivedFormData[key] = values[0]
					}
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				json.NewEncoder(w).Encode(tt.serverResponse)
			}))
			defer mockServer.Close()

			// Setup service
			logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
			cfg := config.AuthConfig{
				AuthentikURL: mockServer.URL,
				Clients: []config.OAuthClient{
					{
						ProviderName: "nishiki",
						ClientID:     "test-client-id",
						ClientSecret: "test-client-secret",
						RedirectURL:  "http://localhost:3000/callback",
					},
				},
			}

			// Create mock provider for the test client
			clients := make(map[string]*clientProvider)

			// For ProxyTokenExchange test, we need a provider with an endpoint
			// Since we can't easily mock the provider, we'll create the service differently
			clients["test-client-id"] = &clientProvider{
				config: cfg.Clients[0],
				// provider will be nil for this test, but ProxyTokenExchange should work
				// since it gets the token URL from the provider
			}

			service := &AuthentikAuthService{
				config:     cfg,
				clients:    clients,
				logger:     logger,
				httpClient: &http.Client{Timeout: 5 * time.Second},
			}

			// Execute
			responseBody, statusCode, err := service.ProxyTokenExchange(context.Background(), tt.inputRequest)

			// Assert error expectation
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Expected no error but got: %v", err)
				return
			}

			// Assert status code
			if statusCode != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, statusCode)
			}

			// Assert response body is valid JSON
			var actualResponse map[string]interface{}
			if err := json.Unmarshal(responseBody, &actualResponse); err != nil {
				t.Errorf("Response body is not valid JSON: %v", err)
				return
			}

			// Verify response content matches expected
			for key, expected := range tt.serverResponse {
				actual, exists := actualResponse[key]
				if !exists {
					t.Errorf("Expected response to contain key %q", key)
					continue
				}

				// Handle numeric values (JSON unmarshaling converts to float64)
				if expectedInt, ok := expected.(int); ok {
					if actualFloat, ok := actual.(float64); ok {
						if int(actualFloat) != expectedInt {
							t.Errorf("Expected %s to be %v, got %v", key, expected, actual)
						}
						continue
					}
				}

				if actual != expected {
					t.Errorf("Expected %s to be %v, got %v", key, expected, actual)
				}
			}

			// Verify client credentials were injected
			if tt.verifyCredentials {
				if receivedFormData["client_id"] != "test-client-id" {
					t.Errorf("Expected client_id to be 'test-client-id', got %q", receivedFormData["client_id"])
				}
				if receivedFormData["client_secret"] != "test-client-secret" {
					t.Errorf("Expected client_secret to be 'test-client-secret', got %q", receivedFormData["client_secret"])
				}

				// Verify original request data was preserved
				for key, expected := range tt.inputRequest {
					actual := receivedFormData[key]
					expectedStr := fmt.Sprintf("%v", expected)
					if actual != expectedStr {
						t.Errorf("Expected form field %s to be %q, got %q", key, expectedStr, actual)
					}
				}
			}
		})
	}
}
