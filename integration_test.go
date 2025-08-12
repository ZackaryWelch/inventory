package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"
)

const (
	// Test configuration
	baseURL = "http://localhost:3001"
	
	// Test JWT token (update this when it expires)
	testJWTToken = "eyJhbGciOiJSUzI1NiIsImtpZCI6ImU0M2U4ODE0N2MxMTIyNzg1MjEwYTkyOWIwZDk3OWFmIiwidHlwIjoiSldUIn0.eyJpc3MiOiJodHRwczovLzE5Mi4xNjguMC4xMjU6MzAxNDEvYXBwbGljYXRpb24vby9uaXNoaWtpLyIsInN1YiI6IjdiY2E2MWJkZGFiZDZmMjg3ZTVjMTgzMzQ3ZTAzMThjY2UxMmI5ZWU2MzUyNTg2OWNiMDcyZjAxZTAyY2QzODUiLCJhdWQiOiJWVnlwaDdNbmJHRHF0aVBxOHZmZ2w1MUVDSU8yR2NnWjEyc2tBNFZSIiwiZXhwIjoxNzU0MjcwNDkyLCJpYXQiOjE3NTQyNjk1OTIsImF1dGhfdGltZSI6MTc1NDI2NjAwOSwiYWNyIjoiZ29hdXRoZW50aWsuaW8vcHJvdmlkZXJzL29hdXRoMi9kZWZhdWx0IiwiYW1yIjpbInB3ZCJdLCJzaWQiOiI2ZGYzODZkMzZiY2YzYjU5Y2VmNTNkN2FlODk5MjYyYTYyM2Q3ZTliZmE1OWYzYWFjNzdmNmU1ZjJiMWRhZjAxIiwiZW1haWwiOiJjb3Vwb24yNTE4QGdtYWlsLmNvbSIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJuYW1lIjoiWmFjayIsImdpdmVuX25hbWUiOiJaYWNrIiwicHJlZmVycmVkX3VzZXJuYW1lIjoiendlbGNoIiwibmlja25hbWUiOiJ6d2VsY2giLCJncm91cHMiOlsiYXV0aGVudGlrIEFkbWlucyIsIlRlc3QiXSwiYXpwIjoiVlZ5cGg3TW5iR0RxdGlQcTh2ZmdsNTFFQ0lPMkdjZ1oxMnNrQTRWUiIsInVpZCI6Ikg3d2xsamkwd1duRkJ4R0xEY1A2VEN2aXlnZ2hTdFZKVnpHb3pwQjEifQ.TFOBut3K40zPkRlpbd4k_ICM-R3lPUhtp8iA2ZIQTr2NBd4ce02aZ-6sKk89dnqmewOU-KUFNI2ZY90qumlMvSEMi7jcHeEpuzBGDe6rufGmWD6A6KhyLl9x2dzuG9E2Hc4tEi4cXxUwfAZPDwhgvScl40W_cX1V_PO-D-BRZ6RzL3d-UbFa1NTpdJwYJPrebcDaxyuvHi0C3lg5PNC4yCp4G259Gh2K1OHYECMhFciQBvi5m5am6cvG0ARPVRfT-keeMS6_kwVl4jwaEP9q7w61TEESFeVX-l1FgbL7FSYLc3J06vjro7767tNMYh0VtNQ3_1SKmG0RCKP9MG8onhRCYorzcmukiWl71VoPWAhNls5OiT167GNP6R0hlACUG0E7rFbn9UzTPE3IyCE2dD9zmxV1k-rjTgEpTIAOWIZ3WqsIGDaDRIBZRhSdV33TexI4eCY4ecW1QNN21Jfny0n2_MBYn4OXHNk1kuAUdr3BIbjhdoREdyGTrKFeB6igpEhwjY-ooewckMQlxCKj4kPu3fEjMfyCAeJtCKc6nqXTikAH6VJ0itKl3M_Eu4xFnhcCKD5ozmn8YIIHoXNyfS44qzq9zUk_j-l57_LR6OvuQk-mL3K9YT81C-xJ7xRo8avm7WBvI83PSbSkLwr2vLaGKFuXF6dtzwXfshI-Pt8"
)

