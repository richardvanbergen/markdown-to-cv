package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newVersionCommand creates the version subcommand.
func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number of m2cv",
		Long:  "Display the version, commit hash, and build date of m2cv.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("m2cv version %s\n", version)
			fmt.Printf("  commit:  %s\n", commit)
			fmt.Printf("  built:   %s\n", date)
		},
	}
}
