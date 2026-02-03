package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_ValidYAML(t *testing.T) {
	// Create temp directory and file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "m2cv.yml")

	content := `base_cv_path: cv.md
default_theme: flat
themes:
  - flat
  - modern
default_model: claude-3-opus
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	repo := NewRepository()
	cfg, err := repo.Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	if cfg.BaseCVPath != "cv.md" {
		t.Errorf("BaseCVPath = %q, want %q", cfg.BaseCVPath, "cv.md")
	}
	if cfg.DefaultTheme != "flat" {
		t.Errorf("DefaultTheme = %q, want %q", cfg.DefaultTheme, "flat")
	}
	if len(cfg.Themes) != 2 || cfg.Themes[0] != "flat" || cfg.Themes[1] != "modern" {
		t.Errorf("Themes = %v, want [flat, modern]", cfg.Themes)
	}
	if cfg.DefaultModel != "claude-3-opus" {
		t.Errorf("DefaultModel = %q, want %q", cfg.DefaultModel, "claude-3-opus")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "m2cv.yml")

	// Write invalid YAML
	content := `base_cv_path: [invalid yaml
this is not valid`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	repo := NewRepository()
	_, err := repo.Load(configPath)
	if err == nil {
		t.Error("Load() error = nil, want error for invalid YAML")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	repo := NewRepository()
	_, err := repo.Load("/nonexistent/path/m2cv.yml")
	if err == nil {
		t.Error("Load() error = nil, want error for missing file")
	}
}

func TestSave_WritesValidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "m2cv.yml")

	cfg := &Config{
		BaseCVPath:   "test.md",
		DefaultTheme: "modern",
		Themes:       []string{"modern", "classic"},
		DefaultModel: "claude-3-sonnet",
	}

	repo := NewRepository()
	if err := repo.Save(configPath, cfg); err != nil {
		t.Fatalf("Save() error = %v, want nil", err)
	}

	// Load it back and verify
	loaded, err := repo.Load(configPath)
	if err != nil {
		t.Fatalf("Load() after Save() error = %v", err)
	}

	if loaded.BaseCVPath != cfg.BaseCVPath {
		t.Errorf("BaseCVPath = %q, want %q", loaded.BaseCVPath, cfg.BaseCVPath)
	}
	if loaded.DefaultTheme != cfg.DefaultTheme {
		t.Errorf("DefaultTheme = %q, want %q", loaded.DefaultTheme, cfg.DefaultTheme)
	}
	if len(loaded.Themes) != len(cfg.Themes) {
		t.Errorf("Themes length = %d, want %d", len(loaded.Themes), len(cfg.Themes))
	}
	if loaded.DefaultModel != cfg.DefaultModel {
		t.Errorf("DefaultModel = %q, want %q", loaded.DefaultModel, cfg.DefaultModel)
	}
}

func TestFind_WalksUpDirectoryTree(t *testing.T) {
	// Create structure: tmpDir/m2cv.yml, tmpDir/a/b/c (nested dirs without config)
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "m2cv.yml")

	content := `base_cv_path: found.md
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Create nested directories
	nestedDir := filepath.Join(tmpDir, "a", "b", "c")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("failed to create nested dirs: %v", err)
	}

	repo := NewRepository()
	found, err := repo.Find(nestedDir)
	if err != nil {
		t.Fatalf("Find() error = %v, want nil", err)
	}

	if found != configPath {
		t.Errorf("Find() = %q, want %q", found, configPath)
	}
}

func TestFind_ReturnsErrorWhenNoConfigFound(t *testing.T) {
	// Create a temp directory without any m2cv.yml
	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "no", "config", "here")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("failed to create nested dirs: %v", err)
	}

	repo := NewRepository()
	_, err := repo.Find(nestedDir)
	if err == nil {
		t.Error("Find() error = nil, want error when no config found")
	}
}

func TestFindWithOverrides_PrefersFlagOverEnv(t *testing.T) {
	// Set env var
	os.Setenv("M2CV_CONFIG", "/from/env/m2cv.yml")
	defer os.Unsetenv("M2CV_CONFIG")

	result, err := FindWithOverrides("/from/flag/m2cv.yml", "/some/dir")
	if err != nil {
		t.Fatalf("FindWithOverrides() error = %v, want nil", err)
	}

	if result != "/from/flag/m2cv.yml" {
		t.Errorf("FindWithOverrides() = %q, want %q", result, "/from/flag/m2cv.yml")
	}
}

func TestFindWithOverrides_PrefersEnvOverWalkUp(t *testing.T) {
	// Set env var
	os.Setenv("M2CV_CONFIG", "/from/env/m2cv.yml")
	defer os.Unsetenv("M2CV_CONFIG")

	result, err := FindWithOverrides("", "/some/dir")
	if err != nil {
		t.Fatalf("FindWithOverrides() error = %v, want nil", err)
	}

	if result != "/from/env/m2cv.yml" {
		t.Errorf("FindWithOverrides() = %q, want %q", result, "/from/env/m2cv.yml")
	}
}

func TestFindWithOverrides_FallsBackToWalkUp(t *testing.T) {
	// Ensure no env var is set
	os.Unsetenv("M2CV_CONFIG")

	// Create temp dir with config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "m2cv.yml")
	if err := os.WriteFile(configPath, []byte("base_cv_path: test.md\n"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result, err := FindWithOverrides("", tmpDir)
	if err != nil {
		t.Fatalf("FindWithOverrides() error = %v, want nil", err)
	}

	if result != configPath {
		t.Errorf("FindWithOverrides() = %q, want %q", result, configPath)
	}
}
