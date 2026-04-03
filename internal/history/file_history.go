package history

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// FileHistory handles snapshots of files for rewind functionality
type FileHistory struct {
	BaseDir   string
	SessionID string
}

type Snapshot struct {
	ID        string
	Timestamp time.Time
	Files     map[string]string // Original path -> Backup path
}

func NewFileHistory(sessionID string) *FileHistory {
	home, _ := os.UserHomeDir()
	baseDir := filepath.Join(home, ".openshannon", "file-history", sessionID)
	return &FileHistory{
		BaseDir:   baseDir,
		SessionID: sessionID,
	}
}

func (fh *FileHistory) TrackFile(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil // File doesn't exist yet, nothing to backup
	}

	// Create backup
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	hash := sha256.Sum256([]byte(path))
	backupName := hex.EncodeToString(hash[:16]) + "_" + fmt.Sprint(time.Now().UnixNano())
	backupPath := filepath.Join(fh.BaseDir, backupName)

	if err := os.MkdirAll(fh.BaseDir, 0755); err != nil {
		return err
	}

	return os.WriteFile(backupPath, content, 0644)
}

func (fh *FileHistory) BackupFile(path string, snapshotID string) (string, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", nil 
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256([]byte(path))
	backupName := fmt.Sprintf("%s_%s", hex.EncodeToString(hash[:16]), snapshotID)
	backupPath := filepath.Join(fh.BaseDir, backupName)

	if err := os.MkdirAll(fh.BaseDir, 0755); err != nil {
		return "", err
	}

	if err := os.WriteFile(backupPath, content, 0644); err != nil {
		return "", err
	}

	return backupPath, nil
}

func (fh *FileHistory) Restore(backupPath, originalPath string) error {
	if backupPath == "" {
		// If backup is empty, it means file didn't exist at that time
		return os.Remove(originalPath)
	}

	content, err := os.ReadFile(backupPath)
	if err != nil {
		return err
	}

	return os.WriteFile(originalPath, content, 0644)
}

func (fh *FileHistory) Cleanup() error {
	return os.RemoveAll(fh.BaseDir)
}
