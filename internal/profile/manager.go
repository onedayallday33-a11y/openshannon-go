package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ProfileManager handles CRUD operations for profiles
type ProfileManager struct {
	BaseDir string
}

// NewProfileManager returns a manager pointing to ~/.openshannon/profiles/
func NewProfileManager() *ProfileManager {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".openshannon", "profiles")
	return &ProfileManager{BaseDir: dir}
}

func (m *ProfileManager) ensureDir() error {
	return os.MkdirAll(m.BaseDir, 0755)
}

// SaveProfile writes a profile to disk
func (m *ProfileManager) SaveProfile(p Profile) error {
	if err := m.ensureDir(); err != nil {
		return err
	}

	p.UpdatedAt = time.Now()
	path := filepath.Join(m.BaseDir, p.Name+".json")
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// GetProfile loads a profile by name
func (m *ProfileManager) GetProfile(name string) (Profile, error) {
	path := filepath.Join(m.BaseDir, name+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return Profile{}, err
	}

	var p Profile
	if err := json.Unmarshal(data, &p); err != nil {
		return Profile{}, err
	}
	return p, nil
}

// ListProfiles returns a list of all profile names
func (m *ProfileManager) ListProfiles() ([]string, error) {
	if err := m.ensureDir(); err != nil {
		return nil, err
	}

	files, err := os.ReadDir(m.BaseDir)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, f := range files {
		if !f.IsDir() && filepath.Ext(f.Name()) == ".json" {
			names = append(names, f.Name()[:len(f.Name())-len(".json")])
		}
	}
	return names, nil
}

// GetDefaultProfile finds the profile marked as default
func (m *ProfileManager) GetDefaultProfile() (Profile, error) {
	names, err := m.ListProfiles()
	if err != nil {
		return Profile{}, err
	}

	for _, name := range names {
		p, err := m.GetProfile(name)
		if err == nil && p.IsDefault {
			return p, nil
		}
	}

	return Profile{}, fmt.Errorf("no default profile found")
}