// Test response structures
type GroupResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateGroupRequest struct {
	Name string `json:"name"`
}

type AuthMeResponse struct {
	User struct {
		ID        string    `json:"id"`
		Name      string    `json:"name"`
		Email     string    `json:"email"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	} `json:"user"`
	Claims struct {
		Subject   string   `json:"subject"`
		Email     string   `json:"email"`
		Username  string   `json:"username"`
		Groups    []string `json:"groups"`
		ExpiresAt int64    `json:"expires_at"`
		IssuedAt  int64    `json:"issued_at"`
	} `json:"claims"`
}

// Helper function to make authenticated HTTP requests
func makeAuthenticatedRequest(method, endpoint string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, baseURL+endpoint, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+testJWTToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	return client.Do(req)
}

func TestHealthCheck(t *testing.T) {
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		t.Fatalf("Failed to make health check request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	t.Logf("Health check response: %s", string(body))
}

func TestAuthMe(t *testing.T) {
	resp, err := makeAuthenticatedRequest("GET", "/auth/me", nil)
	if err != nil {
		t.Fatalf("Failed to make auth/me request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d. Response: %s", resp.StatusCode, string(body))
	}

	var authResponse AuthMeResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResponse); err != nil {
		t.Fatalf("Failed to decode auth response: %v", err)
	}

	// Validate response structure
	if authResponse.User.ID == "" {
		t.Error("User ID should not be empty")
	}
	if authResponse.User.Email == "" {
		t.Error("User email should not be empty")
	}
	if authResponse.Claims.Subject == "" {
		t.Error("Claims subject should not be empty")
	}
	if len(authResponse.Claims.Groups) == 0 {
		t.Error("User should have at least one group")
	}

	t.Logf("Authenticated user: %s (%s)", authResponse.User.Name, authResponse.User.Email)
	t.Logf("User groups: %v", authResponse.Claims.Groups)
}

func TestGetGroupsWithTestGroup(t *testing.T) {
	resp, err := makeAuthenticatedRequest("GET", "/groups", nil)
	if err != nil {
		t.Fatalf("Failed to make GET /groups request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d. Response: %s", resp.StatusCode, string(body))
	}

	var groups []GroupResponse
	if err := json.NewDecoder(resp.Body).Decode(&groups); err != nil {
		t.Fatalf("Failed to decode groups response: %v", err)
	}

	t.Logf("Found %d groups", len(groups))

	// User should have access to the "Test" group which has the nishiki role
	found := false
	var testGroup GroupResponse
	for _, group := range groups {
		if group.Name == "Test" {
			found = true
			testGroup = group
			break
		}
	}

	if !found {
		t.Error("Expected to find 'Test' group in user's groups, but it was not found")
		t.Logf("Available groups:")
		for _, group := range groups {
			t.Logf("  - %s (ID: %s)", group.Name, group.ID)
		}
	} else {
		t.Logf("Successfully found Test group: %s (ID: %s)", testGroup.Name, testGroup.ID)
		
		// Test accessing the specific group
		t.Run("AccessTestGroup", func(t *testing.T) {
			resp, err := makeAuthenticatedRequest("GET", "/groups/"+testGroup.ID, nil)
			if err != nil {
				t.Fatalf("Failed to access Test group: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				t.Fatalf("Expected status 200, got %d. Response: %s", resp.StatusCode, string(body))
			}

			var fetchedGroup GroupResponse
			if err := json.NewDecoder(resp.Body).Decode(&fetchedGroup); err != nil {
				t.Fatalf("Failed to decode Test group response: %v", err)
			}

			if fetchedGroup.ID != testGroup.ID {
				t.Errorf("Expected group ID %s, got %s", testGroup.ID, fetchedGroup.ID)
			}
			if fetchedGroup.Name != "Test" {
				t.Errorf("Expected group name 'Test', got %s", fetchedGroup.Name)
			}

			t.Logf("Successfully accessed Test group: %s", fetchedGroup.Name)
		})
	}
}

func TestCreateGroup(t *testing.T) {
	// Create a unique group name for this test
	groupName := fmt.Sprintf("Integration Test Group %d", time.Now().Unix())
	
	createReq := CreateGroupRequest{
		Name: groupName,
	}

	resp, err := makeAuthenticatedRequest("POST", "/groups", createReq)
	if err != nil {
		t.Fatalf("Failed to make POST /groups request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 201 or 200, got %d. Response: %s", resp.StatusCode, string(body))
	}

	var group GroupResponse
	if err := json.NewDecoder(resp.Body).Decode(&group); err != nil {
		t.Fatalf("Failed to decode group response: %v", err)
	}

	// Validate created group
	if group.ID == "" {
		t.Error("Created group should have an ID")
	}
	if group.Name != groupName {
		t.Errorf("Expected group name '%s', got '%s'", groupName, group.Name)
	}
	if group.CreatedAt.IsZero() {
		t.Error("Created group should have a creation timestamp")
	}

	t.Logf("Successfully created group: %s (ID: %s)", group.Name, group.ID)

	// Verify we can fetch the created group
	t.Run("FetchCreatedGroup", func(t *testing.T) {
		resp, err := makeAuthenticatedRequest("GET", "/groups/"+group.ID, nil)
		if err != nil {
			t.Fatalf("Failed to fetch created group: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected status 200, got %d. Response: %s", resp.StatusCode, string(body))
		}

		var fetchedGroup GroupResponse
		if err := json.NewDecoder(resp.Body).Decode(&fetchedGroup); err != nil {
			t.Fatalf("Failed to decode fetched group: %v", err)
		}

		if fetchedGroup.ID != group.ID {
			t.Errorf("Expected group ID '%s', got '%s'", group.ID, fetchedGroup.ID)
		}
		if fetchedGroup.Name != group.Name {
			t.Errorf("Expected group name '%s', got '%s'", group.Name, fetchedGroup.Name)
		}

		t.Logf("Successfully fetched group: %s", fetchedGroup.Name)
	})
}

func TestGetGroupsAfterCreation(t *testing.T) {
	resp, err := makeAuthenticatedRequest("GET", "/groups", nil)
	if err != nil {
		t.Fatalf("Failed to make GET /groups request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d. Response: %s", resp.StatusCode, string(body))
	}

	var groups []GroupResponse
	if err := json.NewDecoder(resp.Body).Decode(&groups); err != nil {
		t.Fatalf("Failed to decode groups response: %v", err)
	}

	t.Logf("Found %d groups after creation", len(groups))
	
	for _, group := range groups {
		t.Logf("  - Group: %s (ID: %s, Created: %s)", 
			group.Name, group.ID, group.CreatedAt.Format(time.RFC3339))
	}
}

func TestUnauthenticatedRequest(t *testing.T) {
	req, err := http.NewRequest("GET", baseURL+"/groups", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make unauthenticated request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401 for unauthenticated request, got %d", resp.StatusCode)
	}

	t.Logf("Correctly rejected unauthenticated request with status %d", resp.StatusCode)
}

func TestInvalidToken(t *testing.T) {
	req, err := http.NewRequest("GET", baseURL+"/groups", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer invalid-token")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request with invalid token: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401 for invalid token, got %d", resp.StatusCode)
	}

	t.Logf("Correctly rejected invalid token with status %d", resp.StatusCode)
}

// Integration test that runs all tests in sequence
func TestIntegrationFlow(t *testing.T) {
	t.Run("1_HealthCheck", TestHealthCheck)
	t.Run("2_AuthMe", TestAuthMe) 
	t.Run("3_GetGroupsWithTestGroup", TestGetGroupsWithTestGroup)
	t.Run("4_CreateGroup", TestCreateGroup)
	t.Run("5_GetGroupsAfterCreation", TestGetGroupsAfterCreation)
	t.Run("6_UnauthenticatedRequest", TestUnauthenticatedRequest)
	t.Run("7_InvalidToken", TestInvalidToken)
}