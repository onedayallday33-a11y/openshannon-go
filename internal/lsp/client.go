package lsp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/onedayallday33-a11y/openshannon-go/internal/mcp"
)

// LspClient manages a single connection to an LSP server
type LspClient struct {
	Transport *mcp.StdioTransport
	mu        sync.Mutex
	lastID    int
	pending   map[int]chan *LspResponse
}

// NewLspClient creates and initializes a new client
func NewLspClient(ctx context.Context, command string, args []string, rootPath string) (*LspClient, error) {
	transport, err := mcp.NewStdioTransport(ctx, command, args, nil)
	if err != nil {
		return nil, err
	}

	client := &LspClient{
		Transport: transport,
		pending:   make(map[int]chan *LspResponse),
	}

	go client.listen()

	// Perform Handshake
	if err := client.Initialize(ctx, rootPath); err != nil {
		client.Close()
		return nil, fmt.Errorf("LSP initialize failed: %v", err)
	}

	return client, nil
}

func (c *LspClient) listen() {
	for {
		data, err := c.Transport.Receive()
		if err != nil {
			return
		}

		var resp LspResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			continue
		}

		if id, ok := resp.ID.(float64); ok {
			c.mu.Lock()
			if ch, exists := c.pending[int(id)]; exists {
				ch <- &resp
				delete(c.pending, int(id))
			}
			c.mu.Unlock()
		}
	}
}

// Call sends an LSP request and waits for response
func (c *LspClient) Call(ctx context.Context, method string, params interface{}) (*LspResponse, error) {
	c.mu.Lock()
	c.lastID++
	id := c.lastID
	ch := make(chan *LspResponse, 1)
	c.pending[id] = ch
	c.mu.Unlock()

	req := LspRequest{
		Jsonrpc: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	if err := c.Transport.Send(req); err != nil {
		return nil, err
	}

	select {
	case resp := <-ch:
		if resp.Error != nil {
			return nil, fmt.Errorf("LSP error [%d]: %s", resp.Error.Code, resp.Error.Message)
		}
		return resp, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("LSP request timeout")
	}
}

// Initialize performs the mandatory handshake
func (c *LspClient) Initialize(ctx context.Context, rootPath string) error {
	params := InitializeParams{
		ProcessID: os.Getpid(),
		RootPath:  rootPath,
	}

	_, err := c.Call(ctx, "initialize", params)
	return err
}

func (c *LspClient) Close() {
	if c.Transport != nil {
		c.Transport.Close()
	}
}
