package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApplyCommand_WithNameFlag(t *testing.T) {
	t.Parallel()

	// Create temp directory structure
	tmpDir := t.TempDir()
	applicationsDir := filepath.Join(tmpDir, "applications")
	jobFile := filepath.Join(tmpDir, "job.txt")

	// Create job description file
	jobContent := "Software Engineer at Test Company\nBuilding great things."
	if err := os.WriteFile(jobFile, []byte(jobContent), 0644); err != nil {
		t.Fatalf("failed to create job file: %v", err)
	}

	// Build and execute command
	rootCmd := NewRootCommand()
	rootCmd.AddCommand(newApplyCommand())
	rootCmd.SetArgs([]string{"apply", "--name", "test-company-role", "--dir", applicationsDir, jobFile})

	// Disable preflight checks for testing
	rootCmd.PersistentPreRunE = nil

	if err := rootCmd.Execute(); err != nil {
		t.Errorf("apply command failed: %v", err)
	}

	// Verify application folder exists
	appPath := filepath.Join(applicationsDir, "test-company-role")
	if _, err := os.Stat(appPath); os.IsNotExist(err) {
		t.Errorf("application folder not created at %s", appPath)
	}

	// Verify job file was copied
	copiedJobFile := filepath.Join(appPath, "job.txt")
	content, err := os.ReadFile(copiedJobFile)
	if err != nil {
		t.Errorf("job file not copied: %v", err)
	}
	if string(content) != jobContent {
		t.Errorf("copied job content = %q, want %q", string(content), jobContent)
	}
}

func TestApplyCommand_MissingFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	rootCmd := NewRootCommand()
	rootCmd.AddCommand(newApplyCommand())
	rootCmd.SetArgs([]string{"apply", "--name", "test", "--dir", tmpDir, filepath.Join(tmpDir, "nonexistent.txt")})

	// Disable preflight checks for testing
	rootCmd.PersistentPreRunE = nil

	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
	if err != nil && !containsSubstring(err.Error(), "not found") {
		t.Errorf("error = %q, want to contain 'not found'", err.Error())
	}
}

func TestApplyCommand_FolderExists(t *testing.T) {
	t.Parallel()

	// Create temp directory with existing application folder
	tmpDir := t.TempDir()
	applicationsDir := filepath.Join(tmpDir, "applications")
	existingFolder := filepath.Join(applicationsDir, "existing-folder")
	if err := os.MkdirAll(existingFolder, 0755); err != nil {
		t.Fatalf("failed to create existing folder: %v", err)
	}

	// Create job file
	jobFile := filepath.Join(tmpDir, "job.txt")
	if err := os.WriteFile(jobFile, []byte("test job"), 0644); err != nil {
		t.Fatalf("failed to create job file: %v", err)
	}

	rootCmd := NewRootCommand()
	rootCmd.AddCommand(newApplyCommand())
	rootCmd.SetArgs([]string{"apply", "--name", "existing-folder", "--dir", applicationsDir, jobFile})

	// Disable preflight checks for testing
	rootCmd.PersistentPreRunE = nil

	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for existing folder, got nil")
	}
	if err != nil && !containsSubstring(err.Error(), "already exists") {
		t.Errorf("error = %q, want to contain 'already exists'", err.Error())
	}
}

func TestApplyCommand_CustomDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	customAppsDir := filepath.Join(tmpDir, "custom-apps")
	jobFile := filepath.Join(tmpDir, "job.txt")

	// Create job file
	if err := os.WriteFile(jobFile, []byte("test job content"), 0644); err != nil {
		t.Fatalf("failed to create job file: %v", err)
	}

	rootCmd := NewRootCommand()
	rootCmd.AddCommand(newApplyCommand())
	rootCmd.SetArgs([]string{"apply", "--name", "myapp", "--dir", customAppsDir, jobFile})

	// Disable preflight checks for testing
	rootCmd.PersistentPreRunE = nil

	if err := rootCmd.Execute(); err != nil {
		t.Errorf("apply command failed: %v", err)
	}

	// Verify folder exists in custom directory
	appPath := filepath.Join(customAppsDir, "myapp")
	if _, err := os.Stat(appPath); os.IsNotExist(err) {
		t.Errorf("application folder not created at %s", appPath)
	}
}

func TestApplyCommand_SanitizesNameFlag(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	applicationsDir := filepath.Join(tmpDir, "applications")
	jobFile := filepath.Join(tmpDir, "job.txt")

	// Create job file
	if err := os.WriteFile(jobFile, []byte("test job"), 0644); err != nil {
		t.Fatalf("failed to create job file: %v", err)
	}

	rootCmd := NewRootCommand()
	rootCmd.AddCommand(newApplyCommand())
	// Name with special characters should be sanitized
	rootCmd.SetArgs([]string{"apply", "--name", "Test Company / Role Name", "--dir", applicationsDir, jobFile})

	// Disable preflight checks for testing
	rootCmd.PersistentPreRunE = nil

	if err := rootCmd.Execute(); err != nil {
		t.Errorf("apply command failed: %v", err)
	}

	// Verify sanitized folder name
	appPath := filepath.Join(applicationsDir, "test-company-role-name")
	if _, err := os.Stat(appPath); os.IsNotExist(err) {
		t.Errorf("sanitized folder not created at %s", appPath)
	}
}

func TestApplyCommand_MissingArgument(t *testing.T) {
	t.Parallel()

	rootCmd := NewRootCommand()
	rootCmd.AddCommand(newApplyCommand())
	rootCmd.SetArgs([]string{"apply"})

	// Disable preflight checks for testing
	rootCmd.PersistentPreRunE = nil

	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for missing argument, got nil")
	}
}

func TestApplyCommand_HelpOutput(t *testing.T) {
	t.Parallel()

	rootCmd := NewRootCommand()
	rootCmd.AddCommand(newApplyCommand())
	rootCmd.SetArgs([]string{"apply", "--help"})

	// Help should not error
	if err := rootCmd.Execute(); err != nil {
		t.Errorf("help command failed: %v", err)
	}
}

// containsSubstring checks if s contains substr (case-insensitive).
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstringHelper(s, substr))
}

func containsSubstringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
