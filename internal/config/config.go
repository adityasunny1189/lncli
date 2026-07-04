package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const configFileName = ".lncli.json"

// WorkspaceConfig is stored in the project root alongside .project.json.
type WorkspaceConfig struct {
	ProjectID string `json:"project_id"`
	// GitHubUsername is stored here so lncli run can push to the right repo.
	GitHubUsername string `json:"github_username,omitempty"`
}

// Load reads .lncli.json from dir (or a parent directory).
func Load(dir string) (*WorkspaceConfig, error) {
	path := filepath.Join(dir, configFileName)
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var cfg WorkspaceConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Save writes .lncli.json to dir.
func Save(dir string, cfg *WorkspaceConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, configFileName), data, 0644)
}
