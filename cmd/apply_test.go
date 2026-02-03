package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplyCommand_ContentInput_WithJobName(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	applicationsDir := filepath.Join(tmpDir, "applications")

	jobContent := "Software Engineer at Acme Corp\nJoin our team!"

	rootCmd := NewRootCommand()
	rootCmd.AddCommand(newApplyCommand())
	rootCmd.SetArgs([]string{"apply", "--dir", applicationsDir, jobContent, "acme-engineer"})
	rootCmd.PersistentPreRunE = nil

	if err := rootCmd.Execute(); err != nil {
		t.Errorf("apply command failed: %v", err)
	}

	// Verify application folder exists
	appPath := filepath.Join(applicationsDir, "acme-engineer")
	if _, err := os.Stat(appPath); os.IsNotExist(err) {
		t.Errorf("application folder not created at %s", appPath)
	}

	// Verify job content was written to job-description.txt
	destFile := filepath.Join(appPath, "job-description.txt")
	content, err := os.ReadFile(destFile)
	if err != nil {
		t.Errorf("job description not created: %v", err)
	}
	if string(content) != jobContent {
		t.Errorf("job content = %q, want %q", string(content), jobContent)
	}
}

func TestApplyCommand_StdinInput(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	applicationsDir := filepath.Join(tmpDir, "applications")

	jobContent := "DevOps Engineer at CloudCo\nRemote position available."

	rootCmd := NewRootCommand()
	rootCmd.AddCommand(newApplyCommand())
	rootCmd.SetArgs([]string{"apply", "--dir", applicationsDir, "-", "cloudco-devops"})
	rootCmd.SetIn(bytes.NewBufferString(jobContent))
	rootCmd.PersistentPreRunE = nil

	if err := rootCmd.Execute(); err != nil {
		t.Errorf("apply command failed: %v", err)
	}

	// Verify application folder exists
	appPath := filepath.Join(applicationsDir, "cloudco-devops")
	if _, err := os.Stat(appPath); os.IsNotExist(err) {
		t.Errorf("application folder not created at %s", appPath)
	}

	// Verify job content was written to job-description.txt
	destFile := filepath.Join(appPath, "job-description.txt")
	content, err := os.ReadFile(destFile)
	if err != nil {
		t.Errorf("job description not created: %v", err)
	}
	if string(content) != jobContent {
		t.Errorf("job content = %q, want %q", string(content), jobContent)
	}
}

func TestApplyCommand_FileFlag(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	applicationsDir := filepath.Join(tmpDir, "applications")
	jobFile := filepath.Join(tmpDir, "job.txt")

	jobContent := "Software Engineer at Test Company\nBuilding great things."
	if err := os.WriteFile(jobFile, []byte(jobContent), 0644); err != nil {
		t.Fatalf("failed to create job file: %v", err)
	}

	rootCmd := NewRootCommand()
	rootCmd.AddCommand(newApplyCommand())
	rootCmd.SetArgs([]string{"apply", "--dir", applicationsDir, "--file", jobFile, "test-company-role"})
	rootCmd.PersistentPreRunE = nil

	if err := rootCmd.Execute(); err != nil {
		t.Errorf("apply command failed: %v", err)
	}

	// Verify application folder exists
	appPath := filepath.Join(applicationsDir, "test-company-role")
	if _, err := os.Stat(appPath); os.IsNotExist(err) {
		t.Errorf("application folder not created at %s", appPath)
	}

	// Verify job file was copied with original name
	copiedJobFile := filepath.Join(appPath, "job.txt")
	content, err := os.ReadFile(copiedJobFile)
	if err != nil {
		t.Errorf("job file not copied: %v", err)
	}
	if string(content) != jobContent {
		t.Errorf("copied job content = %q, want %q", string(content), jobContent)
	}
}

func TestApplyCommand_FileFlagMissingFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	applicationsDir := filepath.Join(tmpDir, "applications")

	rootCmd := NewRootCommand()
	rootCmd.AddCommand(newApplyCommand())
	rootCmd.SetArgs([]string{"apply", "--dir", applicationsDir, "--file", "/nonexistent/file.txt", "test-job"})
	rootCmd.PersistentPreRunE = nil

	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "file not found") {
		t.Errorf("error = %q, want to contain 'file not found'", err.Error())
	}
}

