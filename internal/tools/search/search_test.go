package search

import (
	"testing"
)

func TestSearchInterface(t *testing.T) {
	var _ SearchProvider = &TavilyProvider{}
	var _ SearchProvider = &SerperProvider{}
	var _ SearchProvider = &DuckDuckGoProvider{}
}
