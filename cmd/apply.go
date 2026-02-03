package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/richq/m2cv/internal/executor"
	"github.com/richq/m2cv/internal/extractor"
	"github.com/richq/m2cv/internal/filesystem"
	"github.com/spf13/cobra"
)

// newApplyCommand creates the apply subcommand.
func newApplyCommand() *cobra.Command {
	var (
		name string
		dir  string
	)

	cmd := &cobra.Command{
		Use:   "apply <job-description-file>",
		Short: "Create a job application folder from a job description",
		Long: `Create a new job application folder with an AI-extracted name from the job description.

The folder is created under the applications directory (default: "applications/").
The job description file is copied into the new folder.

Examples:
  m2cv apply job-posting.txt
  m2cv apply --name acme-corp-engineer job.txt
  m2cv apply --dir my-apps job.txt`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runApply(cmd.Context(), args[0], name, dir)
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "override folder name (skip Claude extraction)")
	cmd.Flags().StringVarP(&dir, "dir", "d", "applications", "applications directory")

	return cmd
}

// runApply executes the apply command logic.
func runApply(ctx context.Context, jobFile, nameOverride, applicationsDir string) error {
	// Validate job description file exists
	if _, err := os.Stat(jobFile); os.IsNotExist(err) {
		return fmt.Errorf("job description file not found: %s", jobFile)
	}

	// Read job description content
	content, err := os.ReadFile(jobFile)
	if err != nil {
		return fmt.Errorf("failed to read job description: %w", err)
	}

	// Determine folder name
	var folderName string
	if nameOverride != "" {
		// Use provided name with sanitization
		folderName = extractor.SanitizeFilename(nameOverride)
	} else {
		// Extract name using Claude
		exec := executor.NewClaudeExecutor()
		folderName, err = extractor.ExtractFolderName(ctx, exec, string(content))
		if err != nil {
			return fmt.Errorf("failed to extract folder name: %w. Use --name to specify manually", err)
		}
	}

	// Build application path
	appPath := filepath.Join(applicationsDir, folderName)

	// Initialize filesystem operations
	fs := filesystem.NewOperations()

	// Check if folder already exists
	if fs.Exists(appPath) {
		return fmt.Errorf("application folder already exists: %s. Use --name to specify a different name", appPath)
	}

	// Create application folder
	if err := fs.CreateDir(appPath, 0755); err != nil {
		return fmt.Errorf("failed to create application folder: %w", err)
	}

	// Copy job description into folder
	destFile := filepath.Join(appPath, filepath.Base(jobFile))
	if err := fs.CopyFile(jobFile, destFile); err != nil {
		return fmt.Errorf("failed to copy job description: %w", err)
	}

	// Print success message
	fmt.Printf("Created application folder: %s\n", appPath)
	fmt.Printf("Job description copied to: %s\n", destFile)

	return nil
}
