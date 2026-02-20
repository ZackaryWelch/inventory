package request

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/nishiki/backend-go/domain/entities"
)

func GetUserIDFromPath(r *http.Request) (entities.UserID, error) {
	idStr := r.PathValue("id")
	if idStr == "" {
		return entities.UserID{}, fmt.Errorf("missing user ID in path")
	}

	if _, err := uuid.Parse(idStr); err != nil {
		return entities.UserID{}, fmt.Errorf("invalid user ID format: not a valid UUID")
	}

	userID, err := entities.UserIDFromString(idStr)
	if err != nil {
		return entities.UserID{}, fmt.Errorf("invalid user ID: %w", err)
	}

	return userID, nil
}
