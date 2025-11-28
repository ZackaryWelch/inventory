# Plan 2: Go Backend Architecture for Recipe Knowledge Base

## Overview
Build a lightweight Go HTTP server that:
1. Loads and indexes the YAML knowledge base at startup
2. Provides query endpoints for recipe context retrieval
3. Integrates with Fabric patterns (via HTTP calls) or directly calls Ollama
4. Serves as the backbone for recipe critique, prep planning, and modification suggestions

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                   Fabric Pattern (or CLI)                    │
│  (Calls Go backend endpoints to retrieve context)            │
└─────────────────┬───────────────────────────────────────────┘
                  │ HTTP GET/POST
                  ↓
┌─────────────────────────────────────────────────────────────┐
│                  Go HTTP Server                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │ Recipe Context Router                                │   │
│  │ - GET /api/ingredient/:name                          │   │
│  │ - GET /api/ingredients/batch                         │   │
│  │ - POST /api/recipe/analyze                           │   │
│  │ - GET /api/substitutions?from=X&to=Y                 │   │
│  │ - GET /api/technique/:name                           │   │
│  │ - POST /api/critique (sends to Ollama)               │   │
│  └──────────────────────────────────────────────────────┘   │
│  ┌──────────────────────────────────────────────────────┐   │
│  │ Knowledge Base Indexer                               │   │
│  │ - Loads YAML at startup                              │   │
│  │ - Builds in-memory index (ingredient → profiles)     │   │
│  │ - Caches frequently accessed queries                 │   │
│  └──────────────────────────────────────────────────────┘   │
│  ┌──────────────────────────────────────────────────────┐   │
│  │ Ollama Client                                        │   │
│  │ - Calls local Ollama instance                        │   │
│  │ - Manages prompt construction with context           │   │
│  │ - Streams or returns responses                       │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                  │
                  ↓
┌─────────────────────────────────────────────────────────────┐
│         YAML Knowledge Base (~/.recipe_context/)            │
│  - flavor_profiles.yaml                                     │
│  - substitutions.yaml                                       │
│  - techniques.yaml                                          │
│  - interactions.yaml                                        │
└─────────────────────────────────────────────────────────────┘
                  │
                  ↓
┌─────────────────────────────────────────────────────────────┐
│              Local Ollama Instance                          │
│              (Mistral 7B or Llama 2 10B)                    │
└─────────────────────────────────────────────────────────────┘
```

## Phase 1: Go Project Structure

### 1.1 Directory Layout

```
recipe-backend/
├── cmd/
│   └── server/
│       └── main.go              # Entry point
├── internal/
│   ├── models/
│   │   ├── ingredient.go        # Data structures
│   │   ├── technique.go
│   │   └── recipe.go
│   ├── loader/
│   │   └── yaml_loader.go       # Load YAML files
│   ├── indexer/
│   │   └── knowledge_index.go   # In-memory index
│   ├── handlers/
│   │   ├── ingredient_handler.go
│   │   ├── recipe_handler.go
│   │   └── critique_handler.go
│   ├── ollama/
│   │   └── client.go            # Ollama HTTP client
│   └── prompt/
│       └── constructor.go       # Build prompts with context
├── go.mod
├── go.sum
└── config.yaml                  # Server config (port, YAML path, Ollama URL)
```

### 1.2 Core Data Structures

```go
// internal/models/ingredient.go

type IngredientProfile struct {
	Name              string            `yaml:"name"`
	FlavorProfile     string            `yaml:"flavor_profile"`
	Texture           string            `yaml:"texture"`
	AbsorbsFlavors    bool              `yaml:"absorbs_flavors"`
	ClassicPairings   []Pairing         `yaml:"classic_pairings"`
	CookingMethods    map[string]Method `yaml:"cooking_methods"`
	PrepNotes         []string          `yaml:"prep_notes"`
	Substitutions     map[string]Sub    `yaml:"substitutions"`
}

type Pairing struct {
	Name      string `yaml:"name"`
	Reasoning string `yaml:"reasoning"`
}

type Method struct {
	Temp  string `yaml:"temp"`
	Time  string `yaml:"time"`
	Notes string `yaml:"notes"`
}

type Sub struct {
	Compatibility string `yaml:"compatibility"`
	Adjustments   string `yaml:"adjustments"`
	Notes         string `yaml:"notes"`
}

// internal/models/technique.go

