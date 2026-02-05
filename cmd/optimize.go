package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/richq/m2cv/internal/application"
	"github.com/richq/m2cv/internal/assets"
	"github.com/richq/m2cv/internal/config"
	"github.com/richq/m2cv/internal/executor"
	"github.com/richq/m2cv/internal/mcp"
	"github.com/spf13/cobra"
)

// interactivePromptTemplate is the system prompt for interactive mode.
const interactivePromptTemplate = `You are helping optimize a resume for a job application.

I've loaded:
- Application: %s

Your task:
1. Summarize the key requirements from the job description
2. Discuss optimization strategy with the user
3. When the user is satisfied, use the write_optimized_resume tool to save the final version

%s
---
BASE CV:
%s

---
JOB DESCRIPTION:
%s

---
Please start by summarizing the key requirements from the job description.`

// atsInstructions provides additional guidance for ATS optimization.
const atsInstructions = `ATS OPTIMIZATION MODE:
- Use standard section headings (Summary, Experience, Skills, Education)
- Include relevant keywords from the job description
- Avoid tables, columns, and complex formatting
- Use standard bullet points
- Keep formatting simple and parseable

`

// newOptimizeCommand creates the optimize subcommand.
func newOptimizeCommand() *cobra.Command {
	var (
		model       string
		atsMode     bool
		interactive bool
	)

	cmd := &cobra.Command{
		Use:   "optimize <application-name>",
		Short: "Tailor CV to job description with AI",
		Long: `Tailor your base CV to a specific job description using Claude AI.

The command reads your base CV and the job description from the application
folder, then uses Claude to produce a tailored version optimized for that role.

Use --ats flag for ATS (Applicant Tracking Systems) optimization which uses
standard section headings and includes keywords from the job description.

Use --interactive flag to launch Claude in conversation mode where you can
discuss the optimization strategy before generating the final resume.

Output is written to a versioned file (optimized-cv-N.md) in the application folder.

Examples:
  m2cv optimize acme-software-engineer
  m2cv optimize --ats google-sre
  m2cv optimize --interactive my-dream-job
  m2cv optimize -m claude-sonnet-4-20250514 my-dream-job`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if interactive {
				return runOptimizeInteractive(cmd.Context(), args[0], model, atsMode)
			}
			return runOptimize(cmd.Context(), args[0], model, atsMode)
		},
	}

	cmd.Flags().StringVarP(&model, "model", "m", "", "override Claude model")
	cmd.Flags().BoolVar(&atsMode, "ats", false, "optimize for ATS (Applicant Tracking Systems)")
	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "launch Claude in conversation mode")

	return cmd
}

// runOptimize executes the optimize command logic.
func runOptimize(ctx context.Context, applicationName, modelOverride string, atsMode bool) error {
	// Validate application folder exists
	appDir := filepath.Join("applications", applicationName)
	if _, err := os.Stat(appDir); os.IsNotExist(err) {
		return fmt.Errorf("application folder not found: %s. Run 'm2cv apply' first", appDir)
	}

	// Load config (required for base CV path)
	configPath, err := config.FindWithOverrides(cfgFile, ".")
	if err != nil {
		return fmt.Errorf("m2cv.yml not found: %w. Run 'm2cv init' first", err)
	}

	configRepo := config.NewRepository()
	cfg, err := configRepo.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Resolve and read base CV
	cvPath := cfg.BaseCVPath
	if baseCVPath != "" {
		// Persistent flag override
		cvPath = baseCVPath
	}

	// Resolve relative paths against config directory
	if !filepath.IsAbs(cvPath) {
		configDir := filepath.Dir(configPath)
		cvPath = filepath.Join(configDir, cvPath)
	}

	baseCV, err := os.ReadFile(cvPath)
	if err != nil {
		return fmt.Errorf("failed to read base CV at %s: %w", cvPath, err)
	}

	// Find and read job description
	txtFiles, err := filepath.Glob(filepath.Join(appDir, "*.txt"))
	if err != nil {
		return fmt.Errorf("failed to search for job description: %w", err)
	}
	if len(txtFiles) == 0 {
		return fmt.Errorf("no .txt file found in %s. Job description required", appDir)
	}

	jobDescription, err := os.ReadFile(txtFiles[0])
	if err != nil {
		return fmt.Errorf("failed to read job description at %s: %w", txtFiles[0], err)
	}

	// Select and build prompt
	promptName := "optimize"
	if atsMode {
		promptName = "optimize-ats"
	}

	promptTemplate, err := assets.GetPrompt(promptName)
	if err != nil {
		return fmt.Errorf("failed to load prompt template: %w", err)
	}

	prompt := strings.ReplaceAll(promptTemplate, "{{.BaseCV}}", string(baseCV))
	prompt = strings.ReplaceAll(prompt, "{{.JobDescription}}", string(jobDescription))

	// Determine model
	model := cfg.DefaultModel
	if modelOverride != "" {
		model = modelOverride
	}

	// Execute Claude
	exec := executor.NewClaudeExecutor()
	var opts []executor.ExecuteOption
	if model != "" {
		opts = append(opts, executor.WithModel(model))
	}

	result, err := exec.Execute(ctx, prompt, opts...)
	if err != nil {
		return fmt.Errorf("failed to optimize CV: %w", err)
	}

	// Write versioned output
	outputPath, err := application.NextVersionPath(appDir)
	if err != nil {
		return fmt.Errorf("failed to determine output path: %w", err)
	}

	if err := os.WriteFile(outputPath, []byte(result), 0644); err != nil {
		return fmt.Errorf("failed to write optimized CV: %w", err)
	}

	fmt.Printf("Optimized CV written to: %s\n", outputPath)
	return nil
}

