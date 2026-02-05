package mcp

import (
	"context"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/richq/m2cv/internal/application"
)

// NewWriteOptimizedResumeTool creates the tool definition for writing an optimized resume.
func NewWriteOptimizedResumeTool() mcp.Tool {
	return mcp.NewTool("write_optimized_resume",
		mcp.WithDescription("Write the optimized resume to a versioned file. Call this when the user is satisfied with the optimized resume."),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("The optimized resume content in markdown format"),
		),
	)
}

// newErrorResult creates a tool result indicating an error.
func newErrorResult(message string) *mcp.CallToolResult {
	result := mcp.NewToolResultText(message)
	result.IsError = true
	return result
}

// WriteOptimizedResumeHandler creates a handler function for the write_optimized_resume tool.
// The handler writes the content to a versioned file in the application directory.
func WriteOptimizedResumeHandler(appDir string) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract content from arguments
		contentArg, ok := request.Params.Arguments["content"]
		if !ok {
			return newErrorResult("missing required parameter: content"), nil
		}

		content, ok := contentArg.(string)
		if !ok {
			return newErrorResult("content parameter must be a string"), nil
		}

		// Determine next version path
		outputPath, err := application.NextVersionPath(appDir)
		if err != nil {
			return newErrorResult(fmt.Sprintf("failed to determine output path: %v", err)), nil
		}

		// Write the file
		if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
			return newErrorResult(fmt.Sprintf("failed to write file: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Optimized resume written to: %s", outputPath)), nil
	}
}
