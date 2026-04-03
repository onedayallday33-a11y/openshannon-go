package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SerperProvider implements SearchProvider using Serper.dev (Google Search API)
type SerperProvider struct {
	APIKey string
}

func (p *SerperProvider) Name() string {
	return "Serper"
}

type serperRequest struct {
	Q   string `json:"q"`
	Num int    `json:"num,omitempty"`
}

type serperResponse struct {
	Organic []struct {
		Title   string `json:"title"`
		Link    string `json:"link"`
		Snippet string `json:"snippet"`
	} `json:"organic"`
}

func (p *SerperProvider) Search(ctx context.Context, query string, opts SearchOptions) ([]SearchResult, error) {
	if p.APIKey == "" {
		return nil, fmt.Errorf("Serper API key is missing")
	}

	reqBody, _ := json.Marshal(serperRequest{
		Q:   query,
		Num: opts.Limit,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", "https://google.serper.dev/search", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-KEY", p.APIKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("serper error: %s", resp.Status)
	}

	var data serperResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	results := make([]SearchResult, len(data.Organic))
	for i, r := range data.Organic {
		results[i] = SearchResult{
			Title:   r.Title,
			URL:     r.Link,
			Snippet: r.Snippet,
		}
	}

	return results, nil
}
