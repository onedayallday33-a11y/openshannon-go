package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// McpClient manages a single connection to an MCP server
type McpClient struct {
	Name      string
	Transport *StdioTransport
	mu        sync.Mutex
	lastID    int
	pending   map[int]chan *JsonRpcResponse
}

// NewMcpClient creates and initializes a new client
func NewMcpClient(ctx context.Context, name string, config McpServerConfig) (*McpClient, error) {
	transport, err := NewStdioTransport(ctx, config.Command, config.Args, config.Env)
	if err != nil {
		return nil, err
	}

	client := &McpClient{
		Name:      name,
		Transport: transport,
		pending:   make(map[int]chan *JsonRpcResponse),
	}

	// Start response listener loop
	go client.listen()

	// Perform Handshake
	if err := client.Initialize(ctx); err != nil {
		client.Close()
		return nil, fmt.Errorf("MCP handshake failed: %v", err)
	}

	return client, nil
}

func (c *McpClient) listen() {
	for {
		data, err := c.Transport.Receive()
		if err != nil {
			return
		}

		var resp JsonRpcResponse
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

// Call sends a JSON-RPC request and waits for response
func (c *McpClient) Call(ctx context.Context, method string, params interface{}) (*JsonRpcResponse, error) {
	c.mu.Lock()
	c.lastID++
	id := c.lastID
	ch := make(chan *JsonRpcResponse, 1)
	c.pending[id] = ch
	c.mu.Unlock()

	req := JsonRpcRequest{
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
			return nil, fmt.Errorf("MCP error [%d]: %s", resp.Error.Code, resp.Error.Message)
		}
		return resp, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("MCP request timeout")
	}
}

// Initialize performs the mandatory handshake
func (c *McpClient) Initialize(ctx context.Context) error {
	params := InitializeParams{
		ProtocolVersion: "2024-11-05", // Spec version
		ClientInfo: ClientInfo{
			Name:    "OpenShannon-Go",
			Version: "1.0.0",
		},
		Capabilities: make(map[string]interface{}),
	}

	_, err := c.Call(ctx, "initialize", params)
	return err
}

// ListTools returns the tools provided by the server
func (c *McpClient) ListTools(ctx context.Context) ([]McpTool, error) {
	resp, err := c.Call(ctx, "tools/list", nil)
	if err != nil {
		return nil, err
	}

	var result ListToolsResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, err
	}
	return result.Tools, nil
}

// CallTool executes a specific tool on the server
func (c *McpClient) CallTool(ctx context.Context, name string, args map[string]interface{}) (json.RawMessage, error) {
	params := CallToolParams{
		Name:      name,
		Arguments: args,
	}
	resp, err := c.Call(ctx, "tools/call", params)
	if err != nil {
		return nil, err
	}
	return resp.Result, nil
}

func (c *McpClient) Close() {
	if c.Transport != nil {
		c.Transport.Close()
	}
}
