# Phase 5: Export Pipeline - Research

**Researched:** 2026-02-03
**Domain:** JSON Resume conversion, JSON Schema validation, PDF export via resumed
**Confidence:** HIGH

## Summary

Phase 5 implements the `m2cv generate` command which converts optimized markdown CVs to JSON Resume format via Claude, validates against JSON Resume schema, and exports PDFs using resumed with theme selection. This phase completes the end-to-end pipeline from job description to professional PDF.

The standard approach involves three sequential steps: (1) Claude CLI conversion from markdown to JSON Resume format with robust JSON extraction, (2) JSON Schema validation using santhosh-tekuri/jsonschema v6 with the embedded resume.schema.json, and (3) PDF export via resumed using npx with theme selection. The critical technical challenges are extracting valid JSON from Claude's potentially verbose output, handling validation errors gracefully, and properly invoking resumed with correct working directory and theme parameters.

**Primary recommendation:** Implement a robust JSON extractor that strips markdown fences and finds JSON object boundaries (handling Claude's tendency to add explanatory text). Use jsonschema v6's Compiler API with the embedded schema for validation. Execute resumed via npx (not direct binary) to ensure proper Node.js environment and theme resolution. Follow the existing executor pattern (NPMExecutor) for subprocess invocation with proper error handling.

## Standard Stack

The established libraries/tools for this domain:

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| github.com/santhosh-tekuri/jsonschema | v6.0.2+ | JSON Schema validation | Supports draft-07 (used by JSON Resume), fast validation, excellent error messages, HIGH reputation in Context7 |
| encoding/json | stdlib | JSON parsing and unmarshaling | Built-in Go JSON handling, sufficient for validation input/output |
| resumed | 6.1.0+ | PDF export from JSON Resume | Actively maintained fork of resume-cli, ~400 JSON Resume themes available, lightweight |
| regexp | stdlib | JSON extraction from text | For stripping markdown fences and finding JSON boundaries |

**Confidence:** HIGH - jsonschema v6 verified via pkg.go.dev (released May 2025), resumed 6.1.0 confirmed active on npm (Oct 2025).

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| bytes | stdlib | Buffer management | For JSON extraction and manipulation |
| strings | stdlib | String manipulation | For prompt substitution (existing pattern) |
| os/exec | stdlib | Subprocess execution | For invoking resumed via npx |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| jsonschema v6 | gojsonschema | jsonschema v6 has better draft-07 support, faster, more actively maintained |
| resumed | resume-cli | resume-cli unmaintained since 2017; resumed is active fork |
| resumed | @jsonresume/cli | @jsonresume/cli is newer but less stable; resumed is proven |
| Direct `resumed` binary | npx resumed | npx ensures correct Node.js environment, theme resolution from node_modules |

**Installation:**
```bash
# Go dependency (add to go.mod)
go get github.com/santhosh-tekuri/jsonschema/v6

# External tool (installed during m2cv init)
npm install resumed jsonresume-theme-{name}
```

## Architecture Patterns

### Recommended Project Structure
```
cmd/
├── generate.go           # Command definition with --theme flag
└── generate_test.go      # Command integration tests

internal/
├── generator/
│   ├── converter.go      # MD -> JSON Resume via Claude
│   ├── converter_test.go
│   ├── validator.go      # JSON Schema validation
│   ├── validator_test.go
│   ├── extractor.go      # Extract JSON from Claude output
│   └── extractor_test.go
├── executor/
│   └── npm.go            # Already exists - for running resumed
└── assets/
    ├── prompts/
    │   └── md-to-json-resume.txt  # Already exists
    └── schema/
        └── resume.schema.json      # Already exists (JSON Resume draft-07)
```

### Pattern 1: JSON Extraction from LLM Output (CRITICAL)

**What:** Extract valid JSON from Claude output that may contain markdown fences, explanatory text, or comments.

**When to use:** All Claude invocations that expect JSON output (md-to-json-resume conversion).

**Why critical:** LLMs are non-deterministic. Claude may return:
- Markdown code fences: `` ```json\n{...}\n``` ``
- Explanatory text before/after JSON
- Comments (not valid JSON)
- Mixed format responses

**Example:**
```go
// internal/generator/extractor.go
package generator

import (
    "bytes"
    "encoding/json"
    "fmt"
    "regexp"
    "strings"
)

// ExtractJSON extracts a JSON object from Claude output that may contain
// markdown fences, explanatory text, or other non-JSON content.
func ExtractJSON(claudeOutput []byte) (json.RawMessage, error) {
    // 1. Strip markdown code fences
    content := stripMarkdownFences(claudeOutput)

    // 2. Find JSON object boundaries
    start := bytes.IndexByte(content, '{')
    end := bytes.LastIndexByte(content, '}')

    if start == -1 || end == -1 || start > end {
        return nil, fmt.Errorf("no JSON object found in output:\n%s", claudeOutput)
    }

    extracted := content[start : end+1]

    // 3. Validate it's parseable JSON
    var raw json.RawMessage
    if err := json.Unmarshal(extracted, &raw); err != nil {
        return nil, fmt.Errorf("extracted content is not valid JSON: %w\nExtracted:\n%s", err, extracted)
    }

    return raw, nil
}

// stripMarkdownFences removes ```json...``` or ```...``` fences
func stripMarkdownFences(data []byte) []byte {
    // Match ```json or ``` followed by content and closing ```
    re := regexp.MustCompile("(?s)```(?:json)?\n?(.*?)\n?```")
    matches := re.FindSubmatch(data)
    if matches != nil && len(matches) > 1 {
        return matches[1]
    }
    return data
}

