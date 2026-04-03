package lsp

import (
	"context"
	"path/filepath"
	"sync"
)

// LspManager manages multiple LSP server connections
type LspManager struct {
	mu      sync.Mutex
	servers map[string]*LspClient // Key: extension (e.g. ".go")
	root    string
}

var (
	DefaultLspManager *LspManager
	once              sync.Once
)

func GetLspManager(root string) *LspManager {
	once.Do(func() {
		DefaultLspManager = &LspManager{
			servers: make(map[string]*LspClient),
			root:    root,
		}
	})
	return DefaultLspManager
}

// GetClientForFile returns or starts a language server for a file type
func (m *LspManager) GetClientForFile(ctx context.Context, path string) (*LspClient, error) {
	ext := filepath.Ext(path)
	if ext == "" {
		return nil, nil // No extension, no LSP
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if client, exists := m.servers[ext]; exists {
		return client, nil
	}

	// 1. Detect Command
	cmd, args := m.detectServer(ext)
	if cmd == "" {
		return nil, nil // No server found for this ext
	}

	// 2. Start Client
	client, err := NewLspClient(ctx, cmd, args, m.root)
	if err != nil {
		return nil, err
	}

	m.servers[ext] = client
	return client, nil
}

func (m *LspManager) detectServer(ext string) (string, []string) {
	switch ext {
	case ".go":
		return "gopls", nil
	case ".py":
		return "pyright-langserver", []string{"--stdio"}
	case ".ts", ".tsx", ".js", ".jsx":
		return "typescript-language-server", []string{"--stdio"}
	case ".rs":
		return "rust-analyzer", nil
	}
	return "", nil
}

func (m *LspManager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, client := range m.servers {
		client.Close()
	}
}
