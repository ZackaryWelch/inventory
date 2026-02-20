package controllers

import (
	"log/slog"
	"net/http"

	"github.com/nishiki/backend-go/app/container"
	"github.com/nishiki/backend-go/app/http/httputil"
	"github.com/nishiki/backend-go/app/http/middleware"
	"github.com/nishiki/backend-go/app/http/response"
	"github.com/nishiki/backend-go/domain/services"
)

type AuthController struct {
	container   *container.Container
	logger      *slog.Logger
	authService services.AuthService
}

func NewAuthController(appContainer *container.Container, logger *slog.Logger) *AuthController {
	return &AuthController{
		container:   appContainer,
		logger:      logger,
		authService: appContainer.AuthService,
	}
}

// GetCurrentUser godoc
// @Summary Get current user information
// @Description Get information about the currently authenticated user
// @Tags auth
// @Produce json
// @Success 200 {object} response.AuthInfoResponse
// @Failure 401 {object} map[string]string
// @Router /auth/me [get]
// @Security BearerAuth
func (ctrl *AuthController) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	user, userExists := middleware.GetCurrentUser(r)
	if !userExists {
		ctrl.logger.Error("No authenticated user found in context")
		httputil.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}

	claims, claimsExist := middleware.GetCurrentClaims(r)
	if !claimsExist {
		ctrl.logger.Error("No auth claims found in context")
		httputil.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}

	ctrl.logger.Debug("Current user retrieved successfully",
		slog.String("user_id", user.ID().String()),
		slog.String("username", user.Username().String()))

	httputil.JSON(w, http.StatusOK, response.NewAuthInfoResponse(user, claims))
}

// HealthCheck godoc
// @Summary Health check endpoint
// @Description Check if the service is healthy
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func (ctrl *AuthController) HealthCheck(w http.ResponseWriter, r *http.Request) {
	httputil.JSON(w, http.StatusOK, map[string]string{
		"status":  "healthy",
		"service": "nishiki-backend-go",
	})
}

// GetOIDCConfig godoc
// @Summary Get OIDC discovery configuration
// @Description Proxy OIDC discovery configuration from Authentik to avoid CORS issues
// @Tags auth
// @Produce json
// @Param client_id query string true "OAuth client ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/oidc-config [get]
func (ctrl *AuthController) GetOIDCConfig(w http.ResponseWriter, r *http.Request) {
	// Get client_id from query parameter
	clientID := r.URL.Query().Get("client_id")
	if clientID == "" {
		ctrl.logger.Error("Missing client_id parameter")
		httputil.Error(w, http.StatusBadRequest, "client_id query parameter is required")
		return
	}

	oidcConfig, err := ctrl.authService.GetOIDCConfig(r.Context(), clientID)
	if err != nil {
		ctrl.logger.Error("Failed to get OIDC config", slog.String("error", err.Error()))
		httputil.Error(w, http.StatusInternalServerError, "failed to fetch OIDC configuration")
		return
	}

	// Set CORS headers and return config
	w.Header().Set("Access-Control-Allow-Origin", "*")
	httputil.JSON(w, http.StatusOK, oidcConfig)
}

// ProxyTokenExchange godoc
// @Summary Proxy token exchange to Authentik
// @Description Proxy token exchange requests to Authentik with proper credentials and CORS handling
// @Tags auth
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Token exchange request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/token [post]
func (ctrl *AuthController) ProxyTokenExchange(w http.ResponseWriter, r *http.Request) {
	// Get request body - handle both form-encoded and JSON
	var requestBody map[string]interface{}

	contentType := r.Header.Get("Content-Type")
	if contentType == "application/x-www-form-urlencoded" {
		// Parse form data first
		if err := r.ParseForm(); err != nil {
			ctrl.logger.Error("Failed to parse form data", slog.String("error", err.Error()))
			httputil.Error(w, http.StatusBadRequest, "invalid form data")
			return
		}

		// Handle form-encoded data (standard OAuth)
		requestBody = make(map[string]interface{})
		for key, values := range r.PostForm {
			if len(values) == 1 {
				requestBody[key] = values[0]
			} else {
				requestBody[key] = values
			}
		}
	} else {
		// Handle JSON data
		if err := httputil.DecodeJSON(r, &requestBody); err != nil {
			ctrl.logger.Error("Failed to parse JSON token exchange request", slog.String("error", err.Error()))
			httputil.Error(w, http.StatusBadRequest, "invalid request body")
			return
		}
	}

	// Call auth service to handle token exchange (redirect_uri used to determine client)
	responseBody, statusCode, err := ctrl.authService.ProxyTokenExchange(r.Context(), requestBody)
	if err != nil {
		ctrl.logger.Error("Failed to exchange token", slog.String("error", err.Error()))
		// If it's a bad request (couldn't determine client), return 400
		if statusCode == http.StatusBadRequest {
			httputil.Error(w, http.StatusBadRequest, err.Error())
		} else {
			httputil.Error(w, http.StatusInternalServerError, "failed to exchange token")
		}
		return
	}

	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Forward status code and response body from auth service
	httputil.Data(w, statusCode, "application/json", responseBody)
}
