package services

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/cast"
	"goauthentik.io/api/v3"

	"github.com/nishiki/backend-go/app/config"
	"github.com/nishiki/backend-go/domain/entities"
	"github.com/nishiki/backend-go/domain/services"
)

// Authentik API error response structures
type AuthentikForbiddenError struct {
	Detail string `json:"detail"`
	Code   string `json:"code"`
}

type AuthentikValidationError struct {
	NonFieldErrors []string `json:"non_field_errors"`
	Code          string   `json:"code"`
}

type clientProvider struct {
	config   config.OAuthClient
	provider *oidc.Provider
	verifier *oidc.IDTokenVerifier
}

type AuthentikAuthService struct {
	config     config.AuthConfig
	clients    map[string]*clientProvider // client_id -> provider/verifier
	logger     *slog.Logger
	httpClient *http.Client
	apiConfig  *api.Configuration
}

func NewAuthentikAuthService(config config.AuthConfig, logger *slog.Logger) (*AuthentikAuthService, error) {
	ctx := context.Background()

	// Create HTTP client with optional self-signed certificate support
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	if config.AllowSelfSigned {
		httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
		logger.Warn("Self-signed certificates are enabled - this should only be used in development")
	}

	ctx = oidc.ClientContext(ctx, httpClient)

	// Initialize all OAuth clients
	clients := make(map[string]*clientProvider)

	for _, clientConfig := range config.Clients {
		// Construct the correct Authentik OIDC provider URL
		providerURL := fmt.Sprintf("%s/application/o/%s/", config.AuthentikURL, clientConfig.ProviderName)
		logger.Info("Creating OIDC provider",
			slog.String("provider_name", clientConfig.ProviderName),
			slog.String("provider_url", providerURL))

		// Create OIDC provider
		provider, err := oidc.NewProvider(ctx, providerURL)
		if err != nil {
			return nil, fmt.Errorf("failed to create OIDC provider for client %s: %w", clientConfig.ProviderName, err)
		}

		// Create ID token verifier
		verifier := provider.Verifier(&oidc.Config{
			ClientID: clientConfig.ClientID,
		})

		clients[clientConfig.ClientID] = &clientProvider{
			config:   clientConfig,
			provider: provider,
			verifier: verifier,
		}

		logger.Info("OAuth client initialized successfully",
			slog.String("provider_name", clientConfig.ProviderName),
			slog.String("client_id", clientConfig.ClientID))
	}

	// Create Authentik API configuration
	apiConfig := api.NewConfiguration()
	apiConfig.Host = strings.TrimPrefix(config.AuthentikURL, "https://")
	apiConfig.Host = strings.TrimPrefix(apiConfig.Host, "http://")
	apiConfig.Scheme = "https"
	if strings.Contains(config.AuthentikURL, "http://") {
		apiConfig.Scheme = "http"
	}
	apiConfig.HTTPClient = httpClient

	return &AuthentikAuthService{
		config:     config,
		clients:    clients,
		logger:     logger,
		httpClient: httpClient,
		apiConfig:  apiConfig,
	}, nil
}

func (s *AuthentikAuthService) ValidateToken(ctx context.Context, tokenString string) (*services.AuthClaims, error) {
	// Remove "Bearer " prefix if present
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	// Try to verify token with each client until one succeeds
	var lastErr error
	for clientID, client := range s.clients {
		idToken, err := client.verifier.Verify(ctx, tokenString)
		if err != nil {
			lastErr = err
			continue
		}

		// Extract claims
		var claims services.AuthClaims
		if err := idToken.Claims(&claims); err != nil {
			s.logger.Error("Failed to extract claims", slog.Any("error", err))
			lastErr = err
			continue
		}

		// Validate token expiration
		if time.Now().Unix() > claims.ExpiresAt {
			s.logger.Warn("Token has expired", slog.Int64("exp", claims.ExpiresAt))
			lastErr = fmt.Errorf("token has expired")
			continue
		}

		s.logger.Debug("Token validated successfully",
			slog.String("client_id", clientID),
			slog.String("subject", claims.Subject),
			slog.String("username", claims.Username),
			slog.String("email", claims.Email))

		return &claims, nil
	}

	s.logger.Error("Token verification failed for all clients", slog.Any("error", lastErr))
	return nil, fmt.Errorf("token verification failed: %w", lastErr)
}