// mcpConfig represents the MCP configuration JSON structure.
type mcpConfig struct {
	MCPServers map[string]mcpServerConfig `json:"mcpServers"`
}

// mcpServerConfig represents a single MCP server configuration.
type mcpServerConfig struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

// runOptimizeInteractive runs the optimize command in interactive mode.
func runOptimizeInteractive(ctx context.Context, applicationName, modelOverride string, atsMode bool) error {
	// Validate application folder exists
	appDir := filepath.Join("applications", applicationName)
	if _, err := os.Stat(appDir); os.IsNotExist(err) {
		return fmt.Errorf("application folder not found: %s. Run 'm2cv apply' first", appDir)
	}

	// Load config (required for base CV path)
	configPath, err := config.FindWithOverrides(cfgFile, ".")
	if err != nil {
		return fmt.Errorf("m2cv.yml not found: %w. Run 'm2cv init' first", err)
	}

	configRepo := config.NewRepository()
	cfg, err := configRepo.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Resolve and read base CV
	cvPath := cfg.BaseCVPath
	if baseCVPath != "" {
		cvPath = baseCVPath
	}
	if !filepath.IsAbs(cvPath) {
		configDir := filepath.Dir(configPath)
		cvPath = filepath.Join(configDir, cvPath)
	}

	baseCV, err := os.ReadFile(cvPath)
	if err != nil {
		return fmt.Errorf("failed to read base CV at %s: %w", cvPath, err)
	}

	// Find and read job description
	txtFiles, err := filepath.Glob(filepath.Join(appDir, "*.txt"))
	if err != nil {
		return fmt.Errorf("failed to search for job description: %w", err)
	}
	if len(txtFiles) == 0 {
		return fmt.Errorf("no .txt file found in %s. Job description required", appDir)
	}

	jobDescription, err := os.ReadFile(txtFiles[0])
	if err != nil {
		return fmt.Errorf("failed to read job description at %s: %w", txtFiles[0], err)
	}

	// Determine model
	model := cfg.DefaultModel
	if modelOverride != "" {
		model = modelOverride
	}

	// Create context for MCP subprocess
	mcpCtx := &mcp.InteractiveContext{
		ApplicationDir: appDir,
		BaseCV:         string(baseCV),
		JobDescription: string(jobDescription),
		ATSMode:        atsMode,
		Model:          model,
	}

	encodedContext, err := mcpCtx.Encode()
	if err != nil {
		return fmt.Errorf("failed to encode context: %w", err)
	}

	// Get our own executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Create temporary MCP config file
	mcpCfg := mcpConfig{
		MCPServers: map[string]mcpServerConfig{
			"m2cv": {
				Command: execPath,
				Args:    []string{"mcp", "--context", encodedContext},
			},
		},
	}

	mcpConfigJSON, err := json.Marshal(mcpCfg)
	if err != nil {
		return fmt.Errorf("failed to marshal MCP config: %w", err)
	}

	tmpFile, err := os.CreateTemp("", "m2cv-mcp-*.json")
	if err != nil {
		return fmt.Errorf("failed to create temp MCP config: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(mcpConfigJSON); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write MCP config: %w", err)
	}
	tmpFile.Close()

	// Build system prompt
	atsText := ""
	if atsMode {
		atsText = atsInstructions
	}
	systemPrompt := fmt.Sprintf(interactivePromptTemplate, applicationName, atsText, string(baseCV), string(jobDescription))

	// Execute Claude interactively
	exec := executor.NewClaudeExecutor()
	interactiveCfg := executor.InteractiveConfig{
		MCPConfigPath: tmpFile.Name(),
		SystemPrompt:  systemPrompt,
		Model:         model,
	}

	return exec.ExecuteInteractive(ctx, interactiveCfg)
}
