package controllers

import (
	"log/slog"
	"net/http"

	"github.com/nishiki/backend-go/app/container"
	"github.com/nishiki/backend-go/app/http/httputil"
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
func (ctrl *UserController) GetUser(w http.ResponseWriter, r *http.Request) {
	currentUser, exists := middleware.GetCurrentUser(r)
	if !exists {
		ctrl.logger.Error("No authenticated user found in context")
		httputil.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}

	userToken, tokenExists := middleware.GetCurrentToken(r)
	if !tokenExists {
		ctrl.logger.Error("No auth token found in context")
		httputil.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}

	userID, err := request.GetUserIDFromPath(r)
	if err != nil {
		ctrl.logger.Warn("Invalid user ID in path", slog.Any("error", err))
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := ctrl.authService.GetUserByID(r.Context(), userToken, userID.String())
	if err != nil {
		ctrl.logger.Error("Failed to get user", slog.Any("error", err))
		if err.Error() == "user not found" {
			httputil.Error(w, http.StatusNotFound, "user not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	ctrl.logger.Debug("User retrieved successfully",
		slog.String("user_id", userID.String()),
		slog.String("requested_by", currentUser.ID().String()))

	httputil.JSON(w, http.StatusOK, response.NewUserResponse(user))
}
