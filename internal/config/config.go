// Package config provides configuration loading and discovery for m2cv.
// It supports walk-up directory tree search to find m2cv.yml files,
// similar to how git finds .git directories.
package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the m2cv.yml configuration file.
type Config struct {
	BaseCVPath   string   `yaml:"base_cv_path"`
	DefaultTheme string   `yaml:"default_theme"`
	Themes       []string `yaml:"themes"`
	DefaultModel string   `yaml:"default_model"`
}

// Repository defines the interface for configuration operations.
type Repository interface {
	// Load reads and parses a configuration file from the given path.
	Load(configPath string) (*Config, error)
	// Save writes the configuration to the given path.
	Save(configPath string, cfg *Config) error
	// Find walks up the directory tree from startDir looking for m2cv.yml.
	Find(startDir string) (string, error)
}

// yamlRepository implements Repository using YAML file storage.
type yamlRepository struct{}

// NewRepository creates a new configuration repository.
func NewRepository() Repository {
	return &yamlRepository{}
}

// Load reads and parses a configuration file from the given path.
func (r *yamlRepository) Load(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Save writes the configuration to the given path.
func (r *yamlRepository) Save(configPath string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// Find walks up the directory tree from startDir looking for m2cv.yml.
// Returns the full path to the config file if found.
func (r *yamlRepository) Find(startDir string) (string, error) {
	// Resolve to absolute path
	absPath, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}

	// Walk up the directory tree
	dir := absPath
	for {
		configPath := filepath.Join(dir, "m2cv.yml")
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}

		// Move to parent directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			break
		}
		dir = parent
	}

	return "", errors.New("m2cv.yml not found in directory tree")
}

// FindWithOverrides implements the full config discovery order:
// 1. If configFlag is non-empty, return it (explicit --config flag)
// 2. If M2CV_CONFIG env var is set, return it
// 3. Otherwise, walk up from startDir looking for m2cv.yml
func FindWithOverrides(configFlag, startDir string) (string, error) {
	// Check explicit flag first
	if configFlag != "" {
		return configFlag, nil
	}

	// Check environment variable
	if envConfig := os.Getenv("M2CV_CONFIG"); envConfig != "" {
		return envConfig, nil
	}

	// Fall back to walk-up discovery
	repo := NewRepository()
	return repo.Find(startDir)
}
