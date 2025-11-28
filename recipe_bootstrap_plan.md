# Plan 1: Bootstrapping Recipe Knowledge Base with Claude/Gemini

## Overview
Use Claude or Gemini to extract and structure recipe principles from your favorite cookbooks into validated YAML files. This is a one-time process that creates your reference base.

## Phase 1: Preparation

### 1.1 Identify Source Material
- List 2-3 cookbooks or recipe collections you trust
- Examples: "Salt Fat Acid Heat," "The Flavor Bible," Serious Eats recipes, Mediterranean cookbook, etc.
- For each, identify:
  - Recipe count
  - Cuisine focus
  - Ingredient range (narrow vs. broad)
  - Substitution patterns the author uses

### 1.2 Define Scope
Rather than extracting *all* ingredients, focus on:
- **Core ingredients** you actually cook with regularly (30-50 to start)
- **High-value ingredients** that appear in multiple recipes and have clear substitution patterns
- **Techniques** used across multiple recipes (not one-off methods)

## Phase 2: Extraction Prompts

### 2.1 Ingredient Profile Extraction

**Prompt template for Claude/Gemini:**

```
I'm building a recipe knowledge base. Here are recipes from [COOKBOOK NAME]:

[PASTE 3-5 RECIPES USING EGGPLANT]

For the ingredient "eggplant," extract:
1. Flavor profile (earthiness, bitterness, how it absorbs flavors, etc.)
2. Classic pairings (list at least 5, with reasoning if the source gives it)
3. Cooking methods used (roasting, frying, grilling, braising, etc.)
4. Timing for each method (e.g., "roasting at 425°F: 25-30 min")
5. Prep notes (salting, moisture management, etc.)
6. Any substitutions the recipes mention or imply

Format as YAML:
```yaml
eggplant:
  flavor_profile: "..."
  texture: "..."
  absorbs_flavors: true
  classic_pairings:
    - name: "tomato"
      reasoning: "..."
    - name: "garlic"
      reasoning: "..."
  cooking_methods:
    roasting:
      temp: "425°F"
      time: "25-30 min"
      notes: "..."
    frying:
      temp: "350-375°F"
      time: "3-4 min per side"
      notes: "..."
  prep_notes:
    - "salt 30 min before cooking to draw out moisture"
    - "pat dry thoroughly"
  substitutions:
    zucchini:
      compatibility: "high"
      adjustments: "reduce cooking time by 20%"
      notes: "lighter flavor, similar texture"
```

Do this for one ingredient at a time to avoid errors.
```

**Workflow:**
1. Create a shared doc with 5-10 recipes from your cookbook
2. Ask Claude to extract one ingredient (eggplant, tomato, garlic, etc.)
3. Review the output for accuracy against the recipes
4. Iterate: "The recipes don't mention salting—remove that. Add the olive oil pairing from recipe 3."
5. Once validated, add to `flavor_profiles.yaml`
6. Repeat for next ingredient

**Estimated effort:** 2-3 minutes per ingredient after you've validated the prompt format.

### 2.2 Substitution Guidelines Extraction

**Prompt template:**

```
I'm creating substitution guidelines. Based on these recipes from [COOKBOOK]:

[PASTE RECIPES]

Extract all substitutions mentioned or implied:
- Original ingredient
- Substitute
- Ratio/adjustment needed
- Why it works
- When it doesn't work
- Flavor/texture impact

Format as:
```yaml
substitutions:
  butter_to_oil:
    ratio: "1:1 by volume"
    adjustments:
      - "oil won't cream; use different mixing method"
      - "slightly less oil (0.9x) if very liquid-heavy recipe"
    works_for: ["baking", "pan-frying", "sautéing"]
    avoid_in: ["creaming recipes", "laminated dough"]
    flavor_impact: "more grassy/fruity if olive oil; neutral if vegetable"
  greek_yogurt_for_sour_cream:
    ratio: "1:1 by volume"
    adjustments:
      - "reduce other liquids by 10%"
      - "reduce salt slightly (greek yogurt is tangier)"
    works_for: ["baking", "cold sauces", "toppings"]
    avoid_in: ["high-heat sauces", "delicate emulsions"]
    flavor_impact: "tangier, higher protein, thicker"
```

Go through the recipes and identify 5-10 substitutions with reasoning.
```

**Workflow:**
1. Paste recipes where you notice ingredient swaps or author-suggested alternatives
2. Claude extracts with reasoning
3. Validate: "Did the author actually suggest this? Is the reasoning sound?"
4. Add to `substitutions.yaml`

### 2.3 Technique Guidelines Extraction

**Prompt template:**