type Technique struct {
	Name           string   `yaml:"name"`
	Equipment      string   `yaml:"equipment"`
	OilType        string   `yaml:"oil_type"`
	Prep           []string `yaml:"prep"`
	HeatLevel      string   `yaml:"heat_level"`
	OilTemp        string   `yaml:"oil_temp"`
	Timing         string   `yaml:"timing"`
	DonenessSign   string   `yaml:"doneness_signs"`
	CommonMistakes []string `yaml:"common_mistakes"`
	Notes          string   `yaml:"notes"`
}

// internal/models/recipe.go

type RecipeAnalysis struct {
	Ingredients       []string          `json:"ingredients"`
	CookingMethods    []string          `json:"cooking_methods"`
	ContextRetrieved  map[string]interface{} `json:"context_retrieved"`
	SuggestedChanges  []string          `json:"suggested_changes"`
}
```

## Phase 2: Core Functionality

### 2.1 YAML Loader

```go
// internal/loader/yaml_loader.go

package loader

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
	"recipe-backend/internal/models"
)

type KnowledgeBase struct {
	FlavorProfiles map[string]*models.IngredientProfile `yaml:"flavor_profiles"`
	Substitutions  map[string]map[string]*models.Sub    `yaml:"substitutions"`
	Techniques     map[string]*models.Technique         `yaml:"techniques"`
	Interactions   []map[string]interface{}             `yaml:"interactions"`
}

func LoadKnowledgeBase(yamlDir string) (*KnowledgeBase, error) {
	kb := &KnowledgeBase{
		FlavorProfiles: make(map[string]*models.IngredientProfile),
		Substitutions:  make(map[string]map[string]*models.Sub),
		Techniques:     make(map[string]*models.Technique),
	}

	// Load each YAML file
	files := []string{
		"flavor_profiles.yaml",
		"substitutions.yaml",
		"techniques.yaml",
		"interactions.yaml",
	}

	for _, file := range files {
		filePath := filepath.Join(yamlDir, file)
		data, err := os.ReadFile(filePath)
		if err != nil {
			// Graceful degradation: warn but continue
			continue
		}

		switch file {
		case "flavor_profiles.yaml":
			var profiles map[string]*models.IngredientProfile
			if err := yaml.Unmarshal(data, &profiles); err != nil {
				return nil, err
			}
			kb.FlavorProfiles = profiles
		case "techniques.yaml":
			var techniques map[string]*models.Technique
			if err := yaml.Unmarshal(data, &techniques); err != nil {
				return nil, err
			}
			kb.Techniques = techniques
		// Similar for others...
		}
	}

	return kb, nil
}
```

### 2.2 Knowledge Index

```go
// internal/indexer/knowledge_index.go

package indexer

import (
	"strings"
	"sync"

	"recipe-backend/internal/models"
)

type KnowledgeIndex struct {
	profiles      map[string]*models.IngredientProfile
	techniques    map[string]*models.Technique
	interactions  []interface{}
	mu            sync.RWMutex
	cache         map[string]interface{}
}

func NewIndex(kb *KnowledgeBase) *KnowledgeIndex {
	return &KnowledgeIndex{
		profiles:     kb.FlavorProfiles,
		techniques:   kb.Techniques,
		interactions: kb.Interactions,
		cache:        make(map[string]interface{}),
	}
}

// GetIngredient retrieves a single ingredient profile
func (ki *KnowledgeIndex) GetIngredient(name string) *models.IngredientProfile {
	ki.mu.RLock()
	defer ki.mu.RUnlock()

	// Normalize: lowercase, trim spaces
	normalized := strings.ToLower(strings.TrimSpace(name))

	if profile, ok := ki.profiles[normalized]; ok {
		return profile
	}
	return nil
}

// GetIngredientBatch retrieves multiple ingredients (efficient for recipe analysis)
func (ki *KnowledgeIndex) GetIngredientBatch(names []string) map[string]*models.IngredientProfile {
	ki.mu.RLock()
	defer ki.mu.RUnlock()

	result := make(map[string]*models.IngredientProfile)
	for _, name := range names {
		normalized := strings.ToLower(strings.TrimSpace(name))
		if profile, ok := ki.profiles[normalized]; ok {
			result[normalized] = profile
		}
	}
	return result
}

// GetTechnique retrieves cooking method details
func (ki *KnowledgeIndex) GetTechnique(methodName string) *models.Technique {
	ki.mu.RLock()
	defer ki.mu.RUnlock()

	normalized := strings.ToLower(strings.TrimSpace(methodName))
	if tech, ok := ki.techniques[normalized]; ok {
		return tech
	}
	return nil
}

