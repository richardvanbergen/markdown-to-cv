// Package cmd contains the CLI command definitions for m2cv.
package cmd

import (
	"os"

	"github.com/richq/m2cv/internal/preflight"
	"github.com/spf13/cobra"
)

var (
	// Version info set via ldflags at build time
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

var (
	// Persistent flags
	cfgFile    string
	baseCVPath string
)

// NewRootCommand creates and returns the root cobra command for m2cv.
func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "m2cv",
		Short: "Markdown to CV - AI-powered resume tailoring",
		Long: `m2cv takes a job description and your base CV in markdown format,
uses Claude AI to tailor your resume for the specific position,
converts it to JSON Resume format, and exports a professionally
themed PDF using resumed.

The pipeline: Job Description + Base CV -> Claude AI -> JSON Resume -> PDF`,
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Skip preflight for non-functional commands
			switch cmd.Name() {
			case "version", "help", "completion":
				return nil
			}
			return preflight.CheckClaude()
		},
	}

	// Persistent flags available to all commands
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path (default: searches for m2cv.yml)")
	cmd.PersistentFlags().StringVar(&baseCVPath, "base-cv", "", "path to base CV markdown file")

	return cmd
}

// Execute runs the root command.
func Execute() {
	rootCmd := NewRootCommand()

	// Add subcommands
	rootCmd.AddCommand(newVersionCommand())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
