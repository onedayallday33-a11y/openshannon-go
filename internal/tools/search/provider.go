package search

import (
	"context"
)

// SearchOptions represents options for a search query
type SearchOptions struct {
	Limit int
}

// SearchResult represents a single result from a search provider
type SearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

// SearchProvider is the interface for all search engine implementations
type SearchProvider interface {
	Name() string
	Search(ctx context.Context, query string, opts SearchOptions) ([]SearchResult, error)
}