// getClientByRedirectURL finds the appropriate OAuth client based on the redirect_uri
func (s *AuthentikAuthService) getClientByRedirectURL(redirectURI string) (*clientProvider, error) {
	if redirectURI == "" {
		return nil, fmt.Errorf("redirect_uri is required")
	}

	// Try exact match first
	for _, client := range s.clients {
		if client.config.RedirectURL == redirectURI {
			s.logger.Debug("Matched client by redirect_uri (exact)",
				slog.String("redirect_uri", redirectURI),
				slog.String("provider_name", client.config.ProviderName))
			return client, nil
		}
	}

	// Try matching by origin if exact match fails
	requestURL, err := url.Parse(redirectURI)
	if err != nil {
		return nil, fmt.Errorf("invalid redirect_uri: %w", err)
	}
	requestOrigin := fmt.Sprintf("%s://%s", requestURL.Scheme, requestURL.Host)

	for _, client := range s.clients {
		configURL, err := url.Parse(client.config.RedirectURL)
		if err != nil {
			s.logger.Warn("Invalid redirect URL for client",
				slog.String("provider_name", client.config.ProviderName),
				slog.String("redirect_url", client.config.RedirectURL))
			continue
		}

		configOrigin := fmt.Sprintf("%s://%s", configURL.Scheme, configURL.Host)

		if requestOrigin == configOrigin {
			s.logger.Debug("Matched client by origin",
				slog.String("origin", requestOrigin),
				slog.String("provider_name", client.config.ProviderName))
			return client, nil
		}
	}

	return nil, fmt.Errorf("no OAuth client configured for redirect_uri: %s", redirectURI)
}

// getClientByClientID finds the appropriate OAuth client based on client_id
func (s *AuthentikAuthService) getClientByClientID(clientID string) (*clientProvider, error) {
	if clientID == "" {
		return nil, fmt.Errorf("client_id is required")
	}

	client, ok := s.clients[clientID]
	if !ok {
		return nil, fmt.Errorf("no OAuth client configured for client_id: %s", clientID)
	}

	s.logger.Debug("Matched client by client_id",
		slog.String("client_id", clientID),
		slog.String("provider_name", client.config.ProviderName))

	return client, nil
}

func (s *AuthentikAuthService) GetUserFromClaims(ctx context.Context, claims *services.AuthClaims) (*entities.User, error) {
	// Create user entity from Authentik claims
	return s.createUserFromClaims(claims)
}

func (s *AuthentikAuthService) CreateUserFromClaims(ctx context.Context, claims *services.AuthClaims) (*entities.User, error) {
	// Create user entity from Authentik claims
	return s.createUserFromClaims(claims)
}

func (s *AuthentikAuthService) createUserFromClaims(claims *services.AuthClaims) (*entities.User, error) {
	// Create username
	username, err := entities.NewUsername(claims.Username)
	if err != nil {
		return nil, fmt.Errorf("invalid username in claims: %w", err)
	}

	// Create email address
	email, err := entities.NewEmailAddress(claims.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email in claims: %w", err)
	}

	// Create user ID from Authentik subject
	userID, err := entities.UserIDFromString(claims.Subject)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID in claims: %w", err)
	}

	// Create user entity (using ReconstructUser since we have all the data)
	user := entities.ReconstructUser(userID, username, email, claims.Subject, time.Now(), time.Now())

	return user, nil
}

