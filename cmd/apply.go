package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/richq/m2cv/internal/extractor"
	"github.com/richq/m2cv/internal/filesystem"
	"github.com/spf13/cobra"
)

// newApplyCommand creates the apply subcommand.
func newApplyCommand() *cobra.Command {
	var dir string
	var fileFlag bool

	cmd := &cobra.Command{
		Use:   "apply <job-posting> <job-name>",
		Short: "Create a job application folder from a job description",
		Long: `Create a new job application folder from a job description.

The first argument is the job posting content by default. Use --file to treat
it as a file path instead.

Input modes:
  - Direct content (default): first argument is the job posting text
  - File input (--file): first argument is a file path
  - Stdin input: use "-" as first argument

The folder is created under the applications directory (default: "applications/").
When using --file, the job description is copied with its original filename.

Examples:
  m2cv apply "$(pbpaste)" acme-engineer         # content input from clipboard
  m2cv apply "Job posting text..." acme-job     # direct content
  m2cv apply - acme-engineer < job.txt          # stdin input
  m2cv apply --file job-posting.txt acme-eng    # file input
  m2cv apply --dir my-apps "$(pbpaste)" acme    # custom applications directory`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runApply(args[0], args[1], dir, fileFlag, cmd.InOrStdin())
		},
	}

	cmd.Flags().StringVarP(&dir, "dir", "d", "applications", "applications directory")
	cmd.Flags().BoolVarP(&fileFlag, "file", "f", false, "treat first argument as file path")

	return cmd
}

// applyInput represents the source of job posting content.
type applyInput struct {
	content  string // the job posting content
	filePath string // original file path (empty if content was passed directly or via stdin)
}

// parseApplyInput determines input based on the file flag and stdin marker.
func parseApplyInput(input string, fileFlag bool, stdin io.Reader) (*applyInput, error) {
	// Check for stdin
	if input == "-" {
		data, err := io.ReadAll(stdin)
		if err != nil {
			return nil, fmt.Errorf("failed to read from stdin: %w", err)
		}
		return &applyInput{content: string(data)}, nil
	}

	// Check if --file flag is set
	if fileFlag {
		if _, err := os.Stat(input); os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", input)
		}
		data, err := os.ReadFile(input)
		if err != nil {
			return nil, fmt.Errorf("failed to read job description file: %w", err)
		}
		return &applyInput{content: string(data), filePath: input}, nil
	}

	// Treat as direct content (default)
	return &applyInput{content: input}, nil
}

// runApply executes the apply command logic.
func runApply(jobInput, jobName, applicationsDir string, fileFlag bool, stdin io.Reader) error {
	// Parse input to get content
	input, err := parseApplyInput(jobInput, fileFlag, stdin)
	if err != nil {
		return err
	}

	if input.content == "" {
		return fmt.Errorf("job posting content is empty")
	}

	// Sanitize the provided job name
	folderName := extractor.SanitizeFilename(jobName)

	// Build application path
	appPath := filepath.Join(applicationsDir, folderName)

	// Initialize filesystem operations
	fs := filesystem.NewOperations()

	// Check if folder already exists
	if fs.Exists(appPath) {
		return fmt.Errorf("application folder already exists: %s. Provide a different job-name", appPath)
	}

	// Create application folder
	if err := fs.CreateDir(appPath, 0755); err != nil {
		return fmt.Errorf("failed to create application folder: %w", err)
	}

	// Write job description to folder
	var destFile string
	if input.filePath != "" {
		// Copy original file if input was from a file
		destFile = filepath.Join(appPath, filepath.Base(input.filePath))
		if err := fs.CopyFile(input.filePath, destFile); err != nil {
			return fmt.Errorf("failed to copy job description: %w", err)
		}
	} else {
		// Write content to job-description.txt if input was direct content or stdin
		destFile = filepath.Join(appPath, "job-description.txt")
		if err := os.WriteFile(destFile, []byte(input.content), 0644); err != nil {
			return fmt.Errorf("failed to write job description: %w", err)
		}
	}

	// Print success message
	fmt.Printf("Created application folder: %s\n", appPath)
	fmt.Printf("Job description saved to: %s\n", destFile)

	return nil
}
