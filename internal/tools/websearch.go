package tools

import (
	"context"
	"fmt"
	"os"

	"github.com/onedayallday33-a11y/openshannon-go/internal/tools/search"
)

// WebSearchTool implements the Tool interface for searching the web
type WebSearchTool struct{}

// Name of the tool
func (t *WebSearchTool) Name() string {
	return "WebSearch"
}

// Description of the tool
func (t *WebSearchTool) Description() string {
	return "Search the web for information"
}

// InputSchema for the tool
func (t *WebSearchTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "The search query",
			},
			"limit": map[string]interface{}{
				"type":        "integer",
				"description": "Number of results to return (default 5)",
			},
		},
		"required": []string{"query"},
	}
}

// Execute the web search logic
func (t *WebSearchTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	query, ok := args["query"].(string)
	if !ok {
		return nil, fmt.Errorf("query is required")
	}

	limit := 5
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	} else if l, ok := args["limit"].(int); ok {
		limit = l
	}

	// 1. Provider Selection (Registry Logic)
	var provider search.SearchProvider

	if key := os.Getenv("TAVILY_API_KEY"); key != "" {
		provider = &search.TavilyProvider{APIKey: key}
	} else if key := os.Getenv("SERPER_API_KEY"); key != "" {
		provider = &search.SerperProvider{APIKey: key}
	} else {
		// Fallback to DuckDuckGo (Scraper)
		provider = &search.DuckDuckGoProvider{}
	}

	// 2. Perform Search
	results, err := provider.Search(ctx, query, search.SearchOptions{Limit: limit})
	if err != nil {
		return nil, fmt.Errorf("search failed using %s: %v", provider.Name(), err)
	}

	return map[string]interface{}{
		"query":    query,
		"provider": provider.Name(),
		"results":  results,
	}, nil
}