// ParseTokenClaims parses JWT token without verification (for debugging)
func (s *AuthentikAuthService) ParseTokenClaims(tokenString string) (*jwt.MapClaims, error) {
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return &claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

// Authentik API response structures
type AuthentikGroup struct {
	ID          string                 `json:"pk"`
	Name        string                 `json:"name"`
	ParentName  string                 `json:"parent_name"`
	IsSuperuser bool                   `json:"is_superuser"`
	NumPK       int                    `json:"num_pk"`
	Attributes  map[string]interface{} `json:"attributes"`
}

type AuthentikUser struct {
	ID         string                 `json:"pk"`
	Username   string                 `json:"username"`
	Email      string                 `json:"email"`
	Name       string                 `json:"name"`
	IsActive   bool                   `json:"is_active"`
	Avatar     string                 `json:"avatar"`
	Attributes map[string]interface{} `json:"attributes"`
}

type AuthentikGroupsResponse struct {
	Results []AuthentikGroup `json:"results"`
	Count   int              `json:"count"`
}

type AuthentikUsersResponse struct {
	Results []AuthentikUser `json:"results"`
	Count   int             `json:"count"`
}

// GetUserGroups fetches groups the user is a member of using JWT token claims and Authentik API with API token
func (s *AuthentikAuthService) GetUserGroups(ctx context.Context, userToken, userID string) ([]*entities.Group, error) {
	s.logger.Debug("Extracting user groups from JWT token",
		slog.String("user_id", userID))

	// Parse token without validation (already validated by auth middleware)
	rawClaims, err := s.ParseTokenClaims(userToken)
	if err != nil {
		s.logger.Error("Failed to parse token claims", slog.Any("error", err))
		return nil, fmt.Errorf("failed to parse token claims: %w", err)
	}

	// Extract groups from claims
	var groupNames []string
	if groupsRaw, ok := (*rawClaims)["groups"]; ok {
		if groupsSlice, ok := groupsRaw.([]interface{}); ok {
			for _, g := range groupsSlice {
				if groupName, ok := g.(string); ok {
					groupNames = append(groupNames, groupName)
				}
			}
		}
	}

	s.logger.Debug("Found groups in token claims",
		slog.String("user_id", userID),
		slog.Any("groups", groupNames))

	// Create authenticated API client using configured API token
	apiClient := api.NewAPIClient(s.apiConfig)
	auth := context.WithValue(ctx, api.ContextAccessToken, s.config.APIToken)

	groups := make([]*entities.Group, 0)
	for _, groupName := range groupNames {
		// Skip admin groups (authentik Admins, etc.)
		if strings.Contains(strings.ToLower(groupName), "admin") {
			continue
		}

		// Query Authentik API to get full group details by name
		groupsResp, _, err := apiClient.CoreApi.CoreGroupsList(auth).Name(groupName).Execute()
		if err != nil {
			s.logger.Warn("Failed to fetch group details from Authentik",
				slog.String("group_name", groupName),
				slog.Any("error", err))
			continue
		}

		// Process each matching group from API response
		for _, ag := range groupsResp.Results {
			// Filter groups to only include those with 'nishiki' role
			if !s.HasNishikiRoleFromAPI(ag) {
				continue
			}

			groupID, err := entities.GroupIDFromString(ag.Pk)
			if err != nil {
				s.logger.Warn("Invalid group ID from Authentik", slog.String("group_id", ag.Pk), slog.Any("error", err))
				continue
			}

			validGroupName, err := entities.NewGroupName(ag.Name)
			if err != nil {
				s.logger.Warn("Invalid group name from Authentik", slog.String("group_name", ag.Name), slog.Any("error", err))
				continue
			}

			// Create group entity - using current time as created/updated since Authentik doesn't provide these
			group := entities.ReconstructGroup(groupID, validGroupName, time.Now(), time.Now())
			groups = append(groups, group)
		}
	}

	s.logger.Debug("Successfully processed user groups",
		slog.String("user_id", userID),
		slog.Int("group_count", len(groups)))

	return groups, nil
}

// hasNishikiRole checks if a group has the 'nishiki' role (legacy method)
func (s *AuthentikAuthService) hasNishikiRole(group AuthentikGroup) bool {
	// Check if group has 'nishiki' role in attributes
	if role, exists := group.Attributes["role"]; exists {
		if roleStr, ok := role.(string); ok && roleStr == "nishiki" {
			return true
		}
	}
	
	// Also check if group name contains 'nishiki' as fallback
	return strings.Contains(strings.ToLower(group.Name), "nishiki")
}

// HasNishikiRoleFromAPI checks if a group has the 'nishiki' role using API client response
func (s *AuthentikAuthService) HasNishikiRoleFromAPI(group api.Group) bool {
	// Check if group has 'nishiki' role in the roles_obj array
	if group.RolesObj != nil {
		for _, role := range group.RolesObj {
			if role.Name == "nishiki" {
				return true
			}
		}
	}

	// Check if group has 'nishiki' role in attributes (legacy support)
	if group.Attributes != nil {
		if role, exists := group.Attributes["role"]; exists {
			if roleStr, ok := role.(string); ok && roleStr == "nishiki" {
				return true
			}
		}
	}

	// Also check if group name contains 'nishiki' as fallback
	if strings.Contains(strings.ToLower(group.Name), "nishiki") {
		return true
	}

	return false
}

// CreateGroup creates a new group in Authentik with nishiki role using API token
func (s *AuthentikAuthService) CreateGroup(ctx context.Context, userToken, name string, creatorID string) (*entities.Group, error) {
	// Create authenticated API client using configured API token
	apiClient := api.NewAPIClient(s.apiConfig)
	auth := context.WithValue(ctx, api.ContextAccessToken, s.config.APIToken)

	s.logger.Debug("Creating group in Authentik", 
		slog.String("group_name", name),
		slog.String("creator_id", creatorID))

	// Create group request with nishiki role attribute
	attributes := map[string]interface{}{
		"role": "nishiki",
	}
	groupRequest := api.GroupRequest{
		Name:         name,
		IsSuperuser:  api.PtrBool(false),
		Attributes:   attributes,
	}

	// Create the group
	createdGroup, httpResp, err := apiClient.CoreApi.CoreGroupsCreate(auth).GroupRequest(groupRequest).Execute()
	if err != nil {
		// Parse detailed error response
		var errorDetail string
		var errorCode string
		
		if httpResp != nil && httpResp.Body != nil {
			if body, readErr := io.ReadAll(httpResp.Body); readErr == nil {
				switch httpResp.StatusCode {
				case http.StatusForbidden:
					var apiError AuthentikForbiddenError
					if jsonErr := json.Unmarshal(body, &apiError); jsonErr == nil {
						errorDetail = apiError.Detail
						errorCode = apiError.Code
					}
				case http.StatusBadRequest:
					var apiError AuthentikValidationError
					if jsonErr := json.Unmarshal(body, &apiError); jsonErr == nil {
						errorCode = apiError.Code
						if len(apiError.NonFieldErrors) > 0 {
							errorDetail = apiError.NonFieldErrors[0]
						}
					}
				}
			}
		}
		
		s.logger.Error("Failed to create group in Authentik", 
			slog.Any("error", err),
			slog.String("detail", errorDetail),
			slog.String("code", errorCode),
			slog.Int("status_code", httpResp.StatusCode))
		
		// Return auth error for 403, validation error for others
		if httpResp != nil && httpResp.StatusCode == http.StatusForbidden {
			return nil, fmt.Errorf("authentication failed: %s", errorDetail)
		}
		return nil, fmt.Errorf("failed to create group: %s", errorDetail)
	}

	// Convert to domain entity
	groupID, err := entities.GroupIDFromString(createdGroup.Pk)
	if err != nil {
		return nil, fmt.Errorf("invalid group ID from Authentik: %w", err)
	}

	groupName, err := entities.NewGroupName(createdGroup.Name)
	if err != nil {
		return nil, fmt.Errorf("invalid group name: %w", err)
	}

	// Add creator as member of the group
	if err := s.addUserToGroupWithToken(auth, createdGroup.Pk, creatorID); err != nil {
		s.logger.Warn("Failed to add creator to group", 
			slog.String("group_id", createdGroup.Pk),
			slog.String("creator_id", creatorID),
			slog.Any("error", err))
	}

	group := entities.ReconstructGroup(groupID, groupName, time.Now(), time.Now())
	
	s.logger.Info("Group created successfully",
		slog.String("group_id", group.ID().String()),
		slog.String("group_name", group.Name().String()),
		slog.String("creator_id", creatorID))

	return group, nil
}

// addUserToGroup adds a user to a group in Authentik (legacy method)
func (s *AuthentikAuthService) addUserToGroup(ctx context.Context, groupID, userID string) error {
	url := fmt.Sprintf("%s/api/v3/core/groups/%s/add_user/", s.config.AuthentikURL, groupID)

	payload := map[string]string{
		"pk": userID,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Note: APIToken removed - using user's JWT token for authentication
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to add user to group: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		s.logger.Error("Failed to add user to group in Authentik", 
			slog.Int("status", resp.StatusCode),
			slog.String("response", string(body)))
		return fmt.Errorf("authentik API returned status %d", resp.StatusCode)
	}

	return nil
}

// addUserToGroupWithToken adds a user to a group in Authentik using API token  
func (s *AuthentikAuthService) addUserToGroupWithToken(ctx context.Context, groupID, userID string) error {
	apiClient := api.NewAPIClient(s.apiConfig)
	auth := context.WithValue(ctx, api.ContextAccessToken, s.config.APIToken)

	s.logger.Debug("Adding user to group", 
		slog.String("group_id", groupID),
		slog.String("user_id", userID))

	// Add user to group
	userPk := cast.ToInt32(userID)
	userAddRequest := api.UserAccountRequest{
		Pk: userPk,
	}

	_, err := apiClient.CoreApi.CoreGroupsAddUserCreate(auth, groupID).UserAccountRequest(userAddRequest).Execute()
	if err != nil {
		s.logger.Error("Failed to add user to group", slog.Any("error", err))
		return fmt.Errorf("failed to add user to group: %w", err)
	}

	return nil
}

// GetGroupUsers fetches users that are members of a group from Authentik using API token
func (s *AuthentikAuthService) GetGroupUsers(ctx context.Context, userToken, groupID string) ([]*entities.User, error) {
	// Create authenticated API client using configured API token
	apiClient := api.NewAPIClient(s.apiConfig)
	auth := context.WithValue(ctx, api.ContextAccessToken, s.config.APIToken)

	s.logger.Debug("Fetching group users from Authentik API", 
		slog.String("group_id", groupID))

	// List users for the group
	usersResp, _, err := apiClient.CoreApi.CoreUsersList(auth).GroupsByPk([]string{groupID}).Execute()
	if err != nil {
		s.logger.Error("Failed to fetch users from Authentik", slog.Any("error", err))
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}

	users := make([]*entities.User, 0, len(usersResp.Results))
	for _, au := range usersResp.Results {
		userID, err := entities.UserIDFromString(cast.ToString(au.Pk))
		if err != nil {
			s.logger.Warn("Invalid user ID from Authentik", slog.String("user_id", cast.ToString(au.Pk)), slog.Any("error", err))
			continue
		}

		username, err := entities.NewUsername(au.Username)
		if err != nil {
			s.logger.Warn("Invalid username from Authentik", slog.String("username", au.Username), slog.Any("error", err))
			continue
		}

		var emailStr string
		if au.Email != nil {
			emailStr = *au.Email
		}
		email, err := entities.NewEmailAddress(emailStr)
		if err != nil {
			s.logger.Warn("Invalid email from Authentik", slog.String("email", emailStr), slog.Any("error", err))
			continue
		}

		user := entities.ReconstructUser(userID, username, email, cast.ToString(au.Pk), time.Now(), time.Now())
		users = append(users, user)
	}

	return users, nil
}

// GetUserByID fetches a single user by ID from Authentik using API token
func (s *AuthentikAuthService) GetUserByID(ctx context.Context, userToken, userID string) (*entities.User, error) {
	// Create authenticated API client using configured API token
	apiClient := api.NewAPIClient(s.apiConfig)
	auth := context.WithValue(ctx, api.ContextAccessToken, s.config.APIToken)

	s.logger.Debug("Fetching user by ID from Authentik API", 
		slog.String("user_id", userID))

	// Get user by ID
	userPk := cast.ToInt32(userID)
	user, _, err := apiClient.CoreApi.CoreUsersRetrieve(auth, userPk).Execute()
	if err != nil {
		s.logger.Error("Failed to fetch user from Authentik", slog.Any("error", err))
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	id, err := entities.UserIDFromString(cast.ToString(user.Pk))
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	username, err := entities.NewUsername(user.Username)
	if err != nil {
		return nil, fmt.Errorf("invalid username: %w", err)
	}

	email, err := entities.NewEmailAddress(*user.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email: %w", err)
	}

	return entities.ReconstructUser(id, username, email, cast.ToString(user.Pk), time.Now(), time.Now()), nil
}

// GetGroupByID fetches a single group by ID from Authentik using API token
func (s *AuthentikAuthService) GetGroupByID(ctx context.Context, userToken, groupID string) (*entities.Group, error) {
	// Create authenticated API client using configured API token
	apiClient := api.NewAPIClient(s.apiConfig)
	auth := context.WithValue(ctx, api.ContextAccessToken, s.config.APIToken)

	s.logger.Debug("Fetching group by ID from Authentik API", 
		slog.String("group_id", groupID))

	// Get group by ID
	group, _, err := apiClient.CoreApi.CoreGroupsRetrieve(auth, groupID).Execute()
	if err != nil {
		s.logger.Error("Failed to fetch group from Authentik", slog.Any("error", err))
		return nil, fmt.Errorf("failed to fetch group: %w", err)
	}

	id, err := entities.GroupIDFromString(group.Pk)
	if err != nil {
		return nil, fmt.Errorf("invalid group ID: %w", err)
	}

	name, err := entities.NewGroupName(group.Name)
	if err != nil {
		return nil, fmt.Errorf("invalid group name: %w", err)
	}

	return entities.ReconstructGroup(id, name, time.Now(), time.Now()), nil
}

// GetOIDCConfig fetches OIDC discovery configuration from Authentik and modifies token endpoint
func (s *AuthentikAuthService) GetOIDCConfig(ctx context.Context, clientID string) (map[string]interface{}, error) {
	// Get the appropriate client
	client, err := s.getClientByClientID(clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to find OAuth client: %w", err)
	}

	// Build discovery URL
	discoveryURL := fmt.Sprintf("%s/application/o/%s/.well-known/openid-configuration",
		s.config.AuthentikURL, client.config.ProviderName)

	s.logger.Debug("Fetching OIDC discovery config",
		slog.String("url", discoveryURL),
		slog.String("provider_name", client.config.ProviderName))

	// Make request to Authentik
	req, err := http.NewRequestWithContext(ctx, "GET", discoveryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error("Failed to fetch OIDC config from Authentik",
			slog.String("error", err.Error()),
			slog.String("url", discoveryURL))
		return nil, fmt.Errorf("failed to fetch OIDC configuration: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("authentik returned status %d for OIDC config", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Error("Failed to read OIDC config response", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to read OIDC configuration: %w", err)
	}

	// Parse JSON response
	var oidcConfig map[string]interface{}
	if err := json.Unmarshal(body, &oidcConfig); err != nil {
		s.logger.Error("Failed to parse OIDC config JSON", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to parse OIDC configuration: %w", err)
	}

	// Replace token_endpoint with our proxy
	backendURL := os.Getenv("BACKEND_URL")
	if backendURL == "" {
		backendURL = "http://localhost:3001" // Default fallback
	}
	oidcConfig["token_endpoint"] = fmt.Sprintf("%s/auth/token", backendURL)

	s.logger.Debug("OIDC config fetched successfully",
		slog.String("token_endpoint", oidcConfig["token_endpoint"].(string)),
		slog.String("provider_name", client.config.ProviderName))

	return oidcConfig, nil
}

// ProxyTokenExchange forwards token exchange requests to Authentik with client credentials
func (s *AuthentikAuthService) ProxyTokenExchange(ctx context.Context, tokenRequest map[string]interface{}) ([]byte, int, error) {
	s.logger.Debug("Processing token exchange request",
		slog.String("grant_type", fmt.Sprintf("%v", tokenRequest["grant_type"])))

	// Determine which client to use based on redirect_uri
	redirectURI, _ := tokenRequest["redirect_uri"].(string)
	client, err := s.getClientByRedirectURL(redirectURI)
	if err != nil {
		s.logger.Error("Failed to determine OAuth client", slog.String("error", err.Error()))
		return nil, http.StatusBadRequest, fmt.Errorf("failed to determine OAuth client: %w", err)
	}

	// Add client credentials from matched client config
	tokenRequest["client_id"] = client.config.ClientID
	tokenRequest["client_secret"] = client.config.ClientSecret

	// Get token URL from OIDC provider configuration
	tokenURL := client.provider.Endpoint().TokenURL

	// Convert to form data
	formData := url.Values{}
	for key, value := range tokenRequest {
		formData.Set(key, fmt.Sprintf("%v", value))
	}

	s.logger.Debug("Forwarding token request to Authentik",
		slog.String("url", tokenURL),
		slog.String("provider_name", client.config.ProviderName))

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Make POST request to Authentik
	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error("Failed to forward token request to Authentik",
			slog.String("error", err.Error()),
			slog.String("url", tokenURL))
		return nil, 0, fmt.Errorf("failed to exchange token: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Error("Failed to read token exchange response", slog.String("error", err.Error()))
		return nil, 0, fmt.Errorf("failed to read token response: %w", err)
	}

	s.logger.Debug("Token exchange completed",
		slog.Int("status_code", resp.StatusCode),
		slog.Int("response_size", len(body)),
		slog.String("provider_name", client.config.ProviderName))

	return body, resp.StatusCode, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
