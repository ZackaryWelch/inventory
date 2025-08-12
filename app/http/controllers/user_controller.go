package controllers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nishiki/backend-go/app/container"
	"github.com/nishiki/backend-go/app/http/middleware"
	"github.com/nishiki/backend-go/app/http/request"
	"github.com/nishiki/backend-go/app/http/response"
	"github.com/nishiki/backend-go/domain/services"
)

type UserController struct {
	authService services.AuthService
	logger      *slog.Logger
}

func NewUserController(
	c *container.Container,
	logger *slog.Logger,
) *UserController {
	return &UserController{
		authService: c.AuthService,
		logger:      logger,
	}
}

// GetUser godoc
// @Summary Get user by ID
// @Description Fetch user information by ID
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} response.UserResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/{id} [get]
// @Security BearerAuth
func (ctrl *UserController) GetUser(c *gin.Context) {
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		ctrl.logger.Error("No authenticated user found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	userToken, tokenExists := middleware.GetCurrentToken(c)
	if !tokenExists {
		ctrl.logger.Error("No auth token found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	userID, err := request.GetUserIDFromPath(c)
	if err != nil {
		ctrl.logger.Warn("Invalid user ID in path", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := ctrl.authService.GetUserByID(c.Request.Context(), userToken, userID.String())
	if err != nil {
		ctrl.logger.Error("Failed to get user", slog.Any("error", err))
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	ctrl.logger.Debug("User retrieved successfully",
		slog.String("user_id", userID.String()),
		slog.String("requested_by", currentUser.ID().String()))

	c.JSON(http.StatusOK, response.NewUserResponse(user))
}
