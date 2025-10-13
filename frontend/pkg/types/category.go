package types

import "time"

// Category represents a category for organizing objects
type Category struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Icon      string    `json:"icon"`  // Emoji or icon identifier
	Color     string    `json:"color"` // Hex color code
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateCategoryRequest represents the request to create a new category
type CreateCategoryRequest struct {
	Name  string `json:"name"`
	Icon  string `json:"icon"`
	Color string `json:"color"`
}

// UpdateCategoryRequest represents the request to update a category
type UpdateCategoryRequest struct {
	Name  string `json:"name"`
	Icon  string `json:"icon"`
	Color string `json:"color"`
}

// FoodCategory represents a predefined food category
type FoodCategory struct {
	Name  string
	Emoji string
}

// FoodCategories maps category IDs to their information (matching React frontend)
var FoodCategories = map[string]FoodCategory{
	"unselected":       {Name: "Unselected", Emoji: "🥣"},
	"beverage":         {Name: "Beverage", Emoji: "☕️"},
	"dairy":            {Name: "Dairy", Emoji: "🥛"},
	"eggs":             {Name: "Egg", Emoji: "🥚"},
	"fatsAndOils":      {Name: "Fat & Oil", Emoji: "🫒"},
	"fruits":           {Name: "Fruit", Emoji: "🍎"},
	"vegetables":       {Name: "Vegetable", Emoji: "🥗"},
	"legumes":          {Name: "Legume", Emoji: "🫘"},
	"nutsAndSeeds":     {Name: "Nut & Seed", Emoji: "🥜"},
	"meat":             {Name: "Meat", Emoji: "🥩"},
	"desserts":         {Name: "Dessert", Emoji: "🍰"},
	"soup":             {Name: "Soup", Emoji: "🍜"},
	"seafoods":         {Name: "Seafood", Emoji: "🍣"},
	"convenienceMeals": {Name: "Convenience Meal", Emoji: "🥡"},
	"seasoning":        {Name: "Seasoning", Emoji: "🧂"},
	"alcohol":          {Name: "Alcohol", Emoji: "🍺"},
	"other":            {Name: "Other", Emoji: "🍽️"},
}
