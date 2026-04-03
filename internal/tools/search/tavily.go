package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// TavilyProvider implements SearchProvider using Tavily AI Search API
type TavilyProvider struct {
	APIKey string
}

func (p *TavilyProvider) Name() string {
	return "Tavily"
}

type tavilyRequest struct {
	Query      string `json:"query"`
	MaxResults int    `json:"max_results,omitempty"`
}

type tavilyResponse struct {
	Results []struct {
		Title   string `json:"title"`
		URL     string `json:"url"`
		Content string `json:"content"`
	} `json:"results"`
}

func (p *TavilyProvider) Search(ctx context.Context, query string, opts SearchOptions) ([]SearchResult, error) {
	if p.APIKey == "" {
		return nil, fmt.Errorf("Tavily API key is missing")
	}

	reqBody, _ := json.Marshal(tavilyRequest{
		Query:      query,
		MaxResults: opts.Limit,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.tavily.com/search", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.APIKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tavily error: %s", resp.Status)
	}

	var data tavilyResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	results := make([]SearchResult, len(data.Results))
	for i, r := range data.Results {
		results[i] = SearchResult{
			Title:   r.Title,
			URL:     r.URL,
			Snippet: r.Content,
		}
	}

	return results, nil
}