// ExtractJSONArray handles array responses (if needed for future features)
func ExtractJSONArray(claudeOutput []byte) (json.RawMessage, error) {
    content := stripMarkdownFences(claudeOutput)

    start := bytes.IndexByte(content, '[')
    end := bytes.LastIndexByte(content, ']')

    if start == -1 || end == -1 || start > end {
        return nil, fmt.Errorf("no JSON array found in output")
    }

    extracted := content[start : end+1]

    var raw json.RawMessage
    if err := json.Unmarshal(extracted, &raw); err != nil {
        return nil, fmt.Errorf("extracted content is not valid JSON array: %w", err)
    }

    return raw, nil
}
```

**Sources:**
- [How to Consistently Retrieve Valid JSON from Claude 3.5 in Go](https://dev.to/embiem/how-to-consistently-retrieve-valid-json-from-claude-35-in-go-1g5b)
- [Claude API Structured Outputs](https://platform.claude.com/docs/en/build-with-claude/structured-outputs)

**Confidence:** HIGH - Well-documented pattern for LLM JSON extraction.

---

### Pattern 2: JSON Schema Validation with jsonschema v6

**What:** Validate JSON Resume output against JSON Schema draft-07 before PDF export.

**When to use:** After Claude conversion, before calling resumed.

**Why:** Catch schema violations early with clear error messages instead of cryptic resumed failures.

**Example:**
```go
// internal/generator/validator.go
package generator

import (
    "encoding/json"
    "fmt"

    "github.com/richq/m2cv/internal/assets"
    "github.com/santhosh-tekuri/jsonschema/v6"
)

// Validator validates JSON Resume documents against the JSON Resume schema.
type Validator struct {
    schema *jsonschema.Schema
}

// NewValidator creates a new Validator with the embedded JSON Resume schema.
func NewValidator() (*Validator, error) {
    // Load embedded schema
    schemaData, err := assets.GetSchema("resume.schema.json")
    if err != nil {
        return nil, fmt.Errorf("failed to load schema: %w", err)
    }

    // Parse schema
    var schemaObj interface{}
    if err := json.Unmarshal(schemaData, &schemaObj); err != nil {
        return nil, fmt.Errorf("failed to parse schema: %w", err)
    }

    // Compile schema
    compiler := jsonschema.NewCompiler()
    if err := compiler.AddResource("resume.schema.json", schemaObj); err != nil {
        return nil, fmt.Errorf("failed to add schema resource: %w", err)
    }

    schema, err := compiler.Compile("resume.schema.json")
    if err != nil {
        return nil, fmt.Errorf("failed to compile schema: %w", err)
    }

    return &Validator{schema: schema}, nil
}

