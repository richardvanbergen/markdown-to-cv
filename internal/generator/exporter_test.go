package generator

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/richq/m2cv/internal/executor"
)

func TestNewExporter(t *testing.T) {
	// This test checks if npx can be found on the system
	// Skip if npx is not available
	e, err := NewExporter()
	if err != nil {
		if strings.Contains(err.Error(), "npx not found") {
			t.Skip("npx not available on this system")
		}
		t.Fatalf("NewExporter() error = %v", err)
	}

	if e == nil {
		t.Fatal("NewExporter() returned nil exporter")
	}

	if e.npxPath == "" {
		t.Fatal("NewExporter() returned exporter with empty npxPath")
	}

	// Verify the path exists and is executable
	info, err := os.Stat(e.npxPath)
	if err != nil {
		t.Fatalf("npx path %q does not exist: %v", e.npxPath, err)
	}
	if info.IsDir() {
		t.Fatalf("npx path %q is a directory, expected file", e.npxPath)
	}
	if info.Mode()&0111 == 0 {
		t.Fatalf("npx path %q is not executable", e.npxPath)
	}
}

func TestNewExporterWithOptions(t *testing.T) {
	// Test that SkipSystemPaths option is respected
	opts := &executor.FindOptions{SkipSystemPaths: true}

	// This may or may not find npx depending on the test environment
	// We're mainly testing that the option is passed through
	_, err := NewExporterWithOptions(opts)
	if err != nil && !strings.Contains(err.Error(), "npx not found") {
		t.Fatalf("NewExporterWithOptions() unexpected error = %v", err)
	}
}

func TestExporter_CheckThemeInstalled(t *testing.T) {
	// Create a mock exporter (we don't need real npx for this test)
	e := &Exporter{npxPath: "/usr/bin/npx"}

	// Create temp directory structure
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		setupFunc   func() string // returns projectDir
		theme       string
		wantErr     bool
		errContains string
	}{
		{
			name: "theme installed",
			setupFunc: func() string {
				projectDir := filepath.Join(tmpDir, "project1")
				themePath := filepath.Join(projectDir, "node_modules", "jsonresume-theme-even")
				if err := os.MkdirAll(themePath, 0755); err != nil {
					t.Fatalf("failed to create theme dir: %v", err)
				}
				return projectDir
			},
			theme:   "even",
			wantErr: false,
		},
		{
			name: "theme not installed",
			setupFunc: func() string {
				projectDir := filepath.Join(tmpDir, "project2")
				nodeModules := filepath.Join(projectDir, "node_modules")
				if err := os.MkdirAll(nodeModules, 0755); err != nil {
					t.Fatalf("failed to create node_modules: %v", err)
				}
				return projectDir
			},
			theme:       "stackoverflow",
			wantErr:     true,
			errContains: "not installed",
		},
		{
			name: "node_modules does not exist",
			setupFunc: func() string {
				projectDir := filepath.Join(tmpDir, "project3")
				if err := os.MkdirAll(projectDir, 0755); err != nil {
					t.Fatalf("failed to create project dir: %v", err)
				}
				return projectDir
			},
			theme:       "elegant",
			wantErr:     true,
			errContains: "not installed",
		},
		{
			name: "theme path is file not directory",
			setupFunc: func() string {
				projectDir := filepath.Join(tmpDir, "project4")
				themePath := filepath.Join(projectDir, "node_modules", "jsonresume-theme-flat")
				if err := os.MkdirAll(filepath.Dir(themePath), 0755); err != nil {
					t.Fatalf("failed to create dir: %v", err)
				}
				if err := os.WriteFile(themePath, []byte("not a dir"), 0644); err != nil {
					t.Fatalf("failed to create file: %v", err)
				}
				return projectDir
			},
			theme:       "flat",
			wantErr:     true,
			errContains: "not a directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectDir := tt.setupFunc()
			err := e.CheckThemeInstalled(projectDir, tt.theme)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CheckThemeInstalled() error = nil, wantErr = true")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("CheckThemeInstalled() error = %q, want error containing %q", err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("CheckThemeInstalled() error = %v, wantErr = false", err)
			}
		})
	}
}

