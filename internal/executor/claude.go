package executor

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// ClaudeExecutor executes Claude CLI commands for AI-powered text generation.
// It uses stdin for prompt input (avoiding shell argument limits) and
// bytes.Buffer for output capture (avoiding deadlocks with large output).
type ClaudeExecutor interface {
	// Execute runs claude with the given prompt and returns the result.
	// Options can modify the command (e.g., WithModel, WithOutputFormat).
	Execute(ctx context.Context, prompt string, opts ...ExecuteOption) (string, error)
}

// claudeExecutor is the default implementation of ClaudeExecutor.
type claudeExecutor struct {
	claudePath string
}

// ExecuteOption modifies the behavior of Execute.
type ExecuteOption func(*executeConfig)

// executeConfig holds configuration for a single Execute call.
type executeConfig struct {
	model        string
	outputFormat string
}

// NewClaudeExecutor creates a new ClaudeExecutor.
// Use WithClaudePath to specify a custom claude binary location.
func NewClaudeExecutor(opts ...ClaudeOption) ClaudeExecutor {
	e := &claudeExecutor{
		claudePath: "claude", // default to PATH lookup
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// ClaudeOption modifies the ClaudeExecutor construction.
type ClaudeOption func(*claudeExecutor)

// WithClaudePath sets a custom path to the claude binary.
func WithClaudePath(path string) ClaudeOption {
	return func(e *claudeExecutor) {
		e.claudePath = path
	}
}

// WithModel sets the model to use for execution.
func WithModel(model string) ExecuteOption {
	return func(c *executeConfig) {
		c.model = model
	}
}

// WithOutputFormat sets the output format (text, json, etc.).
func WithOutputFormat(format string) ExecuteOption {
	return func(c *executeConfig) {
		c.outputFormat = format
	}
}

// Execute runs claude with the given prompt.
// Prompts are passed via stdin to avoid shell argument length limits.
// Output is captured using bytes.Buffer to avoid deadlocks with large output.
//
// By default, uses:
//   - -p flag (print mode)
//   - --output-format text (plain text output)
//
// Use WithModel and WithOutputFormat to customize behavior.
func (e *claudeExecutor) Execute(ctx context.Context, prompt string, opts ...ExecuteOption) (string, error) {
	// Apply options
	cfg := &executeConfig{
		outputFormat: "text", // default
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// Build command arguments
	args := []string{"-p", "--output-format", cfg.outputFormat}
	if cfg.model != "" {
		args = append(args, "--model", cfg.model)
	}

	// Create command with context for cancellation support
	cmd := exec.CommandContext(ctx, e.claudePath, args...)

	// Pass prompt via stdin (Pattern 2: stdin piping for large prompts)
	cmd.Stdin = strings.NewReader(prompt)

	// Use bytes.Buffer for stdout/stderr (Pattern 1: streaming subprocess execution)
	// This avoids deadlocks that can occur with cmd.Output() when buffers fill
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Start the command (don't use cmd.Run() or cmd.Output())
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start claude: %w (not found or not executable)", err)
	}

	// Wait for completion
	if err := cmd.Wait(); err != nil {
		// Include stderr in error message for debugging
		stderrContent := strings.TrimSpace(stderr.String())
		if stderrContent != "" {
			return "", fmt.Errorf("claude execution failed: %w\nstderr: %s", err, stderrContent)
		}
		return "", fmt.Errorf("claude execution failed: %w", err)
	}

	return stdout.String(), nil
}
