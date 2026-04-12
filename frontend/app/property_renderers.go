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

// RenderPropertyValueFromMap is like RenderPropertyValue but uses a pre-built map for O(1) lookup.
func RenderPropertyValueFromMap(key string, tv TypedValue, defMap map[string]*PropertyDefinition) string {
	def := defMap[key]
	return renderTypedValue(tv, def)
}

// renderTypedValue formats a TypedValue using an already-resolved definition (may be nil).
func renderTypedValue(tv TypedValue, def *PropertyDefinition) string {
	if tv.Val == nil {
		return ""
	}

	// Use the TypedValue's own type for rendering; fall back to def.Type if empty.
	tvType := tv.Type
	if tvType == "" && def != nil {
		tvType = def.Type
	}

	// Currency symbol: prefer TypedValue.Currency, fall back to def.CurrencyCode.
	currencyCode := tv.Currency
	if currencyCode == "" && def != nil {
		currencyCode = def.CurrencyCode
	}

	switch tvType {
	case "currency":
		f, err := toFloat(tv.Val)
		if err != nil {
			return fmt.Sprintf("%v", tv.Val)
		}
		symbol := currencySymbols[strings.ToUpper(currencyCode)]
		if symbol == "" {
			if currencyCode != "" {
				symbol = currencyCode + " "
			} else {
				symbol = "$"
			}
		}
		return fmt.Sprintf("%s%.2f", symbol, f)

	case "date":
		var t time.Time
		switch v := tv.Val.(type) {
		case string:
			var err error
			t, err = time.Parse(time.RFC3339, v)
			if err != nil {
				t, err = time.Parse("2006-01-02", v)
				if err != nil {
					return v
				}
			}
		case float64:
			// JSON numbers for time (ms since epoch) — unlikely but handled
			t = time.UnixMilli(int64(v)).UTC()
		default:
			return fmt.Sprintf("%v", v)
		}
		formatted := t.Format("Jan 2, 2006")
		if tv.Approx {
			return "~" + t.Format("2006")
		}
		return formatted

	case "bool":
		switch v := tv.Val.(type) {
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
			f, err := toFloat(tv.Val)
			if err != nil {
				return fmt.Sprintf("%v", tv.Val)
			}
			if f != 0 {
				return "Yes"
			}
			return "No"
		}

	case "url":
		str := fmt.Sprintf("%v", tv.Val)
		if str == "" {
			return str
		}
		base := path.Base(str)
		if base == "." || base == "/" {
			return str
		}
		return base

	case "numeric":
		f, err := toFloat(tv.Val)
		if err != nil {
			return fmt.Sprintf("%v", tv.Val)
		}
		return strconv.FormatFloat(f, 'f', -1, 64)

	default: // "text", "grouped_text", unknown
		return fmt.Sprintf("%v", tv.Val)
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
