package memory

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/onedayallday33-a11y/openshannon-go/internal/agent"
)

// PersistenceManager handles saving session state to the filesystem
type PersistenceManager struct {
	BaseDir string
}

// NewPersistenceManager returns a manager pointing to ~/.openshannon/sessions/
func NewPersistenceManager() *PersistenceManager {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".openshannon", "sessions")
	return &PersistenceManager{BaseDir: dir}
}

func (pm *PersistenceManager) ensureDir() error {
	return os.MkdirAll(pm.BaseDir, 0755)
}

// SaveSession writes the session state to a JSON file
func (pm *PersistenceManager) SaveSession(state SessionState) error {
	if err := pm.ensureDir(); err != nil {
		return err
	}

	path := filepath.Join(pm.BaseDir, state.ID+".json")
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// LoadSession reads a session state from a JSON file
func (pm *PersistenceManager) LoadSession(id string) (SessionState, error) {
	path := filepath.Join(pm.BaseDir, id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return SessionState{}, err
	}

	var state SessionState
	if err := json.Unmarshal(data, &state); err != nil {
		return SessionState{}, err
	}
	return state, nil
}

// ListSessions returns a list of session IDs
func (pm *PersistenceManager) ListSessions() ([]string, error) {
	if err := pm.ensureDir(); err != nil {
		return nil, err
	}

	files, err := os.ReadDir(pm.BaseDir)
	if err != nil {
		return nil, err
	}

	var ids []string
	for _, f := range files {
		if !f.IsDir() && filepath.Ext(f.Name()) == ".json" {
			ids = append(ids, f.Name()[:len(f.Name())-len(".json")])
		}
	}
	return ids, nil
}

// AutoSave is a convenience function to save current agent state
func (pm *PersistenceManager) AutoSave(a *agent.Agent) error {
	state := Snapshot(a.ID, a.History, agent.GetTaskManager())
	return pm.SaveSession(state)
}
