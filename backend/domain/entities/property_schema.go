package entities

// PropertyType defines the type of a property value for rich rendering and coercion.
type PropertyType string

const (
	PropertyTypeText        PropertyType = "text"         // default, plain string
	PropertyTypeCurrency    PropertyType = "currency"     // float64, rendered with currency symbol
	PropertyTypeDate        PropertyType = "date"         // time.Time (ISO 8601 string in storage)
	PropertyTypeBool        PropertyType = "bool"         // true/false
	PropertyTypeURL         PropertyType = "url"          // string, rendered as clickable link
	PropertyTypeNumeric     PropertyType = "numeric"      // float64
	PropertyTypeGroupedText PropertyType = "grouped_text" // text with unique value grouping for filtering
)

// PropertyDefinition describes a single property field in a collection's schema.
type PropertyDefinition struct {
	Key          string       `json:"key"`                     // normalized snake_case key for storage
	DisplayName  string       `json:"display_name"`            // original column header name
	Type         PropertyType `json:"type"`                    // semantic type for rendering/coercion
	Required     bool         `json:"required"`                // whether this property must be present
	CurrencyCode string       `json:"currency_code,omitempty"` // e.g. "USD", only for currency type
}

// PropertySchema defines the typed schema for object properties in a collection.
type PropertySchema struct {
	Definitions []PropertyDefinition `json:"definitions"`
}

// Validate checks that the properties map conforms to required fields in the schema.
// Returns a list of validation error messages; empty slice means valid.
func (ps *PropertySchema) Validate(properties map[string]interface{}) []string {
	if ps == nil {
		return nil
	}
	var errs []string
	for _, def := range ps.Definitions {
		if def.Required {
			if _, ok := properties[def.Key]; !ok {
				errs = append(errs, "missing required property: "+def.Key)
			}
		}
	}
	return errs
}

// GetDefinition looks up a property definition by key. Returns nil if not found.
func (ps *PropertySchema) GetDefinition(key string) *PropertyDefinition {
	if ps == nil {
		return nil
	}
	for i := range ps.Definitions {
		if ps.Definitions[i].Key == key {
			return &ps.Definitions[i]
		}
	}
	return nil
}
