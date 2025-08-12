package controllers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nishiki/backend-go/app/container"
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
func (ctrl *AuthController) GetCurrentUser(c *gin.Context) {
	user, userExists := middleware.GetCurrentUser(c)
	if !userExists {
		ctrl.logger.Error("No authenticated user found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	claims, claimsExist := middleware.GetCurrentClaims(c)
	if !claimsExist {
		ctrl.logger.Error("No auth claims found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	ctrl.logger.Debug("Current user retrieved successfully",
		slog.String("user_id", user.ID().String()),
		slog.String("username", user.Username().String()))

	c.JSON(http.StatusOK, response.NewAuthInfoResponse(user, claims))
}

// HealthCheck godoc
// @Summary Health check endpoint
// @Description Check if the service is healthy
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func (ctrl *AuthController) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "nishiki-backend-go",
	})
}

// GetOIDCConfig godoc
// @Summary Get OIDC discovery configuration
// @Description Proxy OIDC discovery configuration from Authentik to avoid CORS issues
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /auth/oidc-config [get]
func (ctrl *AuthController) GetOIDCConfig(c *gin.Context) {
	oidcConfig, err := ctrl.authService.GetOIDCConfig(c.Request.Context())
	if err != nil {
		ctrl.logger.Error("Failed to get OIDC config", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch OIDC configuration"})
		return
	}

	// Set CORS headers and return config
	c.Header("Access-Control-Allow-Origin", "*")
	c.JSON(http.StatusOK, oidcConfig)
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
func (ctrl *AuthController) ProxyTokenExchange(c *gin.Context) {
	// Get request body - handle both form-encoded and JSON
	var requestBody map[string]interface{}

	contentType := c.GetHeader("Content-Type")
	if contentType == "application/x-www-form-urlencoded" {
		// Parse form data first
		if err := c.Request.ParseForm(); err != nil {
			ctrl.logger.Error("Failed to parse form data", slog.String("error", err.Error()))
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid form data"})
			return
		}

		// Handle form-encoded data (standard OAuth) using Gin's PostForm
		requestBody = make(map[string]interface{})
		for key, values := range c.Request.PostForm {
			if len(values) == 1 {
				requestBody[key] = values[0]
			} else {
				requestBody[key] = values
			}
		}
	} else {
		// Handle JSON data
		if err := c.ShouldBindJSON(&requestBody); err != nil {
			ctrl.logger.Error("Failed to parse JSON token exchange request", slog.String("error", err.Error()))
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
	}

	// Call auth service to handle token exchange
	responseBody, statusCode, err := ctrl.authService.ProxyTokenExchange(c.Request.Context(), requestBody)
	if err != nil {
		ctrl.logger.Error("Failed to exchange token", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to exchange token"})
		return
	}

	// Set CORS headers
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "POST, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Content-Type")

	// Forward status code and response body from auth service
	c.Data(statusCode, "application/json", responseBody)
}
