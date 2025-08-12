package request

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/nishiki/backend-go/domain/entities"
)

func GetUserIDFromPath(c *gin.Context) (entities.UserID, error) {
	idStr := c.Param("id")
	if idStr == "" {
		return entities.UserID{}, fmt.Errorf("missing user ID in path")
	}

	userID, err := entities.UserIDFromString(idStr)
	if err != nil {
		return entities.UserID{}, fmt.Errorf("invalid user ID: %w", err)
	}

	return userID, nil
}
