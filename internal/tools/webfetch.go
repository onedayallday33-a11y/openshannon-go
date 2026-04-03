package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/JohannesKaufmann/html-to-markdown/v2"
)

const (
	// MaxFetchSize (500KB text)
	MaxFetchSize = 500 * 1024
	// DefaultFetchTimeout
	DefaultFetchTimeout = 30 * time.Second
)

// WebFetchTool implements the Tool interface for fetching web content
type WebFetchTool struct{}

// Name of the tool
func (t *WebFetchTool) Name() string {
	return "WebFetch"
}

// Description of the tool
func (t *WebFetchTool) Description() string {
	return "Fetch content from a URL and convert to Markdown"
}

// InputSchema for the tool
func (t *WebFetchTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"url": map[string]interface{}{
				"type":        "string",
				"description": "The URL to fetch content from",
			},
		},
		"required": []string{"url"},
	}
}

// Execute the web fetch logic
func (t *WebFetchTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	url, ok := args["url"].(string)
	if !ok {
		return nil, fmt.Errorf("url is required")
	}

	// 1. Basic URL Validation
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return nil, fmt.Errorf("invalid URL: must start with http:// or https://")
	}

	// 2. Fetch with Timeout
	client := &http.Client{
		Timeout: DefaultFetchTimeout,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	
	// Add common User-Agent to avoid simple bot blocking
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 OpenShannon/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http error: %s", resp.Status)
	}

	// 3. Read Body (with limit)
	limitReader := io.LimitReader(resp.Body, MaxFetchSize+1)
	body, err := io.ReadAll(limitReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %v", err)
	}

	if len(body) > MaxFetchSize {
		// Truncated
		body = body[:MaxFetchSize]
	}

	contentType := resp.Header.Get("Content-Type")

	// 4. Conversion (only if HTML)
	result := string(body)
	if strings.Contains(contentType, "text/html") {
		markdown, err := htmltomarkdown.ConvertString(result)
		if err == nil {
			result = markdown
		}
	}

	return map[string]interface{}{
		"url":         url,
		"content":     result,
		"contentType": contentType,
		"truncated":   len(body) >= MaxFetchSize,
	}, nil
}
