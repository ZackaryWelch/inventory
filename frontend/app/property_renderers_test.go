package app

import (
	"fmt"
	"testing"
)

func TestRenderPropertyValueFromMap(t *testing.T) {
	defMap := map[string]*PropertyDefinition{
		"price":    {Key: "price", Type: "currency", CurrencyCode: "USD"},
		"bought":   {Key: "bought", Type: "date"},
		"active":   {Key: "active", Type: "bool"},
		"weight":   {Key: "weight", Type: "numeric"},
		"link":     {Key: "link", Type: "url"},
		"category": {Key: "category", Type: "grouped_text"},
		"name":     {Key: "name", Type: "text"},
	}

	tests := []struct {
		key   string
		value interface{}
		want  string
	}{
		{"price", 29.99, "$29.99"},
		{"price", "19.5", "$19.50"},
		{"bought", "2024-06-15", "Jun 15, 2024"},
		{"bought", "2024-06-15T10:00:00Z", "Jun 15, 2024"},
		{"active", true, "Yes"},
		{"active", false, "No"},
		{"active", "true", "Yes"},
		{"active", "false", "No"},
		{"weight", 3.14, "3.14"},
		{"weight", "42", "42"},
		{"link", "https://example.com/path/to/file.pdf", "file.pdf"},
		{"link", "", ""},
		{"category", "Electronics", "Electronics"},
		{"name", "Test Item", "Test Item"},
		{"unknown_key", "value", "value"}, // not in map
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s=%v", tt.key, tt.value), func(t *testing.T) {
			got := RenderPropertyValueFromMap(tt.key, tt.value, defMap)
			if got != tt.want {
				t.Errorf("RenderPropertyValueFromMap(%q, %v) = %q, want %q", tt.key, tt.value, got, tt.want)
			}
		})
	}
}

func TestRenderPropertyValueFromMap_MatchesOriginal(t *testing.T) {
	defs := []PropertyDefinition{
		{Key: "price", Type: "currency", CurrencyCode: "EUR"},
		{Key: "date", Type: "date"},
		{Key: "flag", Type: "bool"},
		{Key: "count", Type: "numeric"},
		{Key: "url", Type: "url"},
		{Key: "tag", Type: "grouped_text"},
	}
	defMap := make(map[string]*PropertyDefinition, len(defs))
	for i := range defs {
		defMap[defs[i].Key] = &defs[i]
	}

	cases := []struct {
		key   string
		value interface{}
	}{
		{"price", 99.99},
		{"date", "2025-01-15"},
		{"flag", true},
		{"flag", false},
		{"count", 42.0},
		{"url", "https://example.com/foo/bar.txt"},
		{"tag", "hello"},
		{"missing", "fallback"},
	}

	for _, tc := range cases {
		v1 := RenderPropertyValue(tc.key, tc.value, defs)
		v2 := RenderPropertyValueFromMap(tc.key, tc.value, defMap)
		if v1 != v2 {
			t.Errorf("mismatch for key=%q value=%v: slice=%q map=%q", tc.key, tc.value, v1, v2)
		}
	}
}

func TestPropertyDisplayNameFromMap(t *testing.T) {
	defMap := map[string]*PropertyDefinition{
		"brand":       {Key: "brand", DisplayName: "Brand Name"},
		"empty_name":  {Key: "empty_name", DisplayName: ""},
		"serial_code": {Key: "serial_code", DisplayName: "Serial Code"},
	}

	tests := []struct {
		key  string
		want string
	}{
		{"brand", "Brand Name"},
		{"empty_name", "Empty Name"},   // falls back to snake_case conversion
		{"serial_code", "Serial Code"}, // uses display name
		{"unknown_key", "Unknown Key"}, // not in map, uses snake_case
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := propertyDisplayNameFromMap(tt.key, defMap)
			if got != tt.want {
				t.Errorf("propertyDisplayNameFromMap(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestPropertyDisplayNameFromMap_MatchesOriginal(t *testing.T) {
	defs := []PropertyDefinition{
		{Key: "brand", DisplayName: "Brand"},
		{Key: "serial_number", DisplayName: ""},
		{Key: "purchase_date", DisplayName: "Date Purchased"},
	}
	defMap := make(map[string]*PropertyDefinition, len(defs))
	for i := range defs {
		defMap[defs[i].Key] = &defs[i]
	}

	for _, key := range []string{"brand", "serial_number", "purchase_date", "not_in_schema"} {
		v1 := propertyDisplayName(key, defs)
		v2 := propertyDisplayNameFromMap(key, defMap)
		if v1 != v2 {
			t.Errorf("mismatch for key=%q: slice=%q map=%q", key, v1, v2)
		}
	}
}

// --- Benchmarks ---

func BenchmarkRenderPropertyValue(b *testing.B) {
	defs := make([]PropertyDefinition, 20)
	for i := range defs {
		defs[i] = PropertyDefinition{
			Key:  fmt.Sprintf("prop_%d", i),
			Type: "text",
		}
	}
	defs[15] = PropertyDefinition{Key: "price", Type: "currency", CurrencyCode: "USD"}

	defMap := make(map[string]*PropertyDefinition, len(defs))
	for i := range defs {
		defMap[defs[i].Key] = &defs[i]
	}

	b.Run("slice/text", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			RenderPropertyValue("prop_19", "hello", defs)
		}
	})
	b.Run("map/text", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			RenderPropertyValueFromMap("prop_19", "hello", defMap)
		}
	})
	b.Run("slice/currency", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			RenderPropertyValue("price", 29.99, defs)
		}
	})
	b.Run("map/currency", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			RenderPropertyValueFromMap("price", 29.99, defMap)
		}
	})
}
