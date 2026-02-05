package cmd

import (
	"fmt"

	"github.com/richq/m2cv/internal/mcp"
	"github.com/spf13/cobra"
)

// newMCPCommand creates the hidden mcp subcommand for running as an MCP server.
// This is used internally by the optimize --interactive command.
func newMCPCommand() *cobra.Command {
	var contextData string

	cmd := &cobra.Command{
		Use:    "mcp",
		Short:  "Run as MCP server (internal use)",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if contextData == "" {
				return fmt.Errorf("--context is required")
			}

			ctx, err := mcp.DecodeContext(contextData)
			if err != nil {
				return fmt.Errorf("failed to decode context: %w", err)
			}

			server := mcp.NewServer(ctx)
			return server.Serve()
		},
	}

	cmd.Flags().StringVar(&contextData, "context", "", "base64-encoded context data")
	_ = cmd.MarkFlagRequired("context")

	return cmd
}
