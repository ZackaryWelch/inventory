package app

import (
	"fmt"
	"io"
	"log/slog"
	"testing"

	"gioui.org/widget"
)

// newTestGioApp creates a minimal GioApp suitable for cache/filter unit tests.
func newTestGioApp() *GioApp {
	return &GioApp{
		logger:      slog.New(slog.NewTextHandler(io.Discard, nil)),
		widgetState: &WidgetState{},
	}
}

// makeObjects creates n objects with the given grouped-text property values.
// Each object gets properties: {"category": values[i % len(values)]}.
func makeObjects(n int, propKey string, values []string) []Object {
	objects := make([]Object, n)
	for i := range objects {
		objects[i] = Object{
			ID:          fmt.Sprintf("obj-%d", i),
			Name:        fmt.Sprintf("Object %d", i),
			Description: fmt.Sprintf("Description for object %d", i),
			ContainerID: fmt.Sprintf("cont-%d", i%5),
			Properties: map[string]TypedValue{
				propKey: TypedValue{Val: values[i%len(values)]},
			},
		}
	}
	return objects
}

func makeSchema(propKey, propType string) *PropertySchema {
	return &PropertySchema{
		Definitions: []PropertyDefinition{
			{Key: propKey, DisplayName: "Category", Type: propType},
		},
	}
}

// --- Unit tests ---

