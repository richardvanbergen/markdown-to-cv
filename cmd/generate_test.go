package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupGenerateTest creates a temp directory and changes to it for testing.
// Returns the temp dir path and a cleanup function to restore the original directory.
func setupGenerateTest(t *testing.T) (string, func()) {
	t.Helper()
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir to temp dir: %v", err)
	}
	return tmpDir, func() {
		if err := os.Chdir(origDir); err != nil {
			t.Logf("warning: failed to restore dir: %v", err)
		}
	}
}

func TestGenerateCommand_Structure(t *testing.T) {
	t.Parallel()

	cmd := newGenerateCommand()

	// Verify command structure
	if cmd.Use != "generate <application-name>" {
		t.Errorf("wrong Use: %q, want %q", cmd.Use, "generate <application-name>")
	}

	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	// Verify flags exist
	if cmd.Flags().Lookup("theme") == nil {
		t.Error("missing --theme flag")
	}
	if cmd.Flags().Lookup("model") == nil {
		t.Error("missing --model flag")
	}

	// Verify model flag has short form
	modelFlag := cmd.Flags().ShorthandLookup("m")
	if modelFlag == nil {
		t.Error("missing -m shorthand for model flag")
	}

	// Verify flag defaults are empty
	themeFlag := cmd.Flags().Lookup("theme")
	if themeFlag.DefValue != "" {
		t.Errorf("theme flag default = %q, want empty string", themeFlag.DefValue)
	}

	modelFlagLong := cmd.Flags().Lookup("model")
	if modelFlagLong.DefValue != "" {
		t.Errorf("model flag default = %q, want empty string", modelFlagLong.DefValue)
	}

	// Verify command requires exactly one argument
	if cmd.Args == nil {
		t.Error("Args function should be set (ExactArgs)")
	}
}

func TestGenerateCommand_MissingApplication(t *testing.T) {
	tmpDir, cleanup := setupGenerateTest(t)
	defer cleanup()

	// Create config file but NOT the application folder
	configContent := `base_cv_path: base-cv.md
default_model: claude-sonnet-4-20250514
default_theme: even
`
	if err := os.WriteFile(filepath.Join(tmpDir, "m2cv.yml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	rootCmd := NewRootCommand()
	generateCmd := newGenerateCommand()
	generateCmd.PreRunE = nil // Disable resumed preflight check
	rootCmd.AddCommand(generateCmd)
	rootCmd.SetArgs([]string{"generate", "nonexistent-app"})

	// Disable preflight checks for testing
	rootCmd.PersistentPreRunE = nil

	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for missing application folder, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "application folder not found") {
		t.Errorf("error = %q, want to contain 'application folder not found'", err.Error())
	}
}

func TestGenerateCommand_NoOptimizedCV(t *testing.T) {
	tmpDir, cleanup := setupGenerateTest(t)
	defer cleanup()

	// Create config file
	configContent := `base_cv_path: base-cv.md
default_model: claude-sonnet-4-20250514
default_theme: even
`
	if err := os.WriteFile(filepath.Join(tmpDir, "m2cv.yml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	// Create application folder WITHOUT any optimized-cv-*.md files
	appDir := filepath.Join(tmpDir, "applications", "test-app")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		t.Fatalf("failed to create app dir: %v", err)
	}
	// Add a job.txt so the folder isn't completely empty
	if err := os.WriteFile(filepath.Join(appDir, "job.txt"), []byte("test job"), 0644); err != nil {
		t.Fatalf("failed to create job file: %v", err)
	}

	rootCmd := NewRootCommand()
	generateCmd := newGenerateCommand()
	generateCmd.PreRunE = nil // Disable resumed preflight check
	rootCmd.AddCommand(generateCmd)
	rootCmd.SetArgs([]string{"generate", "test-app"})

	// Disable preflight checks for testing
	rootCmd.PersistentPreRunE = nil

	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for no optimized CV, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "no optimized CV") {
		t.Errorf("error = %q, want to contain 'no optimized CV'", err.Error())
	}
}

func TestGenerateCommand_MissingConfig(t *testing.T) {
	tmpDir, cleanup := setupGenerateTest(t)
	defer cleanup()

	// Create application folder with optimized CV but NO config file
	appDir := filepath.Join(tmpDir, "applications", "test-app")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		t.Fatalf("failed to create app dir: %v", err)
	}
	// Add an optimized CV
	if err := os.WriteFile(filepath.Join(appDir, "optimized-cv-1.md"), []byte("# CV\n"), 0644); err != nil {
		t.Fatalf("failed to create optimized CV: %v", err)
	}

	rootCmd := NewRootCommand()
	generateCmd := newGenerateCommand()
	generateCmd.PreRunE = nil // Disable resumed preflight check
	rootCmd.AddCommand(generateCmd)
	rootCmd.SetArgs([]string{"generate", "test-app"})

	// Disable preflight checks for testing
	rootCmd.PersistentPreRunE = nil

	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for missing config, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "m2cv.yml not found") {
		t.Errorf("error = %q, want to contain 'm2cv.yml not found'", err.Error())
	}
}

func TestGenerateCommand_MissingArgument(t *testing.T) {
	// Note: Cannot use t.Parallel() - NewRootCommand writes to global vars
	rootCmd := NewRootCommand()
	generateCmd := newGenerateCommand()
	generateCmd.PreRunE = nil // Disable resumed preflight check
	rootCmd.AddCommand(generateCmd)
	rootCmd.SetArgs([]string{"generate"})

	// Disable preflight checks for testing
	rootCmd.PersistentPreRunE = nil

	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for missing argument, got nil")
	}
}

