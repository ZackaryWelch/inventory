package services

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/nishiki/backend-go/domain/entities"
)

// TypeInferenceService infers property schemas from tabular data and coerces values.
type TypeInferenceService struct{}

// NewTypeInferenceService creates a new TypeInferenceService.
func NewTypeInferenceService() *TypeInferenceService {
	return &TypeInferenceService{}
}

var (
	currencyPrefixRe = regexp.MustCompile(`^\s*[\$€£¥]\s*[\d,]+(\.\d+)?\s*$`)
	urlRe            = regexp.MustCompile(`(?i)^https?://`)
	pdfRe            = regexp.MustCompile(`(?i)\.pdf$`)
	boolTrueRe       = regexp.MustCompile(`(?i)^(true|yes|1)$`)
	boolFalseRe      = regexp.MustCompile(`(?i)^(false|no|0|never|todo|null|n/a)$`)
	dateFmts         = []string{
		time.RFC3339,
		"2006-01-02",
		"01/02/2006",
		"1/2/2006",
		"2006/01/02",
		"January 2, 2006",
		"Jan 2, 2006",
	}
	// Matches prefix-dates like "<2020"
	prefixDateRe = regexp.MustCompile(`^<\d{4}$`)
)

// inferColumnType determines the PropertyType for a single column based on sample values.
// Threshold: a type is assigned if >= 80% of non-empty values match.
func inferColumnType(values []string) entities.PropertyType {
	nonEmpty := make([]string, 0, len(values))
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			nonEmpty = append(nonEmpty, strings.TrimSpace(v))
		}
	}
	if len(nonEmpty) == 0 {
		return entities.PropertyTypeText
	}

	threshold := float64(len(nonEmpty)) * 0.8

	counts := map[entities.PropertyType]int{}
	for _, v := range nonEmpty {
		switch {
		case isBool(v):
			counts[entities.PropertyTypeBool]++
		case isDate(v):
			counts[entities.PropertyTypeDate]++
		case isCurrency(v):
			counts[entities.PropertyTypeCurrency]++
		case isURL(v):
			counts[entities.PropertyTypeURL]++
		case isNumeric(v):
			counts[entities.PropertyTypeNumeric]++
		}
	}

	// Check in priority order
	for _, t := range []entities.PropertyType{
		entities.PropertyTypeBool,
		entities.PropertyTypeDate,
		entities.PropertyTypeCurrency,
		entities.PropertyTypeURL,
		entities.PropertyTypeNumeric,
	} {
		if float64(counts[t]) >= threshold {
			return t
		}
	}

	// Check for grouped_text: unique values < 30% of total rows AND > 1 unique value
	uniqueVals := uniqueStrings(nonEmpty)
	if len(uniqueVals) > 1 && float64(len(uniqueVals)) < float64(len(nonEmpty))*0.3 {
		return entities.PropertyTypeGroupedText
	}

	return entities.PropertyTypeText
}

func isBool(v string) bool {
	return boolTrueRe.MatchString(v) || boolFalseRe.MatchString(v)
}

func isDate(v string) bool {
	if prefixDateRe.MatchString(v) {
		return true
	}
	for _, fmt := range dateFmts {
		if _, err := time.Parse(fmt, v); err == nil {
			return true
		}
	}
	return false
}

func isCurrency(v string) bool {
	return currencyPrefixRe.MatchString(v)
}

func isURL(v string) bool {
	return urlRe.MatchString(v) || pdfRe.MatchString(v)
}

func isNumeric(v string) bool {
	cleaned := strings.ReplaceAll(v, ",", "")
	_, err := strconv.ParseFloat(cleaned, 64)
	return err == nil
}

func uniqueStrings(vals []string) []string {
	seen := make(map[string]struct{}, len(vals))
	out := make([]string, 0)
	for _, v := range vals {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			out = append(out, v)
		}
	}
	return out
}

// toSnakeCase converts a display header to a snake_case storage key.
func toSnakeCase(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	// Replace spaces and hyphens with underscores
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "-", "_")
	// Remove characters that aren't alphanumeric or underscore
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// reservedColumns maps known CSV column names to their Object field roles.
// These are NOT stored as properties.
var reservedColumns = map[string]string{
	"name":        "name",
	"title":       "name",
	"item":        "name",
	"description": "description",
	"quantity":    "quantity",
	"tags":        "tags",
	"location":    "location",
}

// InferSchema infers a PropertySchema from CSV headers and sample data rows.
func (s *TypeInferenceService) InferSchema(headers []string, data []map[string]interface{}) *entities.PropertySchema {
	if len(data) == 0 || len(headers) == 0 {
		return nil
	}

	// Collect column values
	colValues := make(map[string][]string, len(headers))
	for _, h := range headers {
		colValues[h] = make([]string, 0, len(data))
	}
	for _, row := range data {
		for _, h := range headers {
			if v, ok := row[h]; ok {
				colValues[h] = append(colValues[h], fmt.Sprintf("%v", v))
			} else {
				colValues[h] = append(colValues[h], "")
			}
		}
	}

	defs := make([]entities.PropertyDefinition, 0, len(headers))
	for _, h := range headers {
		normalizedKey := toSnakeCase(h)
		// Skip reserved columns (they map to Object fields, not properties)
		if _, reserved := reservedColumns[strings.ToLower(normalizedKey)]; reserved {
			continue
		}

		colType := inferColumnType(colValues[h])
		def := entities.PropertyDefinition{
			Key:         normalizedKey,
			DisplayName: h,
			Type:        colType,
			Required:    false,
		}
		// Add USD as default currency code for currency columns
		if colType == entities.PropertyTypeCurrency {
			def.CurrencyCode = "USD"
		}
		defs = append(defs, def)
	}

	if len(defs) == 0 {
		return nil
	}
	return &entities.PropertySchema{Definitions: defs}
}

// CoerceValue converts a raw value to the appropriate Go type for the given PropertyType.
// On failure it returns the original value as a string.
func (s *TypeInferenceService) CoerceValue(value interface{}, targetType entities.PropertyType) interface{} {
	if value == nil {
		return nil
	}
	str := strings.TrimSpace(fmt.Sprintf("%v", value))
	if str == "" {
		return nil
	}

	switch targetType {
	case entities.PropertyTypeCurrency, entities.PropertyTypeNumeric:
		cleaned := strings.ReplaceAll(str, ",", "")
		cleaned = strings.TrimLeft(cleaned, "$€£¥ ")
		if f, err := strconv.ParseFloat(cleaned, 64); err == nil {
			return f
		}
		return str

	case entities.PropertyTypeDate:
		if prefixDateRe.MatchString(str) {
			// Return as-is; represents a partial date like "<2020"
			return str
		}
		for _, fmt := range dateFmts {
			if t, err := time.Parse(fmt, str); err == nil {
				return t.Format(time.RFC3339)
			}
		}
		return str

	case entities.PropertyTypeBool:
		if boolTrueRe.MatchString(str) {
			return true
		}
		return false

	case entities.PropertyTypeURL:
		return str // URLs stored as strings

	default: // text, grouped_text
		return str
	}
}

// CoerceRow applies CoerceValue to all properties in a row according to the schema.
func (s *TypeInferenceService) CoerceRow(row map[string]interface{}, schema *entities.PropertySchema) map[string]interface{} {
	if schema == nil {
		return row
	}
	result := make(map[string]interface{}, len(row))
	for k, v := range row {
		result[k] = v
	}
	for _, def := range schema.Definitions {
		if v, ok := result[def.Key]; ok {
			result[def.Key] = s.CoerceValue(v, def.Type)
		}
	}
	return result
}