func TestMapsEqual(t *testing.T) {
	tests := []struct {
		name string
		a, b map[string]string
		want bool
	}{
		{"both nil", nil, nil, true},
		{"equal", map[string]string{"a": "1"}, map[string]string{"a": "1"}, true},
		{"different value", map[string]string{"a": "1"}, map[string]string{"a": "2"}, false},
		{"different length", map[string]string{"a": "1"}, map[string]string{"a": "1", "b": "2"}, false},
		{"nil vs empty", nil, map[string]string{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mapsEqual(tt.a, tt.b); got != tt.want {
				t.Errorf("mapsEqual(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestCopyStringMap(t *testing.T) {
	orig := map[string]string{"a": "1", "b": "2"}
	cp := copyStringMap(orig)
	if !mapsEqual(orig, cp) {
		t.Fatal("copy should equal original")
	}
	cp["a"] = "changed"
	if orig["a"] == "changed" {
		t.Fatal("modifying copy should not affect original")
	}
}

func TestCopyStringMapNil(t *testing.T) {
	if copyStringMap(nil) != nil {
		t.Fatal("copying nil should return nil")
	}
}

func TestSnakeToTitleCase(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"hello_world", "Hello World"},
		{"single", "Single"},
		{"", ""},
		{"already_Title", "Already Title"},
	}
	for _, tt := range tests {
		if got := snakeToTitleCase(tt.input); got != tt.want {
			t.Errorf("snakeToTitleCase(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestCollectGroupedTextValues(t *testing.T) {
	ga := newTestGioApp()
	ga.objects = makeObjects(10, "category", []string{"Electronics", "Books", "Music"})
	ga.selectedCollection = &Collection{
		PropertySchema: makeSchema("category", "grouped_text"),
	}

	values := ga.collectGroupedTextValues()
	if values == nil {
		t.Fatal("expected non-nil values")
	}
	cats := values["category"]
	if len(cats) != 3 {
		t.Errorf("expected 3 unique categories, got %d: %v", len(cats), cats)
	}
	// Should be sorted
	for i := 1; i < len(cats); i++ {
		if cats[i] < cats[i-1] {
			t.Errorf("values not sorted: %v", cats)
			break
		}
	}
}

func TestCollectGroupedTextValues_NoSchema(t *testing.T) {
	ga := newTestGioApp()
	ga.objects = makeObjects(5, "category", []string{"A"})
	ga.selectedCollection = &Collection{}

	if got := ga.collectGroupedTextValues(); got != nil {
		t.Errorf("expected nil without schema, got %v", got)
	}
}

func TestCollectGroupedTextValues_NonGroupedText(t *testing.T) {
	ga := newTestGioApp()
	ga.objects = makeObjects(5, "category", []string{"A"})
	ga.selectedCollection = &Collection{
		PropertySchema: makeSchema("category", "text"), // not grouped_text
	}

	if got := ga.collectGroupedTextValues(); got != nil {
		t.Errorf("expected nil for non-grouped_text type, got %v", got)
	}
}

func TestGetGroupedTextValues_Cache(t *testing.T) {
	ga := newTestGioApp()
	ga.objects = makeObjects(10, "category", []string{"A", "B"})
	ga.selectedCollection = &Collection{
		PropertySchema: makeSchema("category", "grouped_text"),
	}

	v1 := ga.getGroupedTextValues()
	if !ga.cachedGroupedTextValid {
		t.Fatal("cache should be valid after first call")
	}

	v2 := ga.getGroupedTextValues()
	// Should return same map pointer (cached)
	if fmt.Sprintf("%p", v1) != fmt.Sprintf("%p", v2) {
		t.Error("second call should return cached pointer")
	}

	ga.invalidateObjectCaches()
	if ga.cachedGroupedTextValid {
		t.Fatal("cache should be invalid after invalidation")
	}

	v3 := ga.getGroupedTextValues()
	if fmt.Sprintf("%p", v1) == fmt.Sprintf("%p", v3) {
		t.Error("should return new map after invalidation")
	}
}

func TestGetPropertyDefMap(t *testing.T) {
	ga := newTestGioApp()
	ga.selectedCollection = &Collection{
		PropertySchema: &PropertySchema{
			Definitions: []PropertyDefinition{
				{Key: "brand", DisplayName: "Brand", Type: "text"},
				{Key: "price", DisplayName: "Price", Type: "currency", CurrencyCode: "USD"},
			},
		},
	}

	m := ga.getPropertyDefMap()
	if len(m) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(m))
	}
	if m["brand"].DisplayName != "Brand" {
		t.Error("wrong display name for brand")
	}
	if m["price"].CurrencyCode != "USD" {
		t.Error("wrong currency code for price")
	}

	// Verify caching
	m2 := ga.getPropertyDefMap()
	if fmt.Sprintf("%p", m) != fmt.Sprintf("%p", m2) {
		t.Error("second call should return cached map")
	}
}

func TestGetPropertyDefMap_NoSchema(t *testing.T) {
	ga := newTestGioApp()
	ga.selectedCollection = &Collection{}

	m := ga.getPropertyDefMap()
	if len(m) != 0 {
		t.Errorf("expected empty map without schema, got %v", m)
	}
}

func TestGetFilteredObjects(t *testing.T) {
	ga := newTestGioApp()
	ga.objects = []Object{
		{ID: "1", Name: "Arduino Board", Description: "Microcontroller"},
		{ID: "2", Name: "Raspberry Pi", Description: "Single-board computer"},
		{ID: "3", Name: "HDMI Cable", Description: "Audio/video cable"},
	}

	// No filter — all returned
	objs, indices := ga.getFilteredObjects()
	if len(objs) != 3 {
		t.Errorf("expected 3 objects, got %d", len(objs))
	}

	// Search by name
	ga.widgetState.objectsSearchField.SetText("arduino")
	ga.cachedFilteredObjects = nil // force recompute
	objs, indices = ga.getFilteredObjects()
	if len(objs) != 1 || objs[0].ID != "1" {
		t.Errorf("expected 1 match for 'arduino', got %d", len(objs))
	}
	if indices[0] != 0 {
		t.Errorf("expected index 0, got %d", indices[0])
	}

	// Search by description
	ga.widgetState.objectsSearchField.SetText("cable")
	ga.cachedFilteredObjects = nil
	objs, _ = ga.getFilteredObjects()
	if len(objs) != 1 || objs[0].ID != "3" {
		t.Errorf("expected 1 match for 'cable', got %d", len(objs))
	}
}

func TestGetFilteredObjects_Cache(t *testing.T) {
	ga := newTestGioApp()
	ga.objects = makeObjects(100, "cat", []string{"A", "B"})

	objs1, _ := ga.getFilteredObjects()
	objs2, _ := ga.getFilteredObjects()
	if fmt.Sprintf("%p", objs1) != fmt.Sprintf("%p", objs2) {
		t.Error("second call with same inputs should return cached slice")
	}

	// Changing search should recompute
	ga.widgetState.objectsSearchField.SetText("Object 1")
	objs3, _ := ga.getFilteredObjects()
	if len(objs3) == len(objs1) {
		t.Error("different search should yield different results")
	}
}

func TestGetFilteredObjects_GroupedTextFilter(t *testing.T) {
	ga := newTestGioApp()
	ga.objects = makeObjects(3, "color", []string{"red", "blue", "red"})
	ga.activeGroupedTextFilters = map[string]string{"color": "red"}

	objs, _ := ga.getFilteredObjects()
	if len(objs) != 2 {
		t.Errorf("expected 2 red objects, got %d", len(objs))
	}
}

func TestGetFilteredContainers(t *testing.T) {
	ga := newTestGioApp()
	ga.containers = []Container{
		{ID: "1", Name: "Shelf A", Type: "shelf", Location: "Room 1"},
		{ID: "2", Name: "Cabinet B", Type: "cabinet", Location: "Room 2"},
		{ID: "3", Name: "Shelf C", Type: "shelf", Location: "Room 1"},
	}

	// No filter
	conts, _ := ga.getFilteredContainers()
	if len(conts) != 3 {
		t.Errorf("expected 3 containers, got %d", len(conts))
	}

	// Search by name
	ga.widgetState.containersSearchField.SetText("cabinet")
	ga.cachedFilteredContainers = nil
	conts, _ = ga.getFilteredContainers()
	if len(conts) != 1 || conts[0].ID != "2" {
		t.Errorf("expected 1 match for 'cabinet', got %d", len(conts))
	}

	// Search by location
	ga.widgetState.containersSearchField.SetText("room 1")
	ga.cachedFilteredContainers = nil
	conts, _ = ga.getFilteredContainers()
	if len(conts) != 2 {
		t.Errorf("expected 2 matches for 'room 1', got %d", len(conts))
	}
}

func TestInvalidateObjectCaches(t *testing.T) {
	ga := newTestGioApp()
	ga.objects = makeObjects(10, "cat", []string{"A"})
	ga.selectedCollection = &Collection{
		PropertySchema: makeSchema("cat", "grouped_text"),
	}

	// Populate all caches
	ga.getGroupedTextValues()
	ga.getPropertyDefMap()
	ga.getFilteredObjects()

	ga.containers = []Container{{ID: "1", Name: "X", Type: "shelf"}}
	ga.getFilteredContainers()

	// Verify populated
	if !ga.cachedGroupedTextValid || !ga.cachedPropertyDefValid {
		t.Fatal("caches should be valid")
	}
	if ga.cachedFilteredObjects == nil || ga.cachedFilteredContainers == nil {
		t.Fatal("filtered caches should be populated")
	}

	ga.invalidateObjectCaches()

	if ga.cachedGroupedTextValid || ga.cachedPropertyDefValid {
		t.Error("caches should be invalid after invalidation")
	}
	if ga.cachedFilteredObjects != nil || ga.cachedFilteredContainers != nil {
		t.Error("filtered caches should be nil after invalidation")
	}
}

// --- Benchmarks ---

func BenchmarkCollectGroupedTextValues(b *testing.B) {
	for _, n := range []int{100, 500, 1000, 5000} {
		b.Run(fmt.Sprintf("objects=%d", n), func(b *testing.B) {
			ga := newTestGioApp()
			ga.objects = makeObjects(n, "category", []string{"Electronics", "Books", "Music", "Games", "Tools"})
			ga.selectedCollection = &Collection{
				PropertySchema: makeSchema("category", "grouped_text"),
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ga.collectGroupedTextValues()
			}
		})
	}
}

func BenchmarkGetGroupedTextValues_Cached(b *testing.B) {
	ga := newTestGioApp()
	ga.objects = makeObjects(1000, "category", []string{"A", "B", "C", "D", "E"})
	ga.selectedCollection = &Collection{
		PropertySchema: makeSchema("category", "grouped_text"),
	}
	ga.getGroupedTextValues() // prime cache
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ga.getGroupedTextValues()
	}
}

func BenchmarkFilterObjects(b *testing.B) {
	for _, n := range []int{100, 500, 1000, 5000} {
		b.Run(fmt.Sprintf("objects=%d/no_query", n), func(b *testing.B) {
			ga := newTestGioApp()
			ga.objects = makeObjects(n, "cat", []string{"A", "B"})
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ga.cachedFilteredObjects = nil
				ga.getFilteredObjects()
			}
		})
		b.Run(fmt.Sprintf("objects=%d/with_query", n), func(b *testing.B) {
			ga := newTestGioApp()
			ga.objects = makeObjects(n, "cat", []string{"A", "B"})
			ga.widgetState.objectsSearchField.SetText("Object 1")
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ga.cachedFilteredObjects = nil
				ga.getFilteredObjects()
			}
		})
		b.Run(fmt.Sprintf("objects=%d/cached", n), func(b *testing.B) {
			ga := newTestGioApp()
			ga.objects = makeObjects(n, "cat", []string{"A", "B"})
			ga.getFilteredObjects() // prime cache
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ga.getFilteredObjects()
			}
		})
	}
}

