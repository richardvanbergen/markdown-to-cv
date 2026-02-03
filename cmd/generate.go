package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/richq/m2cv/internal/application"
	"github.com/richq/m2cv/internal/assets"
	"github.com/richq/m2cv/internal/config"
	"github.com/richq/m2cv/internal/executor"
	"github.com/richq/m2cv/internal/generator"
	"github.com/richq/m2cv/internal/preflight"
	"github.com/spf13/cobra"
)

// newGenerateCommand creates the generate subcommand.
func newGenerateCommand() *cobra.Command {
	var (
		theme string
		model string
	)

	cmd := &cobra.Command{
		Use:   "generate <application-name>",
		Short: "Generate PDF resume from optimized CV",
		Long: `Generate a PDF resume from an optimized CV using Claude AI and resumed.

The command reads the latest optimized CV from the application folder,
converts it to JSON Resume format via Claude, validates the schema,
and exports a professionally themed PDF using resumed.

The --theme flag overrides the default theme from config.
The -m/--model flag overrides the default Claude model from config.

Output files written to the application folder:
  - resume.json (intermediate, useful for debugging)
  - resume.pdf (final output)

Examples:
  m2cv generate acme-software-engineer
  m2cv generate --theme stackoverflow my-app
  m2cv generate -m claude-sonnet-4-20250514 my-dream-job`,
		Args: cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Find project directory for resumed check
			configPath, err := config.FindWithOverrides(cfgFile, ".")
			if err != nil {
				// Config not found - will be reported in RunE, skip preflight
				return nil
			}
			projectDir := filepath.Dir(configPath)
			return preflight.CheckResumed(projectDir)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenerate(cmd.Context(), args[0], theme, model)
		},
	}

	cmd.Flags().StringVar(&theme, "theme", "", "override JSON Resume theme")
	cmd.Flags().StringVarP(&model, "model", "m", "", "override Claude model")

	return cmd
}

// runGenerate executes the generate command logic.
func runGenerate(ctx context.Context, applicationName, themeOverride, modelOverride string) error {
	// 1. Validate application folder exists
	appDir := filepath.Join("applications", applicationName)
	if _, err := os.Stat(appDir); os.IsNotExist(err) {
		return fmt.Errorf("application folder not found: %s. Run 'm2cv apply' first", appDir)
	}

	// 2. Load config
	configPath, err := config.FindWithOverrides(cfgFile, ".")
	if err != nil {
		return fmt.Errorf("m2cv.yml not found: %w. Run 'm2cv init' first", err)
	}

	configRepo := config.NewRepository()
	cfg, err := configRepo.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// 3. Determine theme: flag > config.DefaultTheme
	theme := cfg.DefaultTheme
	if themeOverride != "" {
		theme = themeOverride
	}
	if theme == "" {
		theme = "even" // Fallback default
	}

	// 4. Determine model: flag > config.DefaultModel
	model := cfg.DefaultModel
	if modelOverride != "" {
		model = modelOverride
	}

	// 5. Find latest optimized CV
	latestCVPath, err := application.LatestVersionPath(appDir)
	if err != nil {
		return fmt.Errorf("failed to find optimized CV: %w", err)
	}
	if latestCVPath == "" {
		return fmt.Errorf("no optimized CV found in %s. Run 'm2cv optimize %s' first", appDir, applicationName)
	}

	// 6. Read CV content
	cvContent, err := os.ReadFile(latestCVPath)
	if err != nil {
		return fmt.Errorf("failed to read optimized CV at %s: %w", latestCVPath, err)
	}

	// 7. Load md-to-json-resume prompt
	promptTemplate, err := assets.GetPrompt("md-to-json-resume")
	if err != nil {
		return fmt.Errorf("failed to load prompt template: %w", err)
	}

	// 8. Substitute {{.CV}} with content
	prompt := strings.ReplaceAll(promptTemplate, "{{.CV}}", string(cvContent))

	// 9. Execute Claude
	exec := executor.NewClaudeExecutor()
	var opts []executor.ExecuteOption
	if model != "" {
		opts = append(opts, executor.WithModel(model))
	}

	result, err := exec.Execute(ctx, prompt, opts...)
	if err != nil {
		return fmt.Errorf("failed to convert CV to JSON Resume: %w", err)
	}

	// 10. Extract JSON from Claude output
	jsonResume, err := generator.ExtractJSON([]byte(result))
	if err != nil {
		return fmt.Errorf("failed to extract JSON from Claude output: %w", err)
	}

	// 11. Validate against JSON Resume schema
	validator, err := generator.NewValidator()
	if err != nil {
		return fmt.Errorf("failed to initialize validator: %w", err)
	}

	if err := validator.Validate(jsonResume); err != nil {
		return fmt.Errorf("JSON Resume validation failed: %w. Try running 'm2cv generate' again or check the optimized CV", err)
	}

	// 12. Write resume.json to appDir (for debugging)
	jsonPath := filepath.Join(appDir, "resume.json")
	if err := os.WriteFile(jsonPath, jsonResume, 0644); err != nil {
		return fmt.Errorf("failed to write resume.json: %w", err)
	}

	// 13. Export PDF via resumed
	projectDir := filepath.Dir(configPath)
	pdfPath := filepath.Join(appDir, "resume.pdf")

	exporter, err := generator.NewExporter()
	if err != nil {
		return fmt.Errorf("failed to initialize exporter: %w", err)
	}

	if err := exporter.ExportPDF(ctx, jsonPath, pdfPath, theme, projectDir); err != nil {
		return fmt.Errorf("failed to export PDF: %w", err)
	}

	// 14. Print success
	fmt.Printf("JSON written to: %s\n", jsonPath)
	fmt.Printf("PDF written to: %s\n", pdfPath)

	return nil
}