func TestExporter_ValidateResumedInstalled(t *testing.T) {
	e := &Exporter{npxPath: "/usr/bin/npx"}

	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		setupFunc   func() string
		wantErr     bool
		errContains string
	}{
		{
			name: "resumed installed",
			setupFunc: func() string {
				projectDir := filepath.Join(tmpDir, "resumed1")
				resumedPath := filepath.Join(projectDir, "node_modules", "resumed")
				if err := os.MkdirAll(resumedPath, 0755); err != nil {
					t.Fatalf("failed to create resumed dir: %v", err)
				}
				return projectDir
			},
			wantErr: false,
		},
		{
			name: "resumed not installed",
			setupFunc: func() string {
				projectDir := filepath.Join(tmpDir, "resumed2")
				nodeModules := filepath.Join(projectDir, "node_modules")
				if err := os.MkdirAll(nodeModules, 0755); err != nil {
					t.Fatalf("failed to create node_modules: %v", err)
				}
				return projectDir
			},
			wantErr:     true,
			errContains: "resumed not installed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectDir := tt.setupFunc()
			err := e.ValidateResumedInstalled(projectDir)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateResumedInstalled() error = nil, wantErr = true")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ValidateResumedInstalled() error = %q, want error containing %q", err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateResumedInstalled() error = %v, wantErr = false", err)
			}
		})
	}
}

func TestExporter_ExportPDF_ThemeNotInstalled(t *testing.T) {
	// Test that ExportPDF returns early if theme is not installed
	e := &Exporter{npxPath: "/usr/bin/npx"}

	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "project")
	nodeModules := filepath.Join(projectDir, "node_modules")
	if err := os.MkdirAll(nodeModules, 0755); err != nil {
		t.Fatalf("failed to create node_modules: %v", err)
	}

	// Create a dummy JSON file
	jsonPath := filepath.Join(tmpDir, "resume.json")
	if err := os.WriteFile(jsonPath, []byte(`{"basics":{}}`), 0644); err != nil {
		t.Fatalf("failed to create JSON file: %v", err)
	}

	err := e.ExportPDF(context.Background(), jsonPath, "output.pdf", "nonexistent-theme", projectDir)
	if err == nil {
		t.Error("ExportPDF() error = nil, want error about theme not installed")
		return
	}
	if !strings.Contains(err.Error(), "not installed") {
		t.Errorf("ExportPDF() error = %q, want error about theme not installed", err.Error())
	}
}

func TestExporter_NPXPath(t *testing.T) {
	testPath := "/custom/path/to/npx"
	e := &Exporter{npxPath: testPath}

	if got := e.NPXPath(); got != testPath {
		t.Errorf("NPXPath() = %q, want %q", got, testPath)
	}
}

// Integration test that requires actual npx and resumed to be installed
func TestExporter_ExportPDF_Integration(t *testing.T) {
	// Skip if npx is not available
	e, err := NewExporter()
	if err != nil {
		t.Skip("npx not available, skipping integration test")
	}

	// Create a temp project directory with node_modules structure
	tmpDir := t.TempDir()

	// Check if resumed and a theme are available
	// We need both resumed and at least one theme installed to run this test
	nodeModulesPath := filepath.Join(tmpDir, "node_modules")
	if err := os.MkdirAll(nodeModulesPath, 0755); err != nil {
		t.Fatalf("failed to create node_modules: %v", err)
	}

	// This test would require npm install resumed jsonresume-theme-even
	// which is typically done during m2cv init
	// Skip if the required packages are not available
	t.Skip("Integration test requires npm packages - run manually with: npm install resumed jsonresume-theme-even")

	// If we had the packages, the test would look like:
	// jsonPath := filepath.Join(tmpDir, "resume.json")
	// os.WriteFile(jsonPath, []byte(minimalValidResume), 0644)
	// outputPath := filepath.Join(tmpDir, "resume.pdf")
	// err = e.ExportPDF(context.Background(), jsonPath, outputPath, "even", tmpDir)
	// ... verify PDF was created
	_ = e
}