// FindRelatedInteractions finds relevant interaction patterns
func (ki *KnowledgeIndex) FindRelatedInteractions(ingredient1, ingredient2 string) []interface{} {
	ki.mu.RLock()
	defer ki.mu.RUnlock()

	var related []interface{}
	lower1 := strings.ToLower(ingredient1)
	lower2 := strings.ToLower(ingredient2)

	for _, interaction := range ki.interactions {
		// Simple string matching in interaction descriptions
		iStr := fmt.Sprintf("%v", interaction)
		if strings.Contains(iStr, lower1) && strings.Contains(iStr, lower2) {
			related = append(related, interaction)
		}
	}
	return related
}
```

### 2.3 HTTP Handlers

```go
// internal/handlers/ingredient_handler.go

package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"recipe-backend/internal/indexer"
)

type IngredientHandler struct {
	index *indexer.KnowledgeIndex
}

func NewIngredientHandler(index *indexer.KnowledgeIndex) *IngredientHandler {
	return &IngredientHandler{index: index}
}

// GET /api/ingredient/:name
func (h *IngredientHandler) GetIngredient(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/ingredient/")
	
	profile := h.index.GetIngredient(name)
	if profile == nil {
		http.Error(w, "ingredient not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

// POST /api/ingredients/batch
// Body: {"ingredients": ["eggplant", "tomato", "garlic"]}
func (h *IngredientHandler) GetBatch(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Ingredients []string `json:"ingredients"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	profiles := h.index.GetIngredientBatch(req.Ingredients)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profiles)
}
```

### 2.4 Ollama Client

```go
// internal/ollama/client.go

package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type OllamaClient struct {
	baseURL string
	model   string
}

func NewOllamaClient(baseURL, model string) *OllamaClient {
	return &OllamaClient{
		baseURL: baseURL,
		model:   model,
	}
}

type GenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type GenerateResponse struct {
	Response string `json:"response"`
}

// Generate sends a prompt to Ollama and returns the response
func (c *OllamaClient) Generate(prompt string) (string, error) {
	req := GenerateRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: false,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(
		fmt.Sprintf("%s/api/generate", c.baseURL),
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Response, nil
}

// GenerateStream streams response for long outputs
func (c *OllamaClient) GenerateStream(prompt string) (io.ReadCloser, error) {
	req := GenerateRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: true,
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(
		fmt.Sprintf("%s/api/generate", c.baseURL),
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}
```

### 2.5 Prompt Constructor

```go
// internal/prompt/constructor.go

package prompt

import (
	"fmt"
	"strings"

	"recipe-backend/internal/models"
)

type PromptBuilder struct {
	systemPrompt string
	contextParts []string
}

func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{
		systemPrompt: `You are a recipe critic and meal prep planner. 
You have access to a curated knowledge base of cooking techniques, flavor profiles, and ingredient interactions.
Be specific about why substitutions work or fail. Explain timing and technique constraints.
Assume a home kitchen with standard equipment unless stated otherwise.`,
		contextParts: []string{},
	}
}

// AddIngredientContext adds flavor profiles for ingredients in a recipe
func (pb *PromptBuilder) AddIngredientContext(profiles map[string]*models.IngredientProfile) *PromptBuilder {
	if len(profiles) == 0 {
		return pb
	}

	var sb strings.Builder
	sb.WriteString("\n## Ingredient Reference:\n")

	for name, profile := range profiles {
		sb.WriteString(fmt.Sprintf("### %s\n", name))
		sb.WriteString(fmt.Sprintf("Flavor: %s\n", profile.FlavorProfile))
		if len(profile.ClassicPairings) > 0 {
			sb.WriteString("Pairings: ")
			for i, pairing := range profile.ClassicPairings {
				if i > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(pairing.Name)
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	pb.contextParts = append(pb.contextParts, sb.String())
	return pb
}

// AddTechniqueContext adds cooking method details
func (pb *PromptBuilder) AddTechniqueContext(technique *models.Technique) *PromptBuilder {
	if technique == nil {
		return pb
	}

	context := fmt.Sprintf(`
## Cooking Method: %s
Temperature: %s
Timing: %s
Doneness: %s
Common Mistakes: %v
Notes: %s
`, technique.Name, technique.OilTemp, technique.Timing, technique.DonenessSign, 
      technique.CommonMistakes, technique.Notes)

	pb.contextParts = append(pb.contextParts, context)
	return pb
}

// Build constructs the final prompt
func (pb *PromptBuilder) Build(userQuery string) string {
	var sb strings.Builder
	sb.WriteString(pb.systemPrompt)
	sb.WriteString("\n")
	for _, part := range pb.contextParts {
		sb.WriteString(part)
	}
	sb.WriteString("\n\n## User Query:\n")
	sb.WriteString(userQuery)
	return sb.String()
}
```

### 2.6 Critique Handler

```go
// internal/handlers/critique_handler.go

package handlers

import (
	"encoding/json"
	"net/http"

	"recipe-backend/internal/indexer"
	"recipe-backend/internal/ollama"
	"recipe-backend/internal/prompt"
)

type CritiqueHandler struct {
	index  *indexer.KnowledgeIndex
	ollama *ollama.OllamaClient
}

func NewCritiqueHandler(index *indexer.KnowledgeIndex, client *ollama.OllamaClient) *CritiqueHandler {
	return &CritiqueHandler{index: index, ollama: client}
}

// POST /api/critique
// Body: {"recipe": "...", "focus": "prep_timing|substitutions|technique|general"}
func (h *CritiqueHandler) CritiqueRecipe(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Recipe string `json:"recipe"`
		Focus  string `json:"focus"` // optional
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Extract ingredients (simple regex-based or more sophisticated parsing)
	ingredients := extractIngredients(req.Recipe)

	// Retrieve context from knowledge base
	profiles := h.index.GetIngredientBatch(ingredients)

	// Build prompt with context
	builder := prompt.NewPromptBuilder()
	builder.AddIngredientContext(profiles)

	focusedQuery := buildFocusedQuery(req.Recipe, req.Focus)
	finalPrompt := builder.Build(focusedQuery)

	// Call Ollama
	response, err := h.ollama.Generate(finalPrompt)
	if err != nil {
		http.Error(w, "ollama error", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"critique":     response,
		"ingredients":  ingredients,
		"context_used": profiles,
	})
}

func extractIngredients(recipe string) []string {
	// TODO: Implement ingredient extraction
	// Simple approach: split on "ingredient:" lines
	// More sophisticated: regex or NLP
	return []string{}
}

func buildFocusedQuery(recipe string, focus string) string {
	base := fmt.Sprintf("Critique this recipe:\n\n%s\n\n", recipe)

	switch focus {
	case "prep_timing":
		return base + "Focus on: What can be prepped ahead? What timing constraints exist? Create a timeline for a 2-hour Sunday meal prep session."
	case "substitutions":
		return base + "Focus on: What ingredients could be substituted? Why would each substitution work or fail?"
	case "technique":
		return base + "Focus on: Are the cooking techniques appropriate? Any timing or temperature issues?"
	default:
		return base + "Provide a general critique covering technique, timing, ingredient interactions, and prep strategy."
	}
}
```

### 2.7 Main Server

```go
// cmd/server/main.go

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"recipe-backend/internal/handlers"
	"recipe-backend/internal/indexer"
	"recipe-backend/internal/loader"
	"recipe-backend/internal/ollama"
)

func main() {
	// Load configuration
	yamlDir := os.Getenv("RECIPE_YAML_DIR")
	if yamlDir == "" {
		yamlDir = os.ExpandEnv("$HOME/.recipe_context")
	}

	ollamaURL := os.Getenv("OLLAMA_URL")
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}

	ollamaModel := os.Getenv("OLLAMA_MODEL")
	if ollamaModel == "" {
		ollamaModel = "mistral"
	}

	// Load knowledge base
	kb, err := loader.LoadKnowledgeBase(yamlDir)
	if err != nil {
		log.Fatalf("Failed to load knowledge base: %v", err)
	}
	log.Println("Knowledge base loaded successfully")

	// Create index
	index := indexer.NewIndex(kb)

	// Create Ollama client
	ollamaClient := ollama.NewOllamaClient(ollamaURL, ollamaModel)

	// Setup handlers
	ingredientHandler := handlers.NewIngredientHandler(index)
	critiqueHandler := handlers.NewCritiqueHandler(index, ollamaClient)

	// Routes
	http.HandleFunc("/api/ingredient/", ingredientHandler.GetIngredient)
	http.HandleFunc("/api/ingredients/batch", ingredientHandler.GetBatch)
	http.HandleFunc("/api/critique", critiqueHandler.CritiqueRecipe)

	// Health check
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
```

## Phase 3: Fabric Integration

### 3.1 Option A: Call Backend for Context, Then Send to Ollama

Rather than letting Fabric directly call Ollama, Fabric calls the Go backend to retrieve context, then passes it through a local Ollama call.

```bash
# ~/.local/share/fabric-ai/patterns/recipe-critique/system.md
You are a recipe critic with access to a curated cooking knowledge base.
Critique recipes with specificity about technique, timing, and flavor interactions.

# ~/.local/share/fabric-ai/patterns/recipe-critique/user.md
Critique this recipe:

$1
```

```bash
# Helper script: ~/.local/bin/fabric-recipe-critique
#!/bin/bash

RECIPE="$1"
BACKEND="http://localhost:8080"

# Call backend to get ingredient context
CONTEXT=$(curl -s -X POST "$BACKEND/api/critique" \
  -H "Content-Type: application/json" \
  -d "{\"recipe\": $(echo "$RECIPE" | jq -Rs .), \"focus\": \"general\"}")

# The Go backend calls Ollama internally and returns the critique
echo "$CONTEXT" | jq -r '.critique'
```

Usage:
```bash
cat my_recipe.txt | fabric-recipe-critique
```

### 3.2 Option B: Fabric Calls Backend Directly (Simpler)

```bash
# Fabric pattern that POSTs to your Go backend

#!/bin/bash
RECIPE=$(cat)

curl -s -X POST "http://localhost:8080/api/critique" \
  -H "Content-Type: application/json" \
  -d "{\"recipe\": $(echo "$RECIPE" | jq -Rs .), \"focus\": \"prep_timing\"}" | \
  jq -r '.critique'
```

### 3.3 Option C: Skip Fabric Entirely, Use CLI

For simplicity, you might just build a small CLI wrapper:

```bash
# ~/.local/bin/recipe-critique
#!/bin/bash

RECIPE="${1:-$(cat)}"
FOCUS="${2:-general}"

curl -s -X POST "http://localhost:8080/api/critique" \
  -H "Content-Type: application/json" \
  -d "{\"recipe\": $(echo "$RECIPE" | jq -Rs .), \"focus\": \"$FOCUS\"}"
```

Usage:
```bash
cat recipe.txt | recipe-critique
recipe-critique "$(cat recipe.txt)" "prep_timing"
```

## Phase 4: Deployment & Operations

### 4.1 Running the Server

```bash
# Clone/build
git clone <repo> recipe-backend
cd recipe-backend
go build -o recipe-server ./cmd/server

# Start with environment variables
RECIPE_YAML_DIR=$HOME/.recipe_context \
OLLAMA_URL=http://localhost:11434 \
OLLAMA_MODEL=mistral \
PORT=8080 \
./recipe-server
```

### 4.2 Systemd Service (Optional)

```ini
# /etc/systemd/user/recipe-backend.service

[Unit]
Description=Recipe Backend Server
After=network.target

[Service]
Type=simple
ExecStart=%h/.local/bin/recipe-server
Environment="RECIPE_YAML_DIR=%h/.recipe_context"
Environment="OLLAMA_URL=http://localhost:11434"
Environment="OLLAMA_MODEL=mistral"
Restart=on-failure

[Install]
WantedBy=default.target
```

```bash
systemctl --user enable recipe-backend
systemctl --user start recipe-backend
```

### 4.3 Logging & Monitoring

Add basic logging:

```go
// In main.go
log.SetFlags(log.LstdFlags | log.Lshortfile)
```

Monitor server health:
```bash
curl http://localhost:8080/health
```

## Phase 5: Scaling Options

If you want to expand beyond local Ollama:

**Option A: Support multiple models**
- Add `?model=mistral` query param to select model
- Load balance across multiple Ollama instances

**Option B: Caching frequent queries**
- Cache ingredient profile lookups (rarely change)
- Cache Ollama responses (same recipe critique shouldn't re-run)

**Option C: Database backend**
- Replace YAML with SQLite for faster queries on large knowledge bases
- Add full-text search for ingredients/techniques

## Summary of Endpoints

| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/ingredient/:name` | Get profile for single ingredient |
| POST | `/api/ingredients/batch` | Get profiles for multiple ingredients |
| GET | `/api/technique/:name` | Get cooking method details |
| POST | `/api/critique` | Critique a recipe with Ollama |
| GET | `/health` | Health check |

## Expected Workflow

1. **Initialization**: Server loads YAML at startup, builds in-memory index
2. **Query**: User sends recipe via Fabric/CLI to backend
3. **Context Retrieval**: Backend extracts ingredients, looks up profiles
4. **Prompt Construction**: Backend builds prompt with context
5. **LLM Call**: Backend calls Ollama with enriched prompt
6. **Response**: Backend returns critique to user
7. **Output**: Fabric or CLI displays formatted response

This keeps the knowledge base centralized, the LLM call optimized with context, and your workflow scriptable.
