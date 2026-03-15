package mcpserver

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerPrompts(s *mcp.Server) {
	s.AddPrompt(&mcp.Prompt{
		Name:        "inventory_summary",
		Description: "Full inventory summary: all collections, container counts, and object totals",
	}, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		return &mcp.GetPromptResult{
			Description: "Inventory summary report",
			Messages: []*mcp.PromptMessage{{
				Role: "user",
				Content: &mcp.TextContent{Text: `Please generate a full inventory summary:
1. Read all collections (nishiki://collections)
2. For each collection, read its containers (nishiki://collections/{id}/containers)
3. For each collection, read its objects (nishiki://collections/{id}/objects)
4. Summarize:
   - Total collections and their types (food, books, games, etc.)
   - Total containers across all collections
   - Total objects across all collections
   - Collections approaching capacity (if capacity is set)
   - Any objects expiring within the next 30 days
5. Highlight anything that needs attention (near-capacity containers, expiring items)`},
			}},
		}, nil
	})

	s.AddPrompt(&mcp.Prompt{
		Name:        "add_receipt",
		Description: "Parse receipt items and bulk import them into the appropriate collection",
		Arguments: []*mcp.PromptArgument{{
			Name:        "receipt_text",
			Description: "Text content of the receipt to parse and import",
			Required:    true,
		}},
	}, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		receiptText := req.Params.Arguments["receipt_text"]
		return &mcp.GetPromptResult{
			Description: "Import items from a receipt",
			Messages: []*mcp.PromptMessage{{
				Role: "user",
				Content: &mcp.TextContent{Text: fmt.Sprintf(`Parse the following receipt and import the items into the inventory:

Receipt:
%s

Steps:
1. Read all collections (nishiki://collections) to find the best target collection (likely a food collection)
2. Parse the receipt items: extract name, quantity, unit, and any relevant properties (brand, expiration date if present)
3. Format as a data array: [{"name": "...", "quantity": ..., "unit": "...", "brand": "..."}]
4. Use bulk_import to add all items to the appropriate collection
5. Report which items were imported successfully and any that failed`, receiptText)},
			}},
		}, nil
	})

	s.AddPrompt(&mcp.Prompt{
		Name:        "find_item",
		Description: "Search for an item across all collections and containers",
		Arguments: []*mcp.PromptArgument{{
			Name:        "query",
			Description: "Item name or description to search for",
			Required:    true,
		}},
	}, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		query := req.Params.Arguments["query"]
		return &mcp.GetPromptResult{
			Description: "Find item in inventory",
			Messages: []*mcp.PromptMessage{{
				Role: "user",
				Content: &mcp.TextContent{Text: fmt.Sprintf(`Search for "%s" across all inventory:

1. Read all collections (nishiki://collections)
2. For each collection, read its objects (nishiki://collections/{id}/objects)
3. Search for items matching "%s" by name, tags, or properties
4. Report:
   - Which collection and container each matching item is in
   - Item details: quantity, unit, properties, tags, expiration date
   - How many total matches were found
5. If no exact matches, suggest similar items`, query, query)},
			}},
		}, nil
	})

	s.AddPrompt(&mcp.Prompt{
		Name:        "expiration_check",
		Description: "Scan all food collections for items expiring soon",
		Arguments: []*mcp.PromptArgument{{
			Name:        "days",
			Description: "Number of days ahead to check (default: 30)",
			Required:    false,
		}},
	}, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		days := req.Params.Arguments["days"]
		if days == "" {
			days = "30"
		}
		return &mcp.GetPromptResult{
			Description: "Expiration check",
			Messages: []*mcp.PromptMessage{{
				Role: "user",
				Content: &mcp.TextContent{Text: fmt.Sprintf(`Check for items expiring within %s days:

1. Read all collections (nishiki://collections) — focus on food-type collections
2. For each food collection, read its objects (nishiki://collections/{id}/objects)
3. Find all objects with an expires_at date within the next %s days
4. Sort by expiration date (soonest first)
5. Report:
   - Items expiring within 7 days (urgent)
   - Items expiring within %s days (upcoming)
   - Collection and container location for each item
   - Suggested actions (use soon, donate, discard)`, days, days, days)},
			}},
		}, nil
	})

	s.AddPrompt(&mcp.Prompt{
		Name:        "migrate_schema",
		Description: "Review and update a collection's property schema after an import — shows inferred types, lets you correct them, then applies the updated schema",
		Arguments: []*mcp.PromptArgument{{
			Name:        "collection_id",
			Description: "ID of the collection whose schema to review",
			Required:    true,
		}},
	}, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		collectionID := req.Params.Arguments["collection_id"]
		return &mcp.GetPromptResult{
			Description: "Schema migration assistant",
			Messages: []*mcp.PromptMessage{{
				Role: "user",
				Content: &mcp.TextContent{Text: fmt.Sprintf(`Help me review and fix the property schema for collection %s.

Steps:
1. Read the collection resource (nishiki://collections/%s) to see the current property schema (definitions array).
2. Call search_objects with collection_id="%s" to sample the actual objects and their properties.
3. For each property key found in the objects, check whether a schema definition exists and whether the inferred type looks correct.
   - Common corrections needed after a CSV import:
     * Prices / amounts → currency (specify CurrencyCode, e.g. "USD")
     * Dates / timestamps → date
     * True/false columns → bool
     * URLs / links → url
     * Plain numbers → numeric
     * Fields with a small, repeating set of values → grouped_text
     * Everything else → text
4. Present a table like:

   | Key | Current Type | Sample Values | Suggested Type | Display Name |
   |-----|-------------|---------------|----------------|--------------|
   ...

5. Ask me to confirm or adjust the suggested types and display names.
6. Once confirmed, call update_collection_schema with collection_id="%s" and the corrected definitions array.
7. Report which definitions changed and confirm the schema was saved.`, collectionID, collectionID, collectionID, collectionID)},
			}},
		}, nil
	})

	s.AddPrompt(&mcp.Prompt{
		Name:        "find_by_property",
		Description: "Natural-language property search: describe what you're looking for and the assistant translates it to search_objects filters",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "description",
				Description: "Natural-language description of what to find, e.g. \"electronics for sale\" or \"red items under $20\"",
				Required:    true,
			},
			{
				Name:        "collection_id",
				Description: "Limit search to this collection ID (leave empty to search all collections)",
				Required:    false,
			},
		},
	}, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		description := req.Params.Arguments["description"]
		collectionID := req.Params.Arguments["collection_id"]

		var scopeInstructions string
		if collectionID != "" {
			scopeInstructions = fmt.Sprintf(`Search only in collection %s:
1. Read nishiki://collections/%s to see the property schema (so you know the exact property keys).
2. Translate the description into search_objects filters using those keys.
3. Call search_objects with collection_id="%s" and the derived filters.`, collectionID, collectionID, collectionID)
		} else {
			scopeInstructions = `Search across all collections:
1. Read nishiki://collections to list all collections and their schemas.
2. Identify which collections are likely to contain matching objects based on the description.
3. For each candidate collection, translate the description into search_objects filters using that collection's schema keys.
4. Call search_objects once per candidate collection and merge the results.`
		}

		return &mcp.GetPromptResult{
			Description: "Property-based search",
			Messages: []*mcp.PromptMessage{{
				Role: "user",
				Content: &mcp.TextContent{Text: fmt.Sprintf(`Find inventory items matching: "%s"

%s

Translation rules:
- Boolean concepts ("for sale", "in stock", "available") → property_filters: {"<key>": "true"}
- Negations ("not for sale") → property_filters: {"<key>": "false"}
- Value equality ("color red", "brand Nike") → property_filters: {"<key>": "<value>"}
- Name/keyword search → use the "query" field on search_objects instead of property_filters
- Tag-based ("tagged urgent") → use the "tags" field on search_objects

After collecting results:
- Show each matching item with its collection, container, and relevant properties
- If no results, explain which filters were tried and suggest alternatives`, description, scopeInstructions)},
			}},
		}, nil
	})

	s.AddPrompt(&mcp.Prompt{
		Name:        "reorganize",
		Description: "Analyze inventory layout and suggest reorganization for better utilization",
		Arguments: []*mcp.PromptArgument{{
			Name:        "collection_id",
			Description: "ID of the collection to analyze (leave empty for all collections)",
			Required:    false,
		}},
	}, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		collectionID := req.Params.Arguments["collection_id"]

		var target string
		if collectionID != "" {
			target = fmt.Sprintf("Read collection nishiki://collections/%s and its containers (nishiki://collections/%s/containers)", collectionID, collectionID)
		} else {
			target = "Read all collections (nishiki://collections) and for each, read its containers"
		}

		return &mcp.GetPromptResult{
			Description: "Inventory reorganization suggestions",
			Messages: []*mcp.PromptMessage{{
				Role: "user",
				Content: &mcp.TextContent{Text: fmt.Sprintf(`Analyze the inventory and suggest reorganization:

1. %s
2. For each container, check:
   - Current object count vs. capacity (if set)
   - Container type appropriateness for its contents
   - Object types stored vs. collection type
3. Identify:
   - Over-capacity containers (utilization > 90%%)
   - Under-utilized containers (utilization < 20%%)
   - Objects that seem misplaced (wrong container type)
   - Containers that could be merged or split
4. Suggest specific moves:
   - Which objects to move where
   - Which containers to create or remove
   - Better naming or categorization
5. Prioritize suggestions by impact`, target)},
			}},
		}, nil
	})
}
