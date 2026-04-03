package profile

import "time"

// Profile represents a saved agent configuration
type Profile struct {
	Name      string            `json:"name"`
	Model     string            `json:"model"`
	System    string            `json:"system"`
	EnvVars   map[string]string `json:"env_vars"`
	IsDefault bool              `json:"is_default"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// NewProfile creates a new profile instance
func NewProfile(name string) Profile {
	return Profile{
		Name:      name,
		EnvVars:   make(map[string]string),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
