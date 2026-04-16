package usecases

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/nishiki/backend/domain/entities"
	"github.com/nishiki/backend/domain/repositories"
	"github.com/nishiki/backend/domain/services"
)

// fixedFields are the standard Object fields always included in a CSV export, in order.
var fixedFields = []string{"name", "description", "quantity", "unit", "tags", "expires_at"}

var fixedFieldSet = func() map[string]struct{} {
	m := make(map[string]struct{}, len(fixedFields))
	for _, f := range fixedFields {
		m[f] = struct{}{}
	}
	return m
}()

type ExportCollectionRequest struct {
	CollectionID entities.CollectionID
	UserID       entities.UserID
	UserToken    string
}

type ExportCollectionResponse struct {
	CSV            []byte
	CollectionName string
}

type ExportCollectionUseCase struct {
	collectionRepo repositories.CollectionRepository
	authService    services.AuthService
}

func NewExportCollectionUseCase(collectionRepo repositories.CollectionRepository, authService services.AuthService) *ExportCollectionUseCase {
	return &ExportCollectionUseCase{
		collectionRepo: collectionRepo,
		authService:    authService,
	}
}

func (uc *ExportCollectionUseCase) Execute(ctx context.Context, req ExportCollectionRequest) (*ExportCollectionResponse, error) {
	userGroups, err := uc.authService.GetUserGroups(ctx, req.UserToken, req.UserID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get user groups: %w", err)
	}

	collection, err := uc.collectionRepo.GetByID(ctx, req.CollectionID)
	if err != nil {
		return nil, fmt.Errorf("collection not found: %w", err)
	}

	hasAccess := collection.UserID().Equals(req.UserID)
	if !hasAccess && collection.GroupID() != nil {
		for _, group := range userGroups {
			if group.ID().Equals(*collection.GroupID()) {
				hasAccess = true
				break
			}
		}
	}
	if !hasAccess {
		return nil, errors.New("access denied: user does not have access to this collection")
	}

	objects := collection.GetAllObjects()
	schema := collection.PropertySchema()

	// Determine property columns (schema-defined keys that are not fixed fields).
	var propKeys []string
	var propHeaders []string
	if schema != nil && len(schema.Definitions) > 0 {
		for _, def := range schema.Definitions {
			if _, isFixed := fixedFieldSet[def.Key]; !isFixed {
				propKeys = append(propKeys, def.Key)
				propHeaders = append(propHeaders, def.DisplayName)
			}
		}
	} else {
		keySet := make(map[string]struct{})
		for _, obj := range objects {
			for k := range obj.Properties() {
				if _, isFixed := fixedFieldSet[k]; !isFixed {
					keySet[k] = struct{}{}
				}
			}
		}
		for k := range keySet {
			propKeys = append(propKeys, k)
		}
		sort.Strings(propKeys)
		propHeaders = propKeys
	}

	// Build header row: fixed fields first, then property columns.
	header := make([]string, 0, len(fixedFields)+len(propHeaders))
	for _, key := range fixedFields {
		header = append(header, displayNameForKey(key, schema))
	}
	header = append(header, propHeaders...)

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	if err := w.Write(header); err != nil {
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	for _, obj := range objects {
		row := make([]string, 0, len(fixedFields)+len(propKeys))
		for _, key := range fixedFields {
			row = append(row, extractFixedField(obj, key))
		}
		props := obj.Properties()
		for _, key := range propKeys {
			val := ""
			if tv, ok := props[key]; ok && tv.Val != nil {
				val = tv.DisplayString()
			}
			row = append(row, val)
		}
		if err := w.Write(row); err != nil {
			return nil, fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("failed to flush CSV: %w", err)
	}

	return &ExportCollectionResponse{
		CSV:            buf.Bytes(),
		CollectionName: collection.Name().String(),
	}, nil
}

// displayNameForKey returns the schema-defined display name for a key, or auto-derives it
// from the snake_case key (e.g. "expires_at" → "Expires At").
func displayNameForKey(key string, schema *entities.PropertySchema) string {
	if schema != nil {
		if def := schema.GetDefinition(key); def != nil && def.DisplayName != "" {
			return def.DisplayName
		}
	}
	parts := strings.Split(key, "_")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, " ")
}

// extractFixedField returns the string value of a fixed Object field by key.
func extractFixedField(obj entities.Object, key string) string {
	switch key {
	case "name":
		return obj.Name().String()
	case "description":
		return obj.Description().String()
	case "quantity":
		if obj.Quantity() != nil {
			return fmt.Sprintf("%g", *obj.Quantity())
		}
		return ""
	case "unit":
		return obj.Unit()
	case "tags":
		return strings.Join(obj.Tags(), "|")
	case "expires_at":
		if obj.ExpiresAt() != nil {
			return obj.ExpiresAt().Format(time.RFC3339)
		}
		return ""
	}
	return ""
}
