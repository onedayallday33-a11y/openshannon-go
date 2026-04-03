package search

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// DuckDuckGoProvider implements SearchProvider using a simple HTML scraper (no API key needed)
type DuckDuckGoProvider struct{}

func (p *DuckDuckGoProvider) Name() string {
	return "DuckDuckGo"
}

func (p *DuckDuckGoProvider) Search(ctx context.Context, query string, opts SearchOptions) ([]SearchResult, error) {
	searchURL := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, err
	}
	
	// DDG HTML requires a real-looking User-Agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("duckduckgo error: %s", resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var results []SearchResult
	doc.Find(".links_main.links_deep.result__body").Each(func(i int, s *goquery.Selection) {
		if len(results) >= opts.Limit {
			return
		}

		title := strings.TrimSpace(s.Find(".result__title").Text())
		link, _ := s.Find(".result__a").Attr("href")
		snippet := strings.TrimSpace(s.Find(".result__snippet").Text())

		// DDG links are often wrapped /l/?kh=-1&uddg=URL
		if strings.Contains(link, "uddg=") {
			u, _ := url.Parse(link)
			link = u.Query().Get("uddg")
		}

		if title != "" && link != "" {
			results = append(results, SearchResult{
				Title:   title,
				URL:     link,
				Snippet: snippet,
			})
		}
	})

	return results, nil
}