// Validate checks if the JSON Resume document is valid according to the schema.
// Returns nil if valid, or an error describing validation failures.
func (v *Validator) Validate(resumeJSON []byte) error {
    // Parse JSON document
    var doc interface{}
    if err := json.Unmarshal(resumeJSON, &doc); err != nil {
        return fmt.Errorf("invalid JSON: %w", err)
    }

    // Validate against schema
    if err := v.schema.Validate(doc); err != nil {
        // jsonschema returns detailed validation errors
        return fmt.Errorf("schema validation failed:\n%w", err)
    }

    return nil
}

// ValidateWithDetails returns structured validation errors for better UX.
func (v *Validator) ValidateWithDetails(resumeJSON []byte) (*ValidationResult, error) {
    var doc interface{}
    if err := json.Unmarshal(resumeJSON, &doc); err != nil {
        return nil, fmt.Errorf("invalid JSON: %w", err)
    }

    err := v.schema.Validate(doc)

    result := &ValidationResult{
        Valid: err == nil,
    }

    if err != nil {
        // Extract validation errors
        if valErr, ok := err.(*jsonschema.ValidationError); ok {
            result.Errors = formatValidationErrors(valErr)
        } else {
            result.Errors = []string{err.Error()}
        }
    }

    return result, nil
}

// ValidationResult holds the outcome of schema validation.
type ValidationResult struct {
    Valid  bool
    Errors []string
}

// formatValidationErrors converts jsonschema.ValidationError to readable strings.
func formatValidationErrors(err *jsonschema.ValidationError) []string {
    var errors []string

    // Basic error message
    errors = append(errors, err.Error())

    // Add causes if present
    for _, cause := range err.Causes {
        errors = append(errors, "  - "+cause.Error())
    }

    return errors
}
```

**Sources:**
- [jsonschema v6 package documentation](https://pkg.go.dev/github.com/santhosh-tekuri/jsonschema/v6)
- [A JSON schema package for Go - Google Open Source Blog](https://opensource.googleblog.com/2026/01/a-json-schema-package-for-go.html)

**Confidence:** HIGH - Official package documentation, recent blog post (Jan 2026).

---

### Pattern 3: PDF Export via resumed with Theme Selection

**What:** Execute resumed via npx to export JSON Resume to PDF with specified theme.

**When to use:** Final step after validation passes.

**Why:** npx ensures correct Node.js environment, resolves themes from node_modules, handles PATH issues.

**Example:**
```go
// internal/generator/exporter.go
package generator

