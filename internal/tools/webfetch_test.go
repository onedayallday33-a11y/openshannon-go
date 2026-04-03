package tools

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebFetchTool_Execute(t *testing.T) {
	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<html><body><h1>Hello World</h1><p>This is a test.</p></body></html>"))
	}))
	defer server.Close()

	tool := &WebFetchTool{}
	ctx := context.Background()

	t.Run("Fetch and convert to Markdown", func(t *testing.T) {
		args := map[string]interface{}{
			"url": server.URL,
		}
		
		result, err := tool.Execute(ctx, args)
		require.NoError(t, err)
		
		resMap := result.(map[string]interface{})
		assert.Equal(t, server.URL, resMap["url"])
		assert.Contains(t, resMap["content"].(string), "# Hello World")
		assert.Contains(t, resMap["content"].(string), "This is a test.")
	})

	t.Run("Invalid URL", func(t *testing.T) {
		args := map[string]interface{}{
			"url": "ftp://invalid-url",
		}
		_, err := tool.Execute(ctx, args)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid URL")
	})
}