func BenchmarkMatchesGroupedTextFilters(b *testing.B) {
	ga := newTestGioApp()
	ga.activeGroupedTextFilters = map[string]string{
		"category": "Electronics",
		"brand":    "Sony",
	}
	obj := Object{
		Properties: map[string]TypedValue{
			"category": {Val: "Electronics"},
			"brand":    {Val: "Sony"},
			"color":    {Val: "black"},
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ga.matchesGroupedTextFilters(obj)
	}
}

func BenchmarkPropertyDisplayName(b *testing.B) {
	defs := make([]PropertyDefinition, 20)
	for i := range defs {
		defs[i] = PropertyDefinition{
			Key:         fmt.Sprintf("prop_%d", i),
			DisplayName: fmt.Sprintf("Property %d", i),
			Type:        "text",
		}
	}
	targetKey := "prop_19" // worst case: last element

	b.Run("slice_scan", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			propertyDisplayName(targetKey, defs)
		}
	})

	defMap := make(map[string]*PropertyDefinition, len(defs))
	for i := range defs {
		defMap[defs[i].Key] = &defs[i]
	}

	b.Run("map_lookup", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			propertyDisplayNameFromMap(targetKey, defMap)
		}
	})
}

// BenchmarkFilterContainers measures container filtering performance.
func BenchmarkFilterContainers(b *testing.B) {
	for _, n := range []int{50, 200, 500} {
		b.Run(fmt.Sprintf("containers=%d", n), func(b *testing.B) {
			ga := newTestGioApp()
			ga.containers = make([]Container, n)
			for i := range ga.containers {
				ga.containers[i] = Container{
					ID:       fmt.Sprintf("cont-%d", i),
					Name:     fmt.Sprintf("Container %d", i),
					Type:     "shelf",
					Location: fmt.Sprintf("Room %d", i%10),
				}
			}
			ga.widgetState.containersSearchField.SetText("Container 1")
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ga.cachedFilteredContainers = nil
				ga.getFilteredContainers()
			}
		})
	}
}

// BenchmarkWidgetEditorText measures the cost of reading widget.Editor.Text()
// since this is called on every cache-hit check.
func BenchmarkWidgetEditorText(b *testing.B) {
	var ed widget.Editor
	ed.SetText("some search query")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ed.Text()
	}
}
