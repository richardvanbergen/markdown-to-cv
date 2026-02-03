// Package cmd contains the CLI command definitions for m2cv.
package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/richq/m2cv/internal/config"
	"github.com/richq/m2cv/internal/executor"
	initpkg "github.com/richq/m2cv/internal/init"
	"github.com/spf13/cobra"
)

// isInteractive checks if stdin is a terminal.
func isInteractive() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// newInitCommand creates the init subcommand for initializing m2cv projects.
func newInitCommand() *cobra.Command {
	var (
		themeName  string
		baseCVPath string
		force      bool
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new m2cv project",
		Long: `Initialize a new m2cv project in the current directory.

This command will:
1. Create an m2cv.yml configuration file
2. Initialize npm if package.json doesn't exist
3. Install resumed and the selected theme package

If no theme is specified via --theme flag, an interactive theme selector
will be shown (requires a terminal).`,
		Example: `  # Interactive mode - shows theme selector
  m2cv init

  # Non-interactive with specific theme
  m2cv init --theme even

  # With a base CV file
  m2cv init --theme even --base-cv ~/cv/base.md

  # Overwrite existing configuration
  m2cv init --theme even --force`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(cmd.Context(), themeName, baseCVPath, force)
		},
	}

	// Register flags
	cmd.Flags().StringVarP(&themeName, "theme", "t", "", "JSON Resume theme (skips interactive selection)")
	cmd.Flags().StringVar(&baseCVPath, "base-cv", "", "path to base CV markdown file")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "overwrite existing configuration")

	return cmd
}

// runInit executes the init command logic.
func runInit(ctx context.Context, themeName, baseCVPath string, force bool) error {
	// Get current working directory
	projectDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	configPath := filepath.Join(projectDir, "m2cv.yml")

	// Check if already initialized (unless --force)
	if _, err := os.Stat(configPath); err == nil {
		if !force {
			return errors.New("m2cv.yml already exists (use --force to overwrite)")
		}
		// Remove existing config so service can create new one
		if err := os.Remove(configPath); err != nil {
			return fmt.Errorf("failed to remove existing config: %w", err)
		}
	}

	// Handle theme selection
	if themeName == "" {
		if !isInteractive() {
			return errors.New("no terminal detected; use --theme flag to specify theme")
		}
		selected, err := initpkg.SelectTheme()
		if err != nil {
			return fmt.Errorf("theme selection cancelled: %w", err)
		}
		themeName = selected
	}

	// Validate theme
	if !initpkg.IsValidTheme(themeName) {
		return fmt.Errorf("invalid theme %q; available themes: %v", themeName, initpkg.AvailableThemes)
	}

	// Validate base CV path if provided
	if baseCVPath != "" {
		if _, err := os.Stat(baseCVPath); os.IsNotExist(err) {
			return fmt.Errorf("base CV file not found: %s", baseCVPath)
		}
	}

	fmt.Println("Initializing m2cv project...")

	// Create dependencies
	configRepo := config.NewRepository()
	npmExec, err := executor.NewNPMExecutor()
	if err != nil {
		return fmt.Errorf("failed to initialize npm: %w", err)
	}

	// Initialize the project
	initService := initpkg.NewService(configRepo, npmExec)
	opts := initpkg.InitOptions{
		ProjectDir:   projectDir,
		BaseCVPath:   baseCVPath,
		Theme:        themeName,
		DefaultModel: "claude-sonnet-4-20250514", // Sensible default
	}

	if err := initService.Init(ctx, opts); err != nil {
		return fmt.Errorf("initialization failed: %w", err)
	}

	// Success output
	fmt.Println()
	fmt.Println("Project initialized successfully!")
	fmt.Println()
	fmt.Printf("  Config:    %s\n", configPath)
	fmt.Printf("  Theme:     %s\n", themeName)
	if baseCVPath != "" {
		fmt.Printf("  Base CV:   %s\n", baseCVPath)
	}
	fmt.Println()
	fmt.Println("Next steps:")
	if baseCVPath == "" {
		fmt.Println("  1. Set base_cv_path in m2cv.yml to your CV markdown file")
		fmt.Println("  2. Run 'm2cv apply <job-description.txt>' to generate a tailored resume")
	} else {
		fmt.Println("  1. Run 'm2cv apply <job-description.txt>' to generate a tailored resume")
	}

	return nil
}
