package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/richq/m2cv/internal/assets"
)

// setupOptimizeTest creates a temp directory and changes to it for testing.
// Returns the temp dir path and a cleanup function to restore the original directory.
func setupOptimizeTest(t *testing.T) (string, func()) {
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

func TestOptimizeCommand_Structure(t *testing.T) {
	t.Parallel()

	cmd := newOptimizeCommand()

	// Verify command structure
	if cmd.Use != "optimize <application-name>" {
		t.Errorf("wrong Use: %q, want %q", cmd.Use, "optimize <application-name>")
	}

	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	// Verify flags exist
	if cmd.Flags().Lookup("model") == nil {
		t.Error("missing --model flag")
	}
	if cmd.Flags().Lookup("ats") == nil {
		t.Error("missing --ats flag")
	}

	// Verify model flag has short form
	modelFlag := cmd.Flags().ShorthandLookup("m")
	if modelFlag == nil {
		t.Error("missing -m shorthand for model flag")
	}

	// Verify ats flag default is false
	atsFlag := cmd.Flags().Lookup("ats")
	if atsFlag.DefValue != "false" {
		t.Errorf("ats flag default = %q, want %q", atsFlag.DefValue, "false")
	}

	// Verify model flag default is empty
	modelFlagLong := cmd.Flags().Lookup("model")
	if modelFlagLong.DefValue != "" {
		t.Errorf("model flag default = %q, want empty string", modelFlagLong.DefValue)
	}

	// Verify command requires exactly one argument
	if cmd.Args == nil {
		t.Error("Args function should be set (ExactArgs)")
	}
}

func TestOptimizeCommand_MissingAppFolder(t *testing.T) {
	tmpDir, cleanup := setupOptimizeTest(t)
	defer cleanup()

	// Create config file but NOT the application folder
	configContent := `base_cv_path: base-cv.md
default_model: claude-sonnet-4-20250514
`
	if err := os.WriteFile(filepath.Join(tmpDir, "m2cv.yml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	rootCmd := NewRootCommand()
	rootCmd.AddCommand(newOptimizeCommand())
	rootCmd.SetArgs([]string{"optimize", "nonexistent-app"})

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

func TestOptimizeCommand_MissingConfig(t *testing.T) {
	tmpDir, cleanup := setupOptimizeTest(t)
	defer cleanup()

	// Create application folder with job description but NO config file
	appDir := filepath.Join(tmpDir, "applications", "test-app")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		t.Fatalf("failed to create app dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(appDir, "job.txt"), []byte("test job"), 0644); err != nil {
		t.Fatalf("failed to create job file: %v", err)
	}

	rootCmd := NewRootCommand()
	rootCmd.AddCommand(newOptimizeCommand())
	rootCmd.SetArgs([]string{"optimize", "test-app"})

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

func TestOptimizeCommand_MissingJobDescription(t *testing.T) {
	tmpDir, cleanup := setupOptimizeTest(t)
	defer cleanup()

	// Create config file
	configContent := `base_cv_path: base-cv.md
default_model: claude-sonnet-4-20250514
`
	if err := os.WriteFile(filepath.Join(tmpDir, "m2cv.yml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	// Create base CV file
	if err := os.WriteFile(filepath.Join(tmpDir, "base-cv.md"), []byte("# My CV\n\nExperience..."), 0644); err != nil {
		t.Fatalf("failed to create base CV: %v", err)
	}

	// Create application folder WITHOUT job.txt
	appDir := filepath.Join(tmpDir, "applications", "test-app")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		t.Fatalf("failed to create app dir: %v", err)
	}

	rootCmd := NewRootCommand()
	rootCmd.AddCommand(newOptimizeCommand())
	rootCmd.SetArgs([]string{"optimize", "test-app"})

	// Disable preflight checks for testing
	rootCmd.PersistentPreRunE = nil

	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for missing job description, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "no .txt file found") {
		t.Errorf("error = %q, want to contain 'no .txt file found'", err.Error())
	}
}

func TestOptimizeCommand_MissingBaseCV(t *testing.T) {
	tmpDir, cleanup := setupOptimizeTest(t)
	defer cleanup()

	// Create config file pointing to nonexistent base CV
	configContent := `base_cv_path: nonexistent-cv.md
default_model: claude-sonnet-4-20250514
`
	if err := os.WriteFile(filepath.Join(tmpDir, "m2cv.yml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	// Create application folder with job description
	appDir := filepath.Join(tmpDir, "applications", "test-app")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		t.Fatalf("failed to create app dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(appDir, "job.txt"), []byte("test job description"), 0644); err != nil {
		t.Fatalf("failed to create job file: %v", err)
	}

	rootCmd := NewRootCommand()
	rootCmd.AddCommand(newOptimizeCommand())
	rootCmd.SetArgs([]string{"optimize", "test-app"})

	// Disable preflight checks for testing
	rootCmd.PersistentPreRunE = nil

	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for missing base CV, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "failed to read base CV") {
		t.Errorf("error = %q, want to contain 'failed to read base CV'", err.Error())
	}
}

func TestOptimizeCommand_MissingArgument(t *testing.T) {
	// Note: Cannot use t.Parallel() - NewRootCommand writes to global vars
	rootCmd := NewRootCommand()
	rootCmd.AddCommand(newOptimizeCommand())
	rootCmd.SetArgs([]string{"optimize"})

	// Disable preflight checks for testing
	rootCmd.PersistentPreRunE = nil

	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for missing argument, got nil")
	}
}

func TestOptimizeCommand_HelpOutput(t *testing.T) {
	// Note: Cannot use t.Parallel() - NewRootCommand writes to global vars
	rootCmd := NewRootCommand()
	rootCmd.AddCommand(newOptimizeCommand())
	rootCmd.SetArgs([]string{"optimize", "--help"})

	// Help should not error
	if err := rootCmd.Execute(); err != nil {
		t.Errorf("help command failed: %v", err)
	}
}

func TestOptimizeCommand_ATSPromptExists(t *testing.T) {
	t.Parallel()

	// Verify both prompts exist and are different
	optimizePrompt, err := assets.GetPrompt("optimize")
	if err != nil {
		t.Fatalf("failed to get optimize prompt: %v", err)
	}

	atsPrompt, err := assets.GetPrompt("optimize-ats")
	if err != nil {
		t.Fatalf("failed to get optimize-ats prompt: %v", err)
	}

	if optimizePrompt == "" {
		t.Error("optimize prompt is empty")
	}
	if atsPrompt == "" {
		t.Error("optimize-ats prompt is empty")
	}

	// They should be different prompts
	if optimizePrompt == atsPrompt {
		t.Error("optimize and optimize-ats prompts should be different")
	}
}

func TestOptimizeCommand_ModelFlagBinding(t *testing.T) {
	t.Parallel()

	cmd := newOptimizeCommand()

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

func TestOptimizeCommand_ATSFlagBinding(t *testing.T) {
	t.Parallel()

	cmd := newOptimizeCommand()

	// Default should be false
	val, err := cmd.Flags().GetBool("ats")
	if err != nil {
		t.Errorf("failed to get ats flag: %v", err)
	}
	if val != false {
		t.Errorf("ats flag default = %v, want false", val)
	}

	// Test that ats flag can be set
	if err := cmd.Flags().Set("ats", "true"); err != nil {
		t.Errorf("failed to set ats flag: %v", err)
	}

	val, err = cmd.Flags().GetBool("ats")
	if err != nil {
		t.Errorf("failed to get ats flag: %v", err)
	}
	if val != true {
		t.Errorf("ats flag = %v, want true", val)
	}
}

// TestOptimizeCommand_ErrorOrder verifies errors are caught in the expected order:
// 1. Missing application folder (first check)
// 2. Missing config (second check)
// 3. Missing base CV (third check)
// 4. Missing job description (fourth check)
func TestOptimizeCommand_ErrorOrder(t *testing.T) {
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
				configContent := `base_cv_path: base-cv.md`
				if err := os.WriteFile(filepath.Join(tmpDir, "m2cv.yml"), []byte(configContent), 0644); err != nil {
					t.Fatalf("failed to create config: %v", err)
				}
			},
			expectedError: "application folder not found",
		},
		{
			name: "missing config checked second",
			setup: func(t *testing.T, tmpDir string) {
				// Create app folder but no config
				appDir := filepath.Join(tmpDir, "applications", "test-app")
				if err := os.MkdirAll(appDir, 0755); err != nil {
					t.Fatalf("failed to create app dir: %v", err)
				}
			},
			expectedError: "m2cv.yml not found",
		},
		{
			name: "missing base CV checked third",
			setup: func(t *testing.T, tmpDir string) {
				// Create app folder and config, but no base CV
				appDir := filepath.Join(tmpDir, "applications", "test-app")
				if err := os.MkdirAll(appDir, 0755); err != nil {
					t.Fatalf("failed to create app dir: %v", err)
				}
				if err := os.WriteFile(filepath.Join(appDir, "job.txt"), []byte("job"), 0644); err != nil {
					t.Fatalf("failed to create job file: %v", err)
				}
				configContent := `base_cv_path: missing-cv.md`
				if err := os.WriteFile(filepath.Join(tmpDir, "m2cv.yml"), []byte(configContent), 0644); err != nil {
					t.Fatalf("failed to create config: %v", err)
				}
			},
			expectedError: "failed to read base CV",
		},
		{
			name: "missing job description checked fourth",
			setup: func(t *testing.T, tmpDir string) {
				// Create app folder, config, and base CV, but no job description
				appDir := filepath.Join(tmpDir, "applications", "test-app")
				if err := os.MkdirAll(appDir, 0755); err != nil {
					t.Fatalf("failed to create app dir: %v", err)
				}
				configContent := `base_cv_path: base-cv.md`
				if err := os.WriteFile(filepath.Join(tmpDir, "m2cv.yml"), []byte(configContent), 0644); err != nil {
					t.Fatalf("failed to create config: %v", err)
				}
				if err := os.WriteFile(filepath.Join(tmpDir, "base-cv.md"), []byte("# CV"), 0644); err != nil {
					t.Fatalf("failed to create base CV: %v", err)
				}
			},
			expectedError: "no .txt file found",
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
			rootCmd.AddCommand(newOptimizeCommand())
			rootCmd.SetArgs([]string{"optimize", "test-app"})
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