```
I'm documenting cooking techniques. From these recipes in [COOKBOOK]:

[PASTE RECIPES USING FRYING]

Extract technique "frying eggplant":
- Required equipment (skillet size, oil type)
- Preparation steps (patting dry, salting, etc.)
- Temperature/heat level
- Duration
- Signs of doneness
- Common mistakes from the recipes
- Notes on why this method works

Format as:
```yaml
techniques:
  frying_eggplant:
    equipment: "cast iron or heavy-bottomed skillet"
    oil_type: "neutral oil (vegetable, canola) or olive oil"
    prep:
      - "salt eggplant 30 min before cooking"
      - "pat completely dry with paper towels"
      - "cut into 1/4-inch rounds or 1/2-inch batons"
    heat_level: "medium-high"
    oil_temp: "350-375°F"
    timing: "3-4 min per side"
    doneness_signs: "golden brown, tender when pierced"
    common_mistakes:
      - "not patting dry—results in steaming instead of frying"
      - "oil too cool—absorbs oil instead of crisping"
    notes: "moisture is critical; the salting/drying step prevents wateriness"
```
```

**Workflow:**
1. For each technique that appears in 2+ recipes, extract it
2. Validate against actual recipe instructions
3. Add to `techniques.yaml`

### 2.4 Ingredient Interactions Extraction

**Prompt template:**

```
I'm documenting chemical/flavor interactions. From these recipes:

[PASTE RECIPES WHERE ACID MEETS FAT/DAIRY, ETC.]

Extract interaction patterns:
- What happens when X meets Y
- Why it happens
- How to prevent/manage it
- Example from the recipes

Format as:
```yaml
interactions:
  acid_dairy:
    - trigger: "high acid (vinegar, citrus, tomato) + dairy (cream, milk)"
      effect: "can curdle if not managed"
      prevention:
        - "add acid slowly to dairy, not vice versa"
        - "use gentle heat"
        - "temper with a small amount of dairy first"
        - "starch can help stabilize"
      example: "Caesar dressing: egg yolk + lemon juice, whisked while slowly drizzling oil"
  acid_salt:
    - trigger: "high salt + high acid"
      effect: "can overpower if not balanced"
      adjustment: "add small amounts, taste frequently"
      example: "pickles: vinegar and salt work together, but either one alone tastes harsh"
```
```

**Workflow:**
1. Scan recipes for ingredient combinations that "just work" or have warnings
2. Ask Claude to extract the science/principle
3. Add to `interactions.yaml`

## Phase 3: Compilation & Validation

### 3.1 Organize Files

```
~/.recipe_context/
├── flavor_profiles.yaml       # Ingredients with pairings, methods, prep
├── substitutions.yaml         # Ingredient swap rules
├── techniques.yaml            # Cooking methods with timing/temps
├── interactions.yaml          # Chemistry/flavor rules
├── cuisines.yaml              # Flavor combinations by cuisine (optional)
└── README.md                  # Notes on sources, last updated
```

### 3.2 Cross-Reference Validation

After extracting all ingredients, do a pass to check:
- **Consistency**: Does "frying eggplant" timing match across all recipes?
- **Gaps**: Do substitutions reference ingredients in `flavor_profiles.yaml`?
- **Coverage**: Can you reconstruct a recipe using only your YAML files? (You should be able to roughly)

**Claude validation prompt:**

```
I've created these YAML files for recipe guidance. Check them for:
1. Inconsistencies (e.g., eggplant frying time varies wildly)
2. Missing context (substitutions reference undefined ingredients)
3. Outdated info (techniques that don't match modern source books)

[PASTE YAML SECTIONS]

Flag any issues and suggest fixes.
```

### 3.3 Versioning

Add metadata to each file:

```yaml
# flavor_profiles.yaml
_metadata:
  version: "1.0"
  sources:
    - "Salt Fat Acid Heat by Samin Nosrat"
    - "The Flavor Bible by Karen Page"
  last_updated: "2025-01-15"
  coverage: "50 ingredients"
  note: "Focused on Mediterranean and vegetable-forward cooking"
```

## Phase 4: Iteration & Growth

### 4.1 Spot-Check Against New Recipes

When you cook a new recipe:
1. Check if your YAML would generate similar guidance
2. If guidance is off, update the YAML
3. If new ingredients, extract them

### 4.2 Seasonal Updates

Every few months:
- Add new ingredients you're using
- Refine techniques based on real cooking experience
- Remove or update info that doesn't align with how you actually cook

### 4.3 Source Expansion

If you want to add a new cookbook:
1. Extract 3-5 recipes as a test batch
2. Generate YAML using Claude
3. Cross-check with existing entries—does new cookbook agree or contradict?
4. Merge or note differences (useful for understanding alternative approaches)

## Expected Output

After this phase:
- **flavor_profiles.yaml**: 40-50 ingredients with pairings, methods, prep
- **substitutions.yaml**: 20-30 substitution rules with reasoning
- **techniques.yaml**: 15-20 cooking methods with timing/temps/notes
- **interactions.yaml**: 10-15 chemical/flavor interaction patterns
- Total: ~100-150KB of structured reference

This becomes your static knowledge base for the backend to query.
