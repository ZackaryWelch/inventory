package services

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/nishiki/backend/domain/entities"
)

// TypeInferenceService infers property schemas from tabular data and coerces values.
type TypeInferenceService struct {
	reservedCols map[string]struct{}
}

// defaultReservedColumns is the built-in list used when none is supplied.
var defaultReservedColumns = []string{
	"name", "title", "item",
	"description", "quantity", "tags", "location",
}

// NewTypeInferenceService creates a new TypeInferenceService.
// reservedColumns is a list of snake_case column names that map to Object fields
// and must not be stored as properties. Pass nil to use the built-in defaults.
func NewTypeInferenceService(reservedColumns []string) *TypeInferenceService {
	if len(reservedColumns) == 0 {
		reservedColumns = defaultReservedColumns
	}
	cols := make(map[string]struct{}, len(reservedColumns))
	for _, c := range reservedColumns {
		cols[c] = struct{}{}
	}
	return &TypeInferenceService{reservedCols: cols}
}

// IsReserved reports whether the given snake_case key is a reserved Object field
// that should not be stored as a property.
func (s *TypeInferenceService) IsReserved(key string) bool {
	_, ok := s.reservedCols[key]
	return ok
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
	// Matches prefix-dates like "<2020" or "~2020"
	prefixDateRe = regexp.MustCompile(`^<\d{4}$`)
	approxDateRe = regexp.MustCompile(`^~\d{4}$`)
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
	if prefixDateRe.MatchString(v) || approxDateRe.MatchString(v) {
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

// ToSnakeCase converts a display header to a snake_case storage key.
func ToSnakeCase(s string) string {
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

// InferSchema infers a PropertySchema from CSV headers and sample data rows.
func (s *TypeInferenceService) InferSchema(headers []string, data []map[string]any) *entities.PropertySchema {
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
		normalizedKey := ToSnakeCase(h)
		// Skip reserved columns (they map to Object fields, not properties)
		if s.IsReserved(normalizedKey) {
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

// CoerceValue converts a raw value to a TypedValue for the given PropertyType.
// On failure it returns the original value as a text TypedValue.
func (s *TypeInferenceService) CoerceValue(value any, targetType entities.PropertyType) entities.TypedValue {
	if value == nil {
		return entities.TypedValue{Type: targetType, Val: nil}
	}
	str := strings.TrimSpace(fmt.Sprintf("%v", value))
	if str == "" {
		return entities.TypedValue{Type: targetType, Val: nil}
	}

	switch targetType {
	case entities.PropertyTypeCurrency, entities.PropertyTypeNumeric:
		cleaned := strings.ReplaceAll(str, ",", "")
		cleaned = strings.TrimLeft(cleaned, "$€£¥ ")
		if f, err := strconv.ParseFloat(cleaned, 64); err == nil {
			return entities.TypedValue{Type: targetType, Val: f}
		}
		return entities.TypedValue{Type: entities.PropertyTypeText, Val: str}

	case entities.PropertyTypeDate:
		if prefixDateRe.MatchString(str) {
			year, _ := strconv.Atoi(str[1:])
			t := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
			return entities.TypedValue{Type: entities.PropertyTypeDate, Val: t, Approx: true}
		}
		if approxDateRe.MatchString(str) {
			year, _ := strconv.Atoi(str[1:])
			t := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
			return entities.TypedValue{Type: entities.PropertyTypeDate, Val: t, Approx: true}
		}
		for _, fmt := range dateFmts {
			if t, err := time.Parse(fmt, str); err == nil {
				return entities.TypedValue{Type: entities.PropertyTypeDate, Val: t}
			}
		}
		return entities.TypedValue{Type: entities.PropertyTypeText, Val: str}

	case entities.PropertyTypeBool:
		if boolTrueRe.MatchString(str) {
			return entities.TypedValue{Type: entities.PropertyTypeBool, Val: true}
		}
		return entities.TypedValue{Type: entities.PropertyTypeBool, Val: false}

	case entities.PropertyTypeURL:
		return entities.TypedValue{Type: entities.PropertyTypeURL, Val: str}

	default: // text, grouped_text
		return entities.TypedValue{Type: targetType, Val: str}
	}
}

// CoerceValueWithDef converts a raw value to a TypedValue using a PropertyDefinition,
// which also sets the Currency field for currency-typed properties.
func (s *TypeInferenceService) CoerceValueWithDef(value any, def *entities.PropertyDefinition) entities.TypedValue {
	tv := s.CoerceValue(value, def.Type)
	if def.Type == entities.PropertyTypeCurrency && def.CurrencyCode != "" {
		tv.Currency = def.CurrencyCode
	}
	return tv
}

// CoerceRawProperties coerces a raw map[string]interface{} into map[string]TypedValue
// using the collection's PropertySchema. Schema-defined keys use their definition's type;
// unknown keys are wrapped as text.
func (s *TypeInferenceService) CoerceRawProperties(raw map[string]any, schema *entities.PropertySchema) map[string]entities.TypedValue {
	result := make(map[string]entities.TypedValue, len(raw))
	for k, v := range raw {
		if schema != nil {
			if def := schema.GetDefinition(k); def != nil {
				result[k] = s.CoerceValueWithDef(v, def)
				continue
			}
		}
		result[k] = entities.TypedValue{Type: entities.PropertyTypeText, Val: fmt.Sprintf("%v", v)}
	}
	return result
}

// NormalizeRowKeys rewrites every key in a data row to its snake_case equivalent.
// This ensures row keys match the snake_case keys stored in PropertySchema.Definitions.
func (s *TypeInferenceService) NormalizeRowKeys(row map[string]any) map[string]any {
	result := make(map[string]any, len(row))
	for k, v := range row {
		result[ToSnakeCase(k)] = v
	}
	return result
}

// CoerceRow applies CoerceValueWithDef to schema-defined keys and wraps unknown keys as text TypedValues.
func (s *TypeInferenceService) CoerceRow(row map[string]any, schema *entities.PropertySchema) map[string]entities.TypedValue {
	result := make(map[string]entities.TypedValue, len(row))
	for k, v := range row {
		if schema != nil {
			if def := schema.GetDefinition(k); def != nil {
				result[k] = s.CoerceValueWithDef(v, def)
				continue
			}
		}
		result[k] = entities.TypedValue{Type: entities.PropertyTypeText, Val: fmt.Sprintf("%v", v)}
	}
	return result
}