import (
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
func NewExporter() (*Exporter, error) {
    // Use FindNodeExecutable to locate npx (handles version managers)
    npxPath, err := executor.FindNodeExecutable("npx")
    if err != nil {
        return nil, fmt.Errorf("npx not found: %w", err)
    }

    return &Exporter{npxPath: npxPath}, nil
}

// ExportPDF exports a JSON Resume file to PDF using resumed.
// Parameters:
//   - ctx: context for cancellation
//   - jsonPath: path to JSON Resume file
//   - outputPath: path for output PDF
//   - theme: JSON Resume theme name (e.g., "even", "stackoverflow")
//   - projectDir: project directory containing node_modules with resumed + theme
func (e *Exporter) ExportPDF(ctx context.Context, jsonPath, outputPath, theme, projectDir string) error {
    // Validate theme is installed
    themePackage := "jsonresume-theme-" + theme
    themePath := filepath.Join(projectDir, "node_modules", themePackage)
    if _, err := os.Stat(themePath); os.IsNotExist(err) {
        return fmt.Errorf("theme %q not installed. Run: npm install %s", theme, themePackage)
    }

    // Build command: npx resumed export <jsonPath> --output <outputPath> --theme <theme>
    args := []string{
        "resumed",
        "export",
        jsonPath,
        "--output", outputPath,
        "--theme", themePackage,
    }

    cmd := exec.CommandContext(ctx, e.npxPath, args...)
    cmd.Dir = projectDir // Set working directory for node_modules resolution

    // Capture output for error messages
    var stdout, stderr strings.Builder
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    // Execute
    if err := cmd.Start(); err != nil {
        return fmt.Errorf("failed to start resumed: %w", err)
    }

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
func (e *Exporter) ValidateResumedInstalled(projectDir string) error {
    resumedPath := filepath.Join(projectDir, "node_modules", "resumed")
    if _, err := os.Stat(resumedPath); os.IsNotExist(err) {
        return fmt.Errorf("resumed not installed. Run: npm install resumed")
    }
    return nil
}
```

**Command line equivalent:**
```bash
# From project directory with node_modules
npx resumed export resume.json --output resume.pdf --theme jsonresume-theme-even
```

**Sources:**
- [resumed npm package](https://www.npmjs.com/package/resumed)
- [resumed GitHub repository](https://github.com/rbardini/resumed)
- Existing codebase pattern: internal/executor/npm.go

**Confidence:** HIGH - Verified via npm registry, existing NPMExecutor pattern.

---

### Pattern 4: End-to-End Generate Command Flow

**What:** Orchestrate the three-step pipeline in the generate command.

**When to use:** cmd/generate.go implementation.

**Example:**
```go
// cmd/generate.go
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
    "github.com/spf13/cobra"
)

func newGenerateCommand() *cobra.Command {
    var (
        themeFlag string
        modelFlag string
    )

    cmd := &cobra.Command{
        Use:   "generate <application-name>",
        Short: "Generate PDF resume from optimized CV",
        Args:  cobra.ExactArgs(1),
        PreRunE: func(cmd *cobra.Command, args []string) error {
            // Check resumed is installed (command-specific preflight)
            projectDir := "." // or from config
            return preflight.CheckResumed(projectDir)
        },
        RunE: func(cmd *cobra.Command, args []string) error {
            return runGenerate(cmd.Context(), args[0], themeFlag, modelFlag)
        },
    }

    cmd.Flags().StringVar(&themeFlag, "theme", "", "JSON Resume theme (overrides config)")
    cmd.Flags().StringVarP(&modelFlag, "model", "m", "", "Claude model (overrides config)")

    return cmd
}

func runGenerate(ctx context.Context, appName, themeFlag, modelFlag string) error {
    // 1. Load config
    configPath, err := config.FindWithOverrides("", ".")
    if err != nil {
        return fmt.Errorf("m2cv.yml not found: %w. Run 'm2cv init' first", err)
    }

    configRepo := config.NewRepository()
    cfg, err := configRepo.Load(configPath)
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }

    // 2. Determine theme (flag > config)
    theme := cfg.DefaultTheme
    if themeFlag != "" {
        theme = themeFlag
    }

    // 3. Determine model (flag > config)
    model := cfg.DefaultModel
    if modelFlag != "" {
        model = modelFlag
    }

    // 4. Find latest optimized CV
    appDir := filepath.Join("applications", appName)
    versions, err := application.ListVersions(appDir)
    if err != nil {
        return fmt.Errorf("failed to list versions: %w", err)
    }
    if len(versions) == 0 {
        return fmt.Errorf("no optimized CV found in %s. Run 'm2cv optimize %s' first", appDir, appName)
    }

    latestVersion := versions[len(versions)-1]
    cvPath := filepath.Join(appDir, latestVersion.Filename)

    // 5. Read optimized CV
    cvContent, err := os.ReadFile(cvPath)
    if err != nil {
        return fmt.Errorf("failed to read CV: %w", err)
    }

    // 6. Convert MD -> JSON Resume via Claude
    fmt.Println("Converting CV to JSON Resume format...")
    claudeExec := executor.NewClaudeExecutor()

    promptTemplate, err := assets.GetPrompt("md-to-json-resume")
    if err != nil {
        return fmt.Errorf("failed to load prompt: %w", err)
    }

    prompt := strings.ReplaceAll(promptTemplate, "{{.CV}}", string(cvContent))

    var executeOpts []executor.ExecuteOption
    if model != "" {
        executeOpts = append(executeOpts, executor.WithModel(model))
    }

    claudeOutput, err := claudeExec.Execute(ctx, prompt, executeOpts...)
    if err != nil {
        return fmt.Errorf("Claude execution failed: %w", err)
    }

    // 7. Extract JSON from Claude output
    resumeJSON, err := generator.ExtractJSON([]byte(claudeOutput))
    if err != nil {
        return fmt.Errorf("failed to extract JSON: %w", err)
    }

    // 8. Validate against JSON Resume schema
    fmt.Println("Validating JSON Resume schema...")
    validator, err := generator.NewValidator()
    if err != nil {
        return fmt.Errorf("failed to create validator: %w", err)
    }

    if err := validator.Validate(resumeJSON); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }

    // 9. Write JSON to temp file
    jsonPath := filepath.Join(appDir, "resume.json")
    if err := os.WriteFile(jsonPath, resumeJSON, 0644); err != nil {
        return fmt.Errorf("failed to write JSON: %w", err)
    }

    // 10. Export PDF via resumed
    fmt.Printf("Exporting PDF with theme: %s...\n", theme)
    exporter, err := generator.NewExporter()
    if err != nil {
        return fmt.Errorf("failed to create exporter: %w", err)
    }

    pdfPath := filepath.Join(appDir, "resume.pdf")
    projectDir := filepath.Dir(configPath) // Project root

    if err := exporter.ExportPDF(ctx, jsonPath, pdfPath, theme, projectDir); err != nil {
        return fmt.Errorf("PDF export failed: %w", err)
    }

    fmt.Printf("Success! PDF written to: %s\n", pdfPath)
    return nil
}
```

**Confidence:** HIGH - Follows existing command patterns from apply.go and optimize.go.

---

### Anti-Patterns to Avoid

- **Direct JSON unmarshaling of Claude output:** Always use ExtractJSON first; Claude may add markdown fences or explanatory text
- **Skipping schema validation:** Validation catches errors before PDF export; resumed errors are cryptic
- **Using global resumed install only:** Local node_modules ensures theme availability; always prefer local install
- **Ignoring validation error details:** jsonschema v6 provides rich error messages; surface them to users
- **Not setting working directory for resumed:** Theme resolution requires correct working directory with node_modules
- **Using cmd.Output() for resumed:** Follow existing pattern with cmd.Start() + cmd.Wait() + bytes.Buffer

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| JSON extraction from LLM output | Custom string parsing | ExtractJSON with regex + boundaries | Handles markdown fences, arrays vs objects, validation |
| JSON Schema validation | Custom field checks | jsonschema v6 | Supports draft-07, comprehensive error messages, handles $ref, allOf, anyOf |
| PDF generation | Custom PDF library | resumed + JSON Resume themes | 400+ themes, actively maintained, handles layout/styling |
| Theme management | Custom CSS/HTML | JSON Resume theme ecosystem | Community-maintained, proven designs, easy to switch |
| Node.js environment setup | Custom PATH manipulation | npx for execution | Handles version managers, resolves node_modules correctly |

**Key insight:** The complexity in this phase is orchestration and error handling, not the individual operations. Use established tools for conversion (Claude), validation (jsonschema), and export (resumed).

## Common Pitfalls

### Pitfall 1: Unmarshaling Claude Output Without Extraction (CRITICAL)

**What goes wrong:** Directly calling `json.Unmarshal()` on Claude output fails intermittently with "invalid character" errors.

**Why it happens:** LLMs are non-deterministic. Claude may return:
- `` ```json\n{...}\n``` `` (markdown fences)
- Text + JSON mixed
- Comments or explanatory text

**How to avoid:** Always use `ExtractJSON()` before unmarshaling. This strips fences and finds JSON boundaries.

**Warning signs:**
- Intermittent "invalid character 'H'" errors (from "Here is...")
- Different errors on retry with same input
- Works in testing but fails in production

**Sources:**
- [How to Consistently Retrieve Valid JSON from Claude 3.5 in Go](https://dev.to/embiem/how-to-consistently-retrieve-valid-json-from-claude-35-in-go-1g5b)
- Phase 1 Research: Pitfall 5 (No JSON Extraction from Claude Output)

**Confidence:** HIGH - Known LLM integration challenge documented in multiple sources.

---

### Pitfall 2: Schema Validation After PDF Export

**What goes wrong:** resumed fails with cryptic errors like "Cannot read property 'name' of undefined".

**Why it happens:** Invalid JSON Resume passed to resumed. The schema violation isn't caught until resumed tries to render.

**How to avoid:** Always validate with `validator.Validate()` before calling `exporter.ExportPDF()`.

**Warning signs:**
- resumed errors don't mention which field is missing
- PDF generation succeeds with incomplete data
- Different themes fail differently

**Confidence:** HIGH - Fail-fast validation principle, proven in Phase 1-4.

---

### Pitfall 3: Running resumed from Wrong Working Directory

**What goes wrong:** resumed reports "Theme not found" even though `npm install jsonresume-theme-even` succeeded.

**Why it happens:** resumed looks for themes in `node_modules` relative to working directory. If cmd.Dir is not set, it uses parent process directory.

**How to avoid:** Always set `cmd.Dir = projectDir` where projectDir contains node_modules.

**Warning signs:**
- Theme installation succeeds but export fails
- Works with globally installed themes, fails with local
- Different results when running from different directories

**Sources:**
- Existing codebase: internal/executor/npm.go line 104 (cmd.Dir pattern)

**Confidence:** HIGH - Standard subprocess working directory issue.

---

### Pitfall 4: Missing Theme Installation Check

**What goes wrong:** User runs `m2cv generate --theme stackoverflow` but theme isn't installed. resumed fails with "Cannot find module 'jsonresume-theme-stackoverflow'".

**Why it happens:** Theme flag allows overriding config, but override theme may not be in node_modules.

**How to avoid:** Check theme installation before calling resumed:
```go
themePath := filepath.Join(projectDir, "node_modules", "jsonresume-theme-" + theme)
if _, err := os.Stat(themePath); os.IsNotExist(err) {
    return fmt.Errorf("theme %q not installed. Run: npm install jsonresume-theme-%s", theme, theme)
}
```

**Warning signs:**
- Runtime errors from resumed about missing modules
- Works with default theme, fails with --theme flag

**Confidence:** HIGH - User input validation principle.

---

### Pitfall 5: Not Preserving resume.json for Debugging

**What goes wrong:** PDF generation fails but intermediate JSON is lost, making debugging impossible.

**Why it happens:** Using temp files with automatic cleanup via `defer os.Remove()`.

**How to avoid:** Write resume.json to application directory (not temp) so users can inspect it:
```go
jsonPath := filepath.Join(appDir, "resume.json")
os.WriteFile(jsonPath, resumeJSON, 0644)
// Keep file for debugging; don't remove
```

**Warning signs:**
- Users report validation failures but can't provide JSON
- Can't reproduce issues without CV content

**Confidence:** HIGH - Debuggability principle.

---

### Pitfall 6: Assuming schema.json is Valid JSON Schema

**What goes wrong:** Embedded schema fails to compile with "unknown keyword" or "invalid $ref".

**Why it happens:** JSON Resume schema may have typos, use newer draft features, or have incorrect references.

**How to avoid:** Validate the schema itself during initialization:
```go
func NewValidator() (*Validator, error) {
    // ... load schema ...

    compiler := jsonschema.NewCompiler()
    if err := compiler.AddResource("resume.schema.json", schemaObj); err != nil {
        // This catches invalid schema
        return nil, fmt.Errorf("invalid JSON Resume schema: %w", err)
    }

    schema, err := compiler.Compile("resume.schema.json")
    // ...
}
```

**Warning signs:**
- Validation always fails with schema-related errors
- Different results with same JSON on different machines

**Confidence:** MEDIUM - Good defensive programming practice.

## Code Examples

Verified patterns from official sources:

### Complete Converter Implementation

```go
// internal/generator/converter.go
package generator