func TestApplyCommand_FilePreservesOriginalFilename(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	applicationsDir := filepath.Join(tmpDir, "applications")
	jobFile := filepath.Join(tmpDir, "acme-posting.md")

	if err := os.WriteFile(jobFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create job file: %v", err)
	}

	rootCmd := NewRootCommand()
	rootCmd.AddCommand(newApplyCommand())
	rootCmd.SetArgs([]string{"apply", "--dir", applicationsDir, "--file", jobFile, "acme"})
	rootCmd.PersistentPreRunE = nil

	if err := rootCmd.Execute(); err != nil {
		t.Errorf("apply command failed: %v", err)
	}

	// Verify file was copied with original name
	copiedFile := filepath.Join(applicationsDir, "acme", "acme-posting.md")
	if _, err := os.Stat(copiedFile); os.IsNotExist(err) {
		t.Errorf("job file not copied with original name at %s", copiedFile)
	}
}

func TestApplyCommand_SanitizesJobName(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	applicationsDir := filepath.Join(tmpDir, "applications")

	rootCmd := NewRootCommand()
	rootCmd.AddCommand(newApplyCommand())
	// Name with special characters should be sanitized
	rootCmd.SetArgs([]string{"apply", "--dir", applicationsDir, "test job content", "Test Company / Role Name"})
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

func TestApplyCommand_FolderExists(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	applicationsDir := filepath.Join(tmpDir, "applications")
	existingFolder := filepath.Join(applicationsDir, "existing-folder")
	if err := os.MkdirAll(existingFolder, 0755); err != nil {
		t.Fatalf("failed to create existing folder: %v", err)
	}

	rootCmd := NewRootCommand()
	rootCmd.AddCommand(newApplyCommand())
	rootCmd.SetArgs([]string{"apply", "--dir", applicationsDir, "test job content", "existing-folder"})
	rootCmd.PersistentPreRunE = nil

	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for existing folder, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error = %q, want to contain 'already exists'", err.Error())
	}
}

func TestApplyCommand_EmptyContent(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	rootCmd := NewRootCommand()
	rootCmd.AddCommand(newApplyCommand())
	rootCmd.SetArgs([]string{"apply", "--dir", tmpDir, "-", "test-job"})
	rootCmd.SetIn(bytes.NewBufferString(""))
	rootCmd.PersistentPreRunE = nil

	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for empty content, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "empty") {
		t.Errorf("error = %q, want to contain 'empty'", err.Error())
	}
}

func TestApplyCommand_MissingArguments(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		args []string
	}{
		{"no arguments", []string{"apply"}},
		{"only one argument", []string{"apply", "content"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rootCmd := NewRootCommand()
			rootCmd.AddCommand(newApplyCommand())
			rootCmd.SetArgs(tc.args)
			rootCmd.PersistentPreRunE = nil

			err := rootCmd.Execute()
			if err == nil {
				t.Errorf("expected error for %s, got nil", tc.name)
			}
		})
	}
}

func TestApplyCommand_CustomDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	customAppsDir := filepath.Join(tmpDir, "custom-apps")

	rootCmd := NewRootCommand()
	rootCmd.AddCommand(newApplyCommand())
	rootCmd.SetArgs([]string{"apply", "--dir", customAppsDir, "test job content", "myapp"})
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

func TestApplyCommand_TooManyArguments(t *testing.T) {
	t.Parallel()

	rootCmd := NewRootCommand()
	rootCmd.AddCommand(newApplyCommand())
	rootCmd.SetArgs([]string{"apply", "arg1", "arg2", "arg3"})
	rootCmd.PersistentPreRunE = nil

	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for too many arguments, got nil")
	}
}

func TestApplyCommand_HelpOutput(t *testing.T) {
	t.Parallel()

	rootCmd := NewRootCommand()
	rootCmd.AddCommand(newApplyCommand())
	rootCmd.SetArgs([]string{"apply", "--help"})

	if err := rootCmd.Execute(); err != nil {
		t.Errorf("help command failed: %v", err)
	}
}

func TestApplyCommand_ShortFlags(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	applicationsDir := filepath.Join(tmpDir, "applications")
	jobFile := filepath.Join(tmpDir, "job.txt")

	if err := os.WriteFile(jobFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create job file: %v", err)
	}

	rootCmd := NewRootCommand()
	rootCmd.AddCommand(newApplyCommand())
	// Test short flags: -d for --dir and -f for --file
	rootCmd.SetArgs([]string{"apply", "-d", applicationsDir, "-f", jobFile, "test-app"})
	rootCmd.PersistentPreRunE = nil

	if err := rootCmd.Execute(); err != nil {
		t.Errorf("apply command with short flags failed: %v", err)
	}

	// Verify folder was created
	appPath := filepath.Join(applicationsDir, "test-app")
	if _, err := os.Stat(appPath); os.IsNotExist(err) {
		t.Errorf("application folder not created at %s", appPath)
	}
}
