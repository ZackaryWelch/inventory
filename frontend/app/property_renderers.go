package app

import (
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"
)

var currencySymbols = map[string]string{
	"USD": "$",
	"EUR": "€",
	"GBP": "£",
	"JPY": "¥",
	"CAD": "CA$",
	"AUD": "A$",
}

// RenderPropertyValue formats a property value for display using schema type info.
// Pass nil or an empty slice when no schema is available; falls back to plain string conversion.
func RenderPropertyValue(key string, value interface{}, defs []PropertyDefinition) string {
	def := findPropertyDef(key, defs)
	if def == nil {
		return fmt.Sprintf("%v", value)
	}

	str := fmt.Sprintf("%v", value)

	switch def.Type {
	case "currency":
		f, err := toFloat(value)
		if err != nil {
			return str
		}
		symbol := currencySymbols[strings.ToUpper(def.CurrencyCode)]
		if symbol == "" {
			if def.CurrencyCode != "" {
				symbol = def.CurrencyCode + " "
			} else {
				symbol = "$"
			}
		}
		return fmt.Sprintf("%s%.2f", symbol, f)

	case "date":
		t, err := time.Parse(time.RFC3339, str)
		if err != nil {
			t, err = time.Parse("2006-01-02", str)
			if err != nil {
				return str
			}
		}
		return t.Format("Jan 2, 2006")

	case "bool":
		switch v := value.(type) {
		case bool:
			if v {
				return "Yes"
			}
			return "No"
		case string:
			if strings.EqualFold(v, "true") || v == "1" {
				return "Yes"
			}
			return "No"
		default:
			f, err := toFloat(value)
			if err != nil {
				return str
			}
			if f != 0 {
				return "Yes"
			}
			return "No"
		}

	case "url":
		if str == "" {
			return str
		}
		base := path.Base(str)
		if base == "." || base == "/" {
			return str
		}
		return base

	case "numeric":
		f, err := toFloat(value)
		if err != nil {
			return str
		}
		return strconv.FormatFloat(f, 'f', -1, 64)

	default: // "text", "grouped_text", unknown
		return str
	}
}

func findPropertyDef(key string, defs []PropertyDefinition) *PropertyDefinition {
	for i := range defs {
		if defs[i].Key == key {
			return &defs[i]
		}
	}
	return nil
}

// RenderPropertyValueFromMap is like RenderPropertyValue but uses a pre-built map for O(1) lookup.
func RenderPropertyValueFromMap(key string, value interface{}, defMap map[string]*PropertyDefinition) string {
	def := defMap[key]
	return renderPropertyValueWithDef(value, def)
}

// renderPropertyValueWithDef formats a property value using an already-resolved definition.
func renderPropertyValueWithDef(value interface{}, def *PropertyDefinition) string {
	if def == nil {
		return fmt.Sprintf("%v", value)
	}

	str := fmt.Sprintf("%v", value)

	switch def.Type {
	case "currency":
		f, err := toFloat(value)
		if err != nil {
			return str
		}
		symbol := currencySymbols[strings.ToUpper(def.CurrencyCode)]
		if symbol == "" {
			if def.CurrencyCode != "" {
				symbol = def.CurrencyCode + " "
			} else {
				symbol = "$"
			}
		}
		return fmt.Sprintf("%s%.2f", symbol, f)

	case "date":
		t, err := time.Parse(time.RFC3339, str)
		if err != nil {
			t, err = time.Parse("2006-01-02", str)
			if err != nil {
				return str
			}
		}
		return t.Format("Jan 2, 2006")

	case "bool":
		switch v := value.(type) {
		case bool:
			if v {
				return "Yes"
			}
			return "No"
		case string:
			if strings.EqualFold(v, "true") || v == "1" {
				return "Yes"
			}
			return "No"
		default:
			f, err := toFloat(value)
			if err != nil {
				return str
			}
			if f != 0 {
				return "Yes"
			}
			return "No"
		}

	case "url":
		if str == "" {
			return str
		}
		base := path.Base(str)
		if base == "." || base == "/" {
			return str
		}
		return base

	case "numeric":
		f, err := toFloat(value)
		if err != nil {
			return str
		}
		return strconv.FormatFloat(f, 'f', -1, 64)

	default: // "text", "grouped_text", unknown
		return str
	}
}

func toFloat(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", value)
	}
}
