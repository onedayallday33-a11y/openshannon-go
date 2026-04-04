package permissions

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsReadAllowed(t *testing.T) {
	cwd, _ := os.Getwd()

	tests := []struct {
		name    string
		path    string
		allowed bool
	}{
		{"Normal file", "README.md", true},
		{"Sensitive file .bashrc", ".bashrc", false},
		{"Sensitive file .gitconfig", ".gitconfig", false},
		{"Inside .git directory", filepath.Join(".git", "config"), false},
		{"NTFS Stream", "file.txt::$DATA", false},
		{"8.3 Short name", "PROGRA~1", false},
		{"Trailing dot", "file.txt.", false},
		{"Absolute path sensitive", filepath.Join(cwd, ".bashrc"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, msg := IsReadAllowed(tt.path, cwd)
			if allowed != tt.allowed {
				t.Errorf("Path: %s, Expected allowed: %v, Actual: %v, Message: %s", tt.path, tt.allowed, allowed, msg)
			}
		})
	}
}

func TestIsWriteAllowed(t *testing.T) {
	cwd, _ := os.Getwd()
	parentDir := filepath.Dir(cwd)

	tests := []struct {
		name    string
		path    string
		allowed bool
	}{
		{"Inside CWD", "test.txt", true},
		{"Outside CWD (Parent)", filepath.Join(parentDir, "hacker.txt"), false},
		{"Inside CWD/subdir", filepath.Join("internal", "api", "test.go"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, _ := IsWriteAllowed(tt.path, cwd)
			assert.Equal(t, tt.allowed, allowed)
		})
	}
}