import (
    "context"
    "fmt"
    "strings"

    "github.com/richq/m2cv/internal/assets"
    "github.com/richq/m2cv/internal/executor"
)

// Converter converts markdown CVs to JSON Resume format using Claude.
type Converter struct {
    claudeExec executor.ClaudeExecutor
}

// NewConverter creates a new Converter.
func NewConverter(claudeExec executor.ClaudeExecutor) *Converter {
    return &Converter{claudeExec: claudeExec}
}

// Convert converts a markdown CV to JSON Resume format.
// Returns raw JSON (as bytes) and any error.
func (c *Converter) Convert(ctx context.Context, cvContent string, model string) ([]byte, error) {
    // Load prompt template
    promptTemplate, err := assets.GetPrompt("md-to-json-resume")
    if err != nil {
        return nil, fmt.Errorf("failed to load prompt: %w", err)
    }

    // Substitute CV content
    prompt := strings.ReplaceAll(promptTemplate, "{{.CV}}", cvContent)

    // Build executor options
    var opts []executor.ExecuteOption
    if model != "" {
        opts = append(opts, executor.WithModel(model))
    }

    // Execute Claude
    output, err := c.claudeExec.Execute(ctx, prompt, opts...)
    if err != nil {
        return nil, fmt.Errorf("Claude execution failed: %w", err)
    }

    // Extract JSON from output
    jsonData, err := ExtractJSON([]byte(output))
    if err != nil {
        return nil, fmt.Errorf("failed to extract JSON: %w", err)
    }

    return jsonData, nil
}
```

**Source:** Synthesized from existing patterns (extractor/folder_name.go, executor/claude.go).

---

### JSON Schema Validation Error Formatting

```go
// Example of handling validation errors with rich output
func validateAndReport(validator *generator.Validator, resumeJSON []byte) error {
    result, err := validator.ValidateWithDetails(resumeJSON)
    if err != nil {
        return fmt.Errorf("validation check failed: %w", err)
    }

    if !result.Valid {
        fmt.Println("JSON Resume validation failed:")
        for _, errMsg := range result.Errors {
            fmt.Printf("  - %s\n", errMsg)
        }
        return fmt.Errorf("schema validation failed with %d errors", len(result.Errors))
    }

    fmt.Println("✓ JSON Resume schema validation passed")
    return nil
}
```

**Source:** jsonschema v6 package patterns from pkg.go.dev.

---

### Complete Test with Mock Executor

```go
// cmd/generate_test.go
package cmd