func TestGenerateCommand_HelpOutput(t *testing.T) {
	// Note: Cannot use t.Parallel() - NewRootCommand writes to global vars
	rootCmd := NewRootCommand()
	generateCmd := newGenerateCommand()
	generateCmd.PreRunE = nil // Disable resumed preflight check
	rootCmd.AddCommand(generateCmd)
	rootCmd.SetArgs([]string{"generate", "--help"})

	// Help should not error
	if err := rootCmd.Execute(); err != nil {
		t.Errorf("help command failed: %v", err)
	}
}

func TestGenerateCommand_ThemeFlagBinding(t *testing.T) {
	t.Parallel()

	cmd := newGenerateCommand()

	// Test that theme flag can be set
	if err := cmd.Flags().Set("theme", "stackoverflow"); err != nil {
		t.Errorf("failed to set theme flag: %v", err)
	}

	val, err := cmd.Flags().GetString("theme")
	if err != nil {
		t.Errorf("failed to get theme flag: %v", err)
	}
	if val != "stackoverflow" {
		t.Errorf("theme flag = %q, want %q", val, "stackoverflow")
	}
}

func TestGenerateCommand_ModelFlagBinding(t *testing.T) {
	t.Parallel()

	cmd := newGenerateCommand()

	// Test that model flag can be set
	if err := cmd.Flags().Set("model", "claude-opus-4-20250514"); err != nil {
		t.Errorf("failed to set model flag: %v", err)
	}

	val, err := cmd.Flags().GetString("model")
	if err != nil {
		t.Errorf("failed to get model flag: %v", err)
	}
	if val != "claude-opus-4-20250514" {
		t.Errorf("model flag = %q, want %q", val, "claude-opus-4-20250514")
	}
}

// TestGenerateCommand_ErrorOrder verifies errors are caught in the expected order:
// 1. Missing application folder (first check)
// 2. Missing config (second check - but only after app folder is found)
// 3. No optimized CV (third check)
func TestGenerateCommand_ErrorOrder(t *testing.T) {
	// Note: Cannot use t.Parallel() - NewRootCommand writes to global vars

	tests := []struct {
		name          string
		setup         func(t *testing.T, tmpDir string)
		expectedError string
	}{
		{
			name: "missing app folder checked first",
			setup: func(t *testing.T, tmpDir string) {
				// Create config but no app folder
				configContent := `base_cv_path: base-cv.md
default_theme: even`
				if err := os.WriteFile(filepath.Join(tmpDir, "m2cv.yml"), []byte(configContent), 0644); err != nil {
					t.Fatalf("failed to create config: %v", err)
				}
			},
			expectedError: "application folder not found",
		},
		{
			name: "missing config checked second",
			setup: func(t *testing.T, tmpDir string) {
				// Create app folder with optimized CV but no config
				appDir := filepath.Join(tmpDir, "applications", "test-app")
				if err := os.MkdirAll(appDir, 0755); err != nil {
					t.Fatalf("failed to create app dir: %v", err)
				}
				if err := os.WriteFile(filepath.Join(appDir, "optimized-cv-1.md"), []byte("# CV"), 0644); err != nil {
					t.Fatalf("failed to create CV: %v", err)
				}
			},
			expectedError: "m2cv.yml not found",
		},
		{
			name: "no optimized CV checked third",
			setup: func(t *testing.T, tmpDir string) {
				// Create app folder and config, but no optimized CV
				appDir := filepath.Join(tmpDir, "applications", "test-app")
				if err := os.MkdirAll(appDir, 0755); err != nil {
					t.Fatalf("failed to create app dir: %v", err)
				}
				configContent := `base_cv_path: base-cv.md
default_theme: even`
				if err := os.WriteFile(filepath.Join(tmpDir, "m2cv.yml"), []byte(configContent), 0644); err != nil {
					t.Fatalf("failed to create config: %v", err)
				}
			},
			expectedError: "no optimized CV",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// Note: Cannot use t.Parallel() here because we're changing directories
			tmpDir := t.TempDir()
			origDir, _ := os.Getwd()
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatalf("failed to chdir: %v", err)
			}
			defer func() {
				if err := os.Chdir(origDir); err != nil {
					t.Logf("warning: failed to restore dir: %v", err)
				}
			}()

			tt.setup(t, tmpDir)

			rootCmd := NewRootCommand()
			generateCmd := newGenerateCommand()
			generateCmd.PreRunE = nil // Disable resumed preflight check
			rootCmd.AddCommand(generateCmd)
			rootCmd.SetArgs([]string{"generate", "test-app"})
			rootCmd.PersistentPreRunE = nil

			err := rootCmd.Execute()
			if err == nil {
				t.Errorf("expected error containing %q, got nil", tt.expectedError)
				return
			}
			if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("error = %q, want to contain %q", err.Error(), tt.expectedError)
			}
		})
	}
}

// TestGenerateCommand_IntegrationRequiresClaude documents that full integration
// tests require Claude CLI, npm, and resumed to be available.
func TestGenerateCommand_IntegrationRequiresClaude(t *testing.T) {
	t.Skip("Full integration test requires Claude CLI + npm + resumed")

	// This test would:
	// 1. Create an application folder with an optimized CV
	// 2. Run generate command
	// 3. Verify resume.json and resume.pdf are created
	// 4. Verify JSON Resume schema validity
}
