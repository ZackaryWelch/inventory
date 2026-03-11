package request

import (
	"fmt"
	"net/http"

	"github.com/nishiki/backend-go/domain/entities"
)

func GetUserIDFromPath(r *http.Request) (entities.UserID, error) {
	idStr := r.PathValue("id")
	if idStr == "" {
		return entities.UserID{}, fmt.Errorf("missing user ID in path")
	}

	userID, err := entities.UserIDFromString(idStr)
	if err != nil {
		return entities.UserID{}, fmt.Errorf("invalid user ID: %w", err)
	}

	return userID, nil
}