import (
    "context"
    "os"
    "path/filepath"
    "testing"

    "github.com/richq/m2cv/internal/executor"
)

// mockClaudeExecutor for testing
type mockClaudeExecutor struct {
    response string
    err      error
}

func (m *mockClaudeExecutor) Execute(ctx context.Context, prompt string, opts ...executor.ExecuteOption) (string, error) {
    if m.err != nil {
        return "", m.err
    }
    return m.response, nil
}

func TestGenerateCommand(t *testing.T) {
    t.Parallel()

    // Setup test directory
    tmpDir := t.TempDir()
    appDir := filepath.Join(tmpDir, "applications", "test-app")
    os.MkdirAll(appDir, 0755)

    // Create test CV
    cvContent := "# John Doe\n\nSoftware Engineer"
    os.WriteFile(filepath.Join(appDir, "optimized-cv-1.md"), []byte(cvContent), 0644)

    // Create test config
    configContent := `
base_cv_path: base-cv.md
default_theme: even
themes:
  - even
default_model: claude-sonnet-4-20250514
`
    os.WriteFile(filepath.Join(tmpDir, "m2cv.yml"), []byte(configContent), 0644)

    // Mock Claude response with valid JSON Resume
    mockResponse := `{
  "basics": {
    "name": "John Doe",
    "label": "Software Engineer",
    "email": "john@example.com"
  },
  "work": []
}`

    // Test would inject mock executor and verify PDF creation
    // (Full test requires npm/resumed setup, typically integration test)
}
```

**Source:** Existing test patterns from cmd/apply_test.go, cmd/optimize_test.go.

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| resume-cli | resumed | ~2020 | Active maintenance, faster, better API |
| Manual JSON validation | jsonschema v6 | Go 1.20+ | Supports draft-07, excellent errors |
| String contains for JSON | ExtractJSON with regex | LLM era (2023+) | Handles markdown fences, robust |
| Direct binary execution | npx for Node tools | npm 5.2+ (2017) | Better environment handling |
| Inline schema | embed.FS for schema | Go 1.16 (2021) | Single binary, versioned schema |

**Deprecated/outdated:**
- **resume-cli**: Unmaintained since 2017; use resumed
- **@jsonresume/resume-schema (npm)**: Use embedded schema (ensures version consistency)
- **Direct `resumed` execution without npx**: May fail with theme resolution

**Confidence:** HIGH

## Open Questions

Things that couldn't be fully resolved:

1. **Claude JSON Output Reliability**
   - What we know: Claude can return JSON with markdown fences or mixed with text
   - What's unclear: Success rate of ExtractJSON with various Claude models
   - Recommendation: Implement retry logic (1-2 retries) if JSON extraction fails; log failures for monitoring

2. **resumed Output Format Options**
   - What we know: resumed supports PDF export
   - What's unclear: HTML export option, viewport/page size configuration flags
   - Recommendation: Start with PDF only; investigate HTML export for future iteration (web preview feature)

3. **Theme Compatibility with Schema Versions**
   - What we know: 400+ themes available, most support JSON Resume 1.0.0
   - What's unclear: Which themes handle optional fields gracefully
   - Recommendation: Test with 8 curated themes (already selected in init/theme.go); document known issues

4. **JSON Resume Schema Updates**
   - What we know: Current schema is draft-07, embedded in binary
   - What's unclear: Update frequency, breaking changes
   - Recommendation: Check schema updates quarterly; re-embed if format changes; version check in validation

5. **Model Selection for Conversion**
   - What we know: User can override model with -m flag
   - What's unclear: Which models perform best for MD -> JSON conversion (haiku for speed vs sonnet for accuracy)
   - Recommendation: Default to config model; document model tradeoffs in README

## Sources

### Primary (HIGH confidence)

**Libraries:**
- [jsonschema v6 - pkg.go.dev](https://pkg.go.dev/github.com/santhosh-tekuri/jsonschema/v6) - Official Go package docs
- [jsonschema GitHub](https://github.com/santhosh-tekuri/jsonschema) - Source repository, v6.0.2 (May 2025)
- [A JSON schema package for Go - Google Open Source Blog](https://opensource.googleblog.com/2026/01/a-json-schema-package-for-go.html) - Jan 2026 overview

**JSON Resume:**
- [JSON Resume Schema](https://jsonresume.org/schema) - Official schema documentation
- [resume-schema GitHub](https://github.com/jsonresume/resume-schema) - Schema source, version 1.0.0
- [JSON Resume Documentation](https://docs.jsonresume.org/schema) - Usage guide

**resumed:**
- [resumed - npm](https://www.npmjs.com/package/resumed) - Package registry, version 6.1.0 (Oct 2025)
- [resumed GitHub](https://github.com/rbardini/resumed) - Source repository

**Existing Codebase:**
- internal/executor/claude.go - ClaudeExecutor pattern
- internal/executor/npm.go - NPMExecutor pattern, FindNodeExecutable
- internal/extractor/folder_name.go - Prompt substitution pattern
- internal/assets/assets.go - Embedded assets pattern

### Secondary (MEDIUM confidence)

- [How to Consistently Retrieve Valid JSON from Claude 3.5 in Go](https://dev.to/embiem/how-to-consistently-retrieve-valid-json-from-claude-35-in-go-1g5b) - JSON extraction patterns
- [Claude API Structured Outputs](https://platform.claude.com/docs/en/build-with-claude/structured-outputs) - Claude output format docs
- [JSON Resume Themes](https://jsonresume.org/themes) - Theme catalog (400+ themes)

### Tertiary (LOW confidence - marked for validation)

- [Jsonformer Claude](https://github.com/1rgs/jsonformer-claude) - Alternative JSON extraction approach (not verified)
- Theme-specific documentation - Individual theme repos vary in quality

## Metadata

**Confidence breakdown:**
- Standard stack: **HIGH** - jsonschema v6 verified via pkg.go.dev/GitHub, resumed verified via npm, existing patterns proven in codebase
- Architecture patterns: **HIGH** - ExtractJSON pattern documented, jsonschema API from official docs, resumed usage from npm docs
- Pitfalls: **HIGH** - JSON extraction issues well-documented, working directory/validation pitfalls proven in existing code
- resumed command line: **MEDIUM** - Basic usage verified but specific flags not exhaustively tested
- Theme compatibility: **MEDIUM** - 8 curated themes tested during init implementation, broader ecosystem not verified

**Research date:** 2026-02-03
**Valid until:** ~2026-04-03 (60 days - jsonschema stable, resumed active, Claude CLI may evolve)

**Overall confidence:** HIGH

Phase 5 implementation can proceed with high confidence. All critical patterns (JSON extraction from LLM output, schema validation, PDF export via resumed) have authoritative sources or proven existing patterns. The main risk areas (Claude output variability, theme compatibility) have recommended mitigation strategies (robust extraction, curated theme list).
