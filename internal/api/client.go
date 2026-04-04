package api

import (
    "bytes"
    "context"
    "fmt"
    "io"
    "net/http"

    "github.com/onedayallday33-a11y/openshannon-go/internal/config"
    jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Client represents the OpenAI capable HTTP client
type Client struct {
    HTTPClient *http.Client
}

// NewClient init
func NewClient() *Client {
    return &Client{
        HTTPClient: &http.Client{},
    }
}

// DoRequest performs the cross translation network call
func (c *Client) DoRequest(ctx context.Context, req *AnthropicMessageRequest) (*http.Response, error) {
    if !config.UseOpenAI() {
        return nil, fmt.Errorf("CLAUDE_CODE_USE_OPENAI is not enabled")
    }

    // fallback to env model if none set
    if req.Model == "" {
        req.Model = config.OpenAIModel()
    }

    openaiReq := ConvertRequest(req)
    reqBody, err := json.Marshal(openaiReq)
    if err != nil {
        return nil, err
    }

    url := config.OpenAIBaseURL() + "/chat/completions"
    httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
    if err != nil {
        return nil, err
    }

    httpReq.Header.Set("Content-Type", "application/json")
    
    if key := config.OpenAIApiKey(); key != "" {
        httpReq.Header.Set("Authorization", "Bearer "+key)
    }

    resp, err := c.HTTPClient.Do(httpReq)
    if err != nil {
        return nil, err
    }

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        resp.Body.Close()
        return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
    }

    return resp, nil
}
