# Recipe Planning: Notes and Architecture

## Overview

This document covers two things:
1. How recipe/meal planning fits into the Nishiki ecosystem (primarily via Grocy)
2. A custom recipe knowledge base design (retained for the cooking technique / ingredient metadata layer that Grocy doesn't provide)

---

## Integration with Grocy

[Grocy](https://grocy.info) handles the meal planning and recipe management side of this ecosystem via its MCP server. Nishiki's role is tracking *what you have and where it is*; Grocy's role is *what to cook with it*.

### Division of responsibility

| Concern | System |
|---|---|
| Inventory (what you have, where it is) | Nishiki |
| Expiration tracking | Nishiki (expiry fields on objects) |
| Recipes and meal plans | Grocy |
| Shopping lists | Grocy |
| Recipe ↔ inventory linkage | Claude bridges via both MCP servers |

### Workflow example

```
User: "Plan meals for this week using what's about to expire"
Claude:
  1. nishiki://collections → find food collections
  2. nishiki MCP expiration_check prompt → items expiring soon
  3. Grocy MCP → find recipes that use those ingredients
  4. Grocy MCP → generate meal plan + shopping list for gaps
```

No direct API integration between Nishiki and Grocy is required — Claude orchestrates across both MCP servers.

### What Grocy does NOT cover

Grocy stores recipes but doesn't have deep ingredient knowledge (flavor profiles, technique notes, substitution reasoning). The custom knowledge base below addresses this gap if needed.

---

## Custom Recipe Knowledge Base (Optional Layer)

If deeper cooking intelligence is needed beyond what Grocy provides — particularly for ingredient substitution reasoning, flavor pairing, and cooking technique guidance — a lightweight local service can serve a curated YAML knowledge base.

### When to build this

Build this if:
- You want Claude to reason about *why* a substitution works, not just list options
- You want technique-specific advice (temps, timing, doneness signs) in context
- Grocy's recipe store doesn't satisfy the "cooking assistant" use case

Skip this if:
- Grocy + Claude's general knowledge is sufficient for your cooking workflow
- You don't want to maintain a YAML knowledge base

### Bootstrapping the Knowledge Base

Use Claude or Gemini to extract structured cooking principles from trusted sources (cookbooks, Serious Eats, etc.) into YAML files.

#### Phase 1: Define scope

Focus on:
- 30–50 core ingredients you actually cook with
- High-value ingredients with clear substitution patterns
- Techniques used across multiple recipes (not one-offs)

Sources to process:
- *Salt Fat Acid Heat* — principles and technique
- *The Flavor Bible* — pairing reference
- *Serious Eats* articles — technique depth
- Your own recipe collection

#### Phase 2: Extract with AI

**Prompt template for ingredient profiles:**

```
I'm building a recipe knowledge base. Here are recipes from [COOKBOOK]:

[PASTE RECIPES]

For each of these ingredients: [LIST], extract a structured profile with:
- flavor_profile: dominant tastes and aromas
- texture: raw and cooked texture
- absorbs_flavors: boolean
- classic_pairings: list with reasoning
- cooking_methods: per method — temp, time, notes
- prep_notes: common prep steps
- substitutions: what works, what doesn't, and why

Output as YAML.
```

**Prompt template for cooking techniques:**

```
For the technique [TECHNIQUE NAME], extract:
- equipment needed
- oil/fat type
- prep requirements
- heat level and target temperature
- timing guidelines
- doneness signs
- common mistakes
- notes

Output as YAML.
```

#### Phase 3: Validate

Review AI output against source material. Check:
- Substitution compatibility claims
- Temperature and timing accuracy
- Pairing reasoning

### Knowledge Base Structure

```
~/.recipe_context/
├── flavor_profiles.yaml    # Per-ingredient profiles
├── substitutions.yaml      # Cross-ingredient substitution matrix
├── techniques.yaml         # Cooking method reference
└── interactions.yaml       # Ingredient interaction patterns
```

**Example: `flavor_profiles.yaml`**

```yaml
eggplant:
  flavor_profile: "mild, earthy, slightly bitter when raw"
  texture: "spongy raw, creamy when roasted"
  absorbs_flavors: true
  classic_pairings:
    - name: tomato
      reasoning: "acid cuts bitterness; both are Mediterranean staples"
    - name: garlic
      reasoning: "aromatic base that eggplant absorbs well"
  cooking_methods:
    roasting:
      temp: "425°F / 220°C"
      time: "25–35 min"
      notes: "salt and drain first to reduce bitterness; high heat caramelizes sugars"
    sautéing:
      temp: "medium-high"
      time: "8–12 min"
      notes: "use generous oil; eggplant absorbs quickly"
  substitutions:
    zucchini:
      compatibility: "moderate"
      adjustments: "reduce cook time by ~30%; less oil needed"
      notes: "less bitter, more watery; use for mild dishes"
```

**Example: `techniques.yaml`**

```yaml
pan_searing:
  equipment: "heavy skillet (cast iron preferred)"
  oil_type: "high smoke point: avocado, grapeseed, or refined coconut"
  prep:
    - "pat protein completely dry"
    - "bring to room temperature (20–30 min)"
    - "season immediately before cooking"
  heat_level: "high"
  oil_temp: "shimmering, just before smoke point (~375–400°F)"
  timing: "3–5 min per side for 1-inch protein; adjust for thickness"
  doneness_signs: "golden-brown crust releases naturally from pan"
  common_mistakes:
    - "moving protein before crust forms"
    - "overcrowding pan (steams instead of sears)"
    - "cold pan or low heat"
  notes: "rest 5 min before cutting; carry-over cooking continues"
```

### Service Architecture (if building)

A lightweight Go HTTP server loads the YAML at startup and serves query endpoints. The MCP server or Claude can call it for context when answering cooking questions.

```
~/.recipe_context/ (YAML files)
        ↓
Go HTTP server (port 8080)
  GET /api/ingredient/:name
  POST /api/ingredients/batch
  GET /api/technique/:name
  POST /api/critique          → calls local Ollama with context
        ↓
Local Ollama (Mistral / Llama)
```

**Key endpoints:**

| Method | Endpoint | Purpose |
|---|---|---|
| GET | `/api/ingredient/:name` | Single ingredient profile |
| POST | `/api/ingredients/batch` | Multiple ingredient profiles |
| GET | `/api/technique/:name` | Cooking method details |
| POST | `/api/critique` | Recipe critique with Ollama |
| GET | `/health` | Health check |

**Typical flow:**
1. User sends recipe to `/api/critique`
2. Server extracts ingredients, fetches their profiles from the index
3. Builds a context-enriched prompt
4. Calls Ollama with the enriched prompt
5. Returns the critique

### Running the Service

```bash
# Environment
RECIPE_YAML_DIR=$HOME/.recipe_context \
OLLAMA_URL=http://localhost:11434 \
OLLAMA_MODEL=mistral \
PORT=8080 \
./recipe-server
```

Or as a systemd user service:

```ini
[Unit]
Description=Recipe Knowledge Base Server
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

### Integration with Nishiki MCP

The recipe knowledge server is a separate concern from Nishiki. Claude can call both:

```
User: "I have eggplant and tomatoes expiring — what should I make?"

Claude:
  1. nishiki MCP → confirm quantities available
  2. Grocy MCP → find recipes with eggplant + tomato
  3. Recipe KB (optional) → get flavor pairing and technique notes
  4. Synthesize: recommend recipe, explain why it works, provide technique guidance
```

---

## Decision Summary

| Use case | Use |
|---|---|
| Meal planning, shopping lists | Grocy MCP |
| Recipe storage | Grocy MCP |
| Inventory → recipe ingredient matching | Claude bridges Nishiki + Grocy MCPs |
| Cooking technique / substitution reasoning | Recipe knowledge base (optional) |
| General cooking questions | Claude's built-in knowledge (usually sufficient) |
