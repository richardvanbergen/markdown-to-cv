package generator

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/richq/m2cv/internal/executor"
)

// Exporter exports JSON Resume documents to PDF using resumed.
type Exporter struct {
	npxPath string
}

// NewExporter creates a new Exporter.
// It uses FindNodeExecutable to locate npx, supporting various Node.js version managers.
func NewExporter() (*Exporter, error) {
	npxPath, err := executor.FindNodeExecutable("npx")
	if err != nil {
		return nil, fmt.Errorf("npx not found: %w", err)
	}

	return &Exporter{npxPath: npxPath}, nil
}

// NewExporterWithOptions creates a new Exporter with custom FindOptions.
// This is useful for testing to ensure isolation from host system binaries.
func NewExporterWithOptions(opts *executor.FindOptions) (*Exporter, error) {
	npxPath, err := executor.FindNodeExecutableWithOptions("npx", opts)
	if err != nil {
		return nil, fmt.Errorf("npx not found: %w", err)
	}

	return &Exporter{npxPath: npxPath}, nil
}

// CheckThemeInstalled checks if a JSON Resume theme is installed in node_modules.
// Returns nil if the theme is installed, or an error with installation instructions.
func (e *Exporter) CheckThemeInstalled(projectDir, theme string) error {
	themePackage := "jsonresume-theme-" + theme
	themePath := filepath.Join(projectDir, "node_modules", themePackage)

	info, err := os.Stat(themePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("theme %q not installed. Run: npm install %s", theme, themePackage)
	}
	if err != nil {
		return fmt.Errorf("error checking theme %q: %w", theme, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("theme path exists but is not a directory: %s", themePath)
	}

	return nil
}

// ExportPDF exports a JSON Resume file to PDF using resumed.
//
// Parameters:
//   - ctx: context for cancellation
//   - jsonPath: path to the JSON Resume file to export
//   - outputPath: path for the output PDF file
//   - theme: JSON Resume theme name (e.g., "even", "stackoverflow")
//   - projectDir: project directory containing node_modules with resumed and theme
//
// The projectDir is critical - resumed resolves themes from node_modules relative to
// the working directory, so cmd.Dir must be set correctly.
func (e *Exporter) ExportPDF(ctx context.Context, jsonPath, outputPath, theme, projectDir string) error {
	// Validate theme is installed before attempting export
	if err := e.CheckThemeInstalled(projectDir, theme); err != nil {
		return err
	}

	// Build command: npx resumed export <jsonPath> --output <outputPath> --theme <themePackage>
	themePackage := "jsonresume-theme-" + theme
	args := []string{
		"resumed",
		"export",
		jsonPath,
		"--output", outputPath,
		"--theme", themePackage,
	}

	cmd := exec.CommandContext(ctx, e.npxPath, args...)
	cmd.Dir = projectDir // Critical for node_modules resolution

	// Use bytes.Buffer for stdout/stderr capture (consistent with ClaudeExecutor pattern)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Start the command (not cmd.Run() - follow existing pattern)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start resumed: %w", err)
	}

	// Wait for completion
	if err := cmd.Wait(); err != nil {
		stderrContent := strings.TrimSpace(stderr.String())
		if stderrContent != "" {
			return fmt.Errorf("resumed export failed: %w\nstderr: %s", err, stderrContent)
		}
		return fmt.Errorf("resumed export failed: %w", err)
	}

	return nil
}

// ValidateResumedInstalled checks if resumed is available in node_modules.
// Returns nil if installed, or an error with installation instructions.
func (e *Exporter) ValidateResumedInstalled(projectDir string) error {
	resumedPath := filepath.Join(projectDir, "node_modules", "resumed")
	info, err := os.Stat(resumedPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("resumed not installed. Run: npm install resumed")
	}
	if err != nil {
		return fmt.Errorf("error checking resumed installation: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("resumed path exists but is not a directory: %s", resumedPath)
	}
	return nil
}

// NPXPath returns the path to the npx executable.
// Useful for testing to verify the exporter was constructed correctly.
func (e *Exporter) NPXPath() string {
	return e.npxPath
}
