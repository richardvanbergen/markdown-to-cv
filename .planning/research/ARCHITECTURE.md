# Architecture Patterns for m2cv Go CLI

**Domain:** Go CLI tool orchestrating external processes (claude, npm/npx)
**Researched:** 2026-02-03
**Confidence:** HIGH (based on established Go CLI patterns and cobra conventions)

## Executive Summary

m2cv is a Go CLI that follows the **Command-Executor-Service** pattern common in Go tools that orchestrate external processes. The architecture separates concerns into: CLI layer (cobra commands), service layer (business logic), executor layer (process orchestration), and persistence layer (filesystem + config).

Key architectural decision: **Dependency injection through command constructors** enables testability while keeping cobra's flag binding ergonomics.

## Recommended Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         CLI Layer                            │
│  (cobra commands: init, apply, optimize, generate)          │
│                   - Flag parsing                             │
│                   - User interaction                         │
│                   - Output formatting                        │
└──────────────────┬──────────────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────────────┐
│                      Service Layer                           │
│  (business logic for each command)                           │
│    - InitService: project initialization                     │
│    - ApplyService: application folder creation               │
│    - OptimizeService: CV tailoring orchestration             │
│    - GenerateService: JSON conversion + PDF export           │
└──────┬────────────────┬────────────────┬────────────────────┘
       │                │                │
       ▼                ▼                ▼
┌──────────────┐ ┌──────────────┐ ┌─────────────────────────┐
│   Executor   │ │  Persistence │ │   Template Engine       │
│    Layer     │ │    Layer     │ │   (prompts + schema)    │
│              │ │              │ │                         │
│ - ClaudeExec │ │ - ConfigRepo │ │ - embed.FS prompts/     │
│ - NPMExec    │ │ - FileRepo   │ │ - embed.FS schema/      │
│              │ │ - Validator  │ │ - TemplateRenderer      │
└──────┬───────┘ └──────┬───────┘ └────────────┬────────────┘
       │                │                       │
       ▼                ▼                       ▼
┌─────────────────────────────────────────────────────────────┐
│                   External Dependencies                      │
│                                                              │
│  OS Processes:        Filesystem:        Embedded Assets:   │
│  - claude -p          - m2cv.yml         - prompts/*.txt    │
│  - npm install        - applications/    - schema/*.json    │
│  - npx resumed        - package.json     │                  │
└─────────────────────────────────────────────────────────────┘
```

### Component Boundaries

| Component | Responsibility | Dependencies | Interface |
|-----------|---------------|--------------|-----------|
| **CLI Layer** | User interaction, flag parsing, output | Service layer | cobra.Command structs |
| **Service Layer** | Business logic orchestration | Executors, persistence, templates | Exported service structs with methods |
| **Executor Layer** | Process invocation, stdout/stderr handling | os/exec | Interface-based (mockable) |
| **Persistence Layer** | Config/file I/O, validation | Filesystem | Interface-based (mockable) |
| **Template Engine** | Prompt rendering, schema loading | embed.FS | Exported functions |

## Data Flow

### 1. Init Command Flow

```
User runs: m2cv init --base-cv ~/cv.md --theme flat

CLI Layer (cmd/init.go)
  ├─> Parse flags
  └─> InitService.Initialize(config)
        ├─> ConfigRepo.CreateProject()
        │     └─> Write m2cv.yml
        ├─> NPMExec.Init()
        │     └─> Run: npm init -y
        ├─> NPMExec.InstallResumed()
        │     └─> Run: npm install resumed
        └─> NPMExec.InstallTheme(themeName)
              └─> Run: npm install jsonresume-theme-{name}

Output: Project initialized with m2cv.yml, package.json, node_modules/
```

### 2. Apply Command Flow

```
User runs: m2cv apply job-description.md

CLI Layer (cmd/apply.go)
  ├─> Read job-description.md (FileRepo)
  └─> ApplyService.CreateApplication(jobDescContent)
        ├─> TemplateRenderer.RenderPrompt("extract-name", jobDescContent)
        ├─> ClaudeExec.Execute(prompt)
        │     └─> Run: claude -p --output-format text
        ├─> Parse company + role from output
        ├─> FileRepo.CreateAppFolder(sanitizedName)
        └─> FileRepo.WriteFile("applications/{name}/job-description.md", content)

Output: applications/acme-senior-dev/job-description.md created
```

### 3. Optimize Command Flow

```
User runs: m2cv optimize --ats acme-senior-dev

CLI Layer (cmd/optimize.go)
  ├─> Load m2cv.yml (ConfigRepo)
  ├─> Read base CV path from config (FileRepo)
  └─> OptimizeService.OptimizeCV(appName, atsFlag)
        ├─> FileRepo.ReadJobDescription(appName)
        ├─> FileRepo.ReadBaseCV(baseCVPath)
        ├─> FileRepo.GetNextVersion(appName, "optimized-cv")
        ├─> TemplateRenderer.RenderPrompt("optimize", {
        │     baseCv: content,
        │     jobDesc: content,
        │     atsMode: true
        │   })
        ├─> ClaudeExec.Execute(prompt, model)
        │     └─> Run: claude -p -m {model} --output-format text
        └─> FileRepo.WriteFile("applications/{name}/optimized-cv-{version}.md", output)

Output: applications/acme-senior-dev/optimized-cv-1.md created
```

### 4. Generate Command Flow

```
User runs: m2cv generate --theme flat acme-senior-dev

CLI Layer (cmd/generate.go)
  ├─> Load m2cv.yml (ConfigRepo)
  ├─> Preflight: check resumed installed (NPMExec.CheckResumed)
  └─> GenerateService.GeneratePDF(appName, theme)
        ├─> FileRepo.GetLatestOptimizedCV(appName)
        ├─> TemplateRenderer.RenderPrompt("md-to-json-resume", {
        │     markdown: content,
        │     schema: schemaContent
        │   })
        ├─> ClaudeExec.Execute(prompt)
        │     └─> Run: claude -p --output-format text
        ├─> JSONValidator.Validate(output, schema)
        │     └─> Use: santhosh-tekuri/jsonschema/v6
        ├─> FileRepo.WriteFile("applications/{name}/resume.json", validated)
        └─> NPMExec.ExportPDF(theme, jsonPath, pdfPath)
              └─> Run: npx resumed export resume.pdf --resume resume.json --theme {theme}

Output: applications/acme-senior-dev/resume.pdf created
```

## Package Structure

Recommended directory layout for m2cv:

```
markdown-to-cv/
├── go.mod
├── go.sum
├── main.go                    # Entry point: cobra.Execute()
│
├── cmd/                       # CLI Layer
│   ├── root.go               # Root command + persistent flags
│   ├── init.go               # init command
│   ├── apply.go              # apply command
│   ├── optimize.go           # optimize command
│   └── generate.go           # generate command
│
├── internal/
│   ├── service/              # Service Layer
│   │   ├── init.go           # InitService
│   │   ├── apply.go          # ApplyService
│   │   ├── optimize.go       # OptimizeService
│   │   └── generate.go       # GenerateService
│   │
│   ├── executor/             # Executor Layer
│   │   ├── executor.go       # Interface definitions
│   │   ├── claude.go         # ClaudeExecutor implementation
│   │   └── npm.go            # NPMExecutor implementation
│   │
│   ├── persistence/          # Persistence Layer
│   │   ├── config.go         # ConfigRepository (m2cv.yml)
│   │   ├── file.go           # FileRepository (app folders, CV files)
│   │   └── validator.go      # JSONValidator (schema validation)
│   │
│   ├── template/             # Template Engine
│   │   ├── template.go       # TemplateRenderer
│   │   ├── prompts.go        # embed.FS access
│   │   └── schema.go         # Schema loading
│   │
│   └── assets/               # Embedded Assets
│       ├── prompts/
│       │   ├── extract-name.txt
│       │   ├── optimize.txt
│       │   ├── optimize-ats.txt
│       │   └── md-to-json-resume.txt
│       └── schema/
│           └── resume.schema.json
│
└── testdata/                 # Test fixtures
    ├── sample-cv.md
    └── sample-job-desc.md
```

## Patterns to Follow

### Pattern 1: Interface-Based Executors

**What:** Define interfaces for external process execution, inject implementations.

**When:** Any time you shell out to external commands.

**Why:** Enables unit testing without running actual processes, supports mocking.

**Example:**

```go
// internal/executor/executor.go
package executor

import "context"

// ClaudeExecutor abstracts claude CLI invocation
type ClaudeExecutor interface {
    Execute(ctx context.Context, prompt string, opts ...Option) (string, error)
}

// NPMExecutor abstracts npm/npx invocation
type NPMExecutor interface {
    Init(ctx context.Context, dir string) error
    Install(ctx context.Context, dir string, packages ...string) error
    Run(ctx context.Context, dir string, script string, args ...string) error
}

// Option pattern for flexibility
type Option func(*ExecuteOptions)

type ExecuteOptions struct {
    Model         string
    OutputFormat  string
}

func WithModel(model string) Option {
    return func(o *ExecuteOptions) {
        o.Model = model
    }
}
```

```go
// internal/executor/claude.go
package executor

import (
    "context"
    "os/exec"
    "strings"
)

type claudeExecutor struct {
    binaryPath string
}

func NewClaudeExecutor(binaryPath string) ClaudeExecutor {
    return &claudeExecutor{binaryPath: binaryPath}
}

func (c *claudeExecutor) Execute(ctx context.Context, prompt string, opts ...Option) (string, error) {
    options := &ExecuteOptions{
        OutputFormat: "text",
    }
    for _, opt := range opts {
        opt(options)
    }

    args := []string{"-p", "--output-format", options.OutputFormat}
    if options.Model != "" {
        args = append(args, "-m", options.Model)
    }

    cmd := exec.CommandContext(ctx, c.binaryPath, args...)
    cmd.Stdin = strings.NewReader(prompt)

    output, err := cmd.Output()
    if err != nil {
        return "", fmt.Errorf("claude execution failed: %w", err)
    }

    return string(output), nil
}
```

### Pattern 2: Repository Pattern for Filesystem

**What:** Abstract filesystem operations behind interfaces.

**When:** Reading/writing config, application folders, versioned files.

**Why:** Enables testing with in-memory implementations, centralizes path logic.

**Example:**

```go
// internal/persistence/config.go
package persistence

import "context"

type Config struct {
    BaseCVPath   string   `yaml:"base_cv_path"`
    DefaultTheme string   `yaml:"default_theme"`
    Themes       []string `yaml:"themes"`
    DefaultModel string   `yaml:"default_model"`
}

type ConfigRepository interface {
    Load(ctx context.Context, projectDir string) (*Config, error)
    Save(ctx context.Context, projectDir string, cfg *Config) error
    Exists(ctx context.Context, projectDir string) (bool, error)
}
```

```go
// internal/persistence/file.go
package persistence

type FileRepository interface {
    CreateAppFolder(ctx context.Context, projectDir, appName string) error
    WriteJobDescription(ctx context.Context, projectDir, appName, content string) error
    ReadJobDescription(ctx context.Context, projectDir, appName string) (string, error)
    GetNextVersion(ctx context.Context, projectDir, appName, prefix string) (int, error)
    WriteVersionedFile(ctx context.Context, projectDir, appName, prefix string, version int, content string) error
    GetLatestVersionedFile(ctx context.Context, projectDir, appName, prefix string) (string, error)
}
```

### Pattern 3: Service Layer Orchestration

**What:** Services coordinate executors, repositories, and templates to implement business logic.

**When:** Each command's core logic.

**Why:** Keeps commands thin, enables reuse, centralizes transaction-like operations.

**Example:**

```go
// internal/service/optimize.go
package service

import (
    "context"
    "fmt"

    "markdown-to-cv/internal/executor"
    "markdown-to-cv/internal/persistence"
    "markdown-to-cv/internal/template"
)

type OptimizeService struct {
    claudeExec executor.ClaudeExecutor
    fileRepo   persistence.FileRepository
    configRepo persistence.ConfigRepository
    templates  *template.Renderer
}

func NewOptimizeService(
    claudeExec executor.ClaudeExecutor,
    fileRepo persistence.FileRepository,
    configRepo persistence.ConfigRepository,
    templates *template.Renderer,
) *OptimizeService {
    return &OptimizeService{
        claudeExec: claudeExec,
        fileRepo:   fileRepo,
        configRepo: configRepo,
        templates:  templates,
    }
}

func (s *OptimizeService) OptimizeCV(ctx context.Context, projectDir, appName string, atsMode bool) (string, error) {
    // Load config
    cfg, err := s.configRepo.Load(ctx, projectDir)
    if err != nil {
        return "", fmt.Errorf("load config: %w", err)
    }

    // Read base CV
    baseCVContent, err := os.ReadFile(cfg.BaseCVPath)
    if err != nil {
        return "", fmt.Errorf("read base CV: %w", err)
    }

    // Read job description
    jobDesc, err := s.fileRepo.ReadJobDescription(ctx, projectDir, appName)
    if err != nil {
        return "", fmt.Errorf("read job description: %w", err)
    }

    // Select template
    templateName := "optimize"
    if atsMode {
        templateName = "optimize-ats"
    }

    // Render prompt
    prompt, err := s.templates.Render(templateName, map[string]string{
        "base_cv": string(baseCVContent),
        "job_description": jobDesc,
    })
    if err != nil {
        return "", fmt.Errorf("render prompt: %w", err)
    }

    // Execute Claude
    optimizedCV, err := s.claudeExec.Execute(ctx, prompt, executor.WithModel(cfg.DefaultModel))
    if err != nil {
        return "", fmt.Errorf("claude execution: %w", err)
    }

    // Get next version
    version, err := s.fileRepo.GetNextVersion(ctx, projectDir, appName, "optimized-cv")
    if err != nil {
        return "", fmt.Errorf("get next version: %w", err)
    }

    // Write versioned file
    err = s.fileRepo.WriteVersionedFile(ctx, projectDir, appName, "optimized-cv", version, optimizedCV)
    if err != nil {
        return "", fmt.Errorf("write optimized CV: %w", err)
    }

    return fmt.Sprintf("optimized-cv-%d.md", version), nil
}
```

### Pattern 4: Embedded Assets with embed.FS

**What:** Use Go 1.16+ `embed` directive to bundle prompts and schemas into binary.

**When:** Static assets needed at runtime.

**Why:** Single binary distribution, no external file dependencies.

**Example:**

```go
// internal/template/prompts.go
package template

import (
    "embed"
    "fmt"
    "text/template"
)

//go:embed prompts/*.txt
var promptFS embed.FS

//go:embed schema/*.json
var schemaFS embed.FS

type Renderer struct {
    templates map[string]*template.Template
}

func NewRenderer() (*Renderer, error) {
    r := &Renderer{
        templates: make(map[string]*template.Template),
    }

    entries, err := promptFS.ReadDir("prompts")
    if err != nil {
        return nil, fmt.Errorf("read prompt dir: %w", err)
    }

    for _, entry := range entries {
        if entry.IsDir() {
            continue
        }

        name := strings.TrimSuffix(entry.Name(), ".txt")
        content, err := promptFS.ReadFile("prompts/" + entry.Name())
        if err != nil {
            return nil, fmt.Errorf("read prompt %s: %w", name, err)
        }

        tmpl, err := template.New(name).Parse(string(content))
        if err != nil {
            return nil, fmt.Errorf("parse template %s: %w", name, err)
        }

        r.templates[name] = tmpl
    }

    return r, nil
}

func (r *Renderer) Render(name string, data interface{}) (string, error) {
    tmpl, ok := r.templates[name]
    if !ok {
        return "", fmt.Errorf("template not found: %s", name)
    }

    var buf strings.Builder
    if err := tmpl.Execute(&buf, data); err != nil {
        return "", fmt.Errorf("execute template: %w", err)
    }

    return buf.String(), nil
}

func (r *Renderer) GetSchema(name string) ([]byte, error) {
    return schemaFS.ReadFile("schema/" + name)
}
```

### Pattern 5: Cobra Command Initialization with Dependency Injection

**What:** Construct services in `main.go`, inject into command constructors.

**When:** Setting up cobra commands.

**Why:** Testable commands, explicit dependencies, no global state.

**Example:**

```go
// main.go
package main

import (
    "markdown-to-cv/cmd"
    "markdown-to-cv/internal/executor"
    "markdown-to-cv/internal/persistence"
    "markdown-to-cv/internal/service"
    "markdown-to-cv/internal/template"
)

func main() {
    // Initialize dependencies
    claudeExec := executor.NewClaudeExecutor("claude")
    npmExec := executor.NewNPMExecutor("npm")

    configRepo := persistence.NewYAMLConfigRepository()
    fileRepo := persistence.NewOSFileRepository()
    validator := persistence.NewJSONSchemaValidator()

    templates, err := template.NewRenderer()
    if err != nil {
        log.Fatal(err)
    }

    // Initialize services
    initSvc := service.NewInitService(npmExec, configRepo, fileRepo)
    applySvc := service.NewApplyService(claudeExec, fileRepo, configRepo, templates)
    optimizeSvc := service.NewOptimizeService(claudeExec, fileRepo, configRepo, templates)
    generateSvc := service.NewGenerateService(claudeExec, npmExec, fileRepo, configRepo, validator, templates)

    // Build command tree
    rootCmd := cmd.NewRootCommand()
    rootCmd.AddCommand(cmd.NewInitCommand(initSvc))
    rootCmd.AddCommand(cmd.NewApplyCommand(applySvc))
    rootCmd.AddCommand(cmd.NewOptimizeCommand(optimizeSvc))
    rootCmd.AddCommand(cmd.NewGenerateCommand(generateSvc))

    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

```go
// cmd/optimize.go
package cmd

import (
    "context"
    "fmt"

    "github.com/spf13/cobra"
    "markdown-to-cv/internal/service"
)

func NewOptimizeCommand(svc *service.OptimizeService) *cobra.Command {
    var atsMode bool

    cmd := &cobra.Command{
        Use:   "optimize <application-name>",
        Short: "Optimize CV for a job application",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            ctx := context.Background()
            projectDir := "." // or from global flag
            appName := args[0]

            outputFile, err := svc.OptimizeCV(ctx, projectDir, appName, atsMode)
            if err != nil {
                return fmt.Errorf("optimize failed: %w", err)
            }

            fmt.Printf("Created: applications/%s/%s\n", appName, outputFile)
            return nil
        },
    }

    cmd.Flags().BoolVar(&atsMode, "ats", false, "Optimize for ATS (Applicant Tracking Systems)")

    return cmd
}
```

### Pattern 6: Preflight Checks with PersistentPreRunE

**What:** Validate external dependencies before command execution.

**When:** Commands depend on external binaries (claude, npm, resumed).

**Why:** Fail fast with clear error messages instead of cryptic exec errors.

**Example:**

```go
// cmd/root.go
package cmd

import (
    "context"
    "fmt"
    "os/exec"

    "github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "m2cv",
        Short: "Markdown to CV - AI-powered resume tailoring CLI",
        PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
            // Check claude is in PATH
            if _, err := exec.LookPath("claude"); err != nil {
                return fmt.Errorf("claude CLI not found in PATH - install from https://claude.ai/download")
            }

            return nil
        },
    }

    return cmd
}
```

```go
// cmd/generate.go - additional check
func NewGenerateCommand(svc *service.GenerateService) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "generate <application-name>",
        Short: "Generate PDF resume from optimized CV",
        PreRunE: func(cmd *cobra.Command, args []string) error {
            // Check resumed is installed (via npm)
            projectDir := "."
            if err := checkResumedInstalled(projectDir); err != nil {
                return fmt.Errorf("resumed not installed - run 'm2cv init' first")
            }
            return nil
        },
        RunE: func(cmd *cobra.Command, args []string) error {
            // ... implementation
        },
    }

    return cmd
}
```

## Anti-Patterns to Avoid

### Anti-Pattern 1: Global State in Services

**What goes wrong:** Using package-level variables for configuration, executors, or repositories.

**Why bad:**
- Makes testing difficult (need to reset state between tests)
- Hides dependencies (unclear what a function needs)
- Causes race conditions in concurrent scenarios

**Instead:** Use dependency injection through constructors.

**Example:**

```go
// BAD - global state
var (
    claudeExec executor.ClaudeExecutor
    config     *persistence.Config
)

func OptimizeCV(appName string) error {
    // Uses global claudeExec and config
}

// GOOD - explicit dependencies
type OptimizeService struct {
    claudeExec executor.ClaudeExecutor
    config     *persistence.Config
}

func (s *OptimizeService) OptimizeCV(appName string) error {
    // Uses s.claudeExec and s.config
}
```

### Anti-Pattern 2: Command Structs with Business Logic

**What goes wrong:** Implementing business logic directly in cobra command RunE functions.

**Why bad:**
- Can't test logic without invoking cobra framework
- Tight coupling between CLI parsing and business logic
- Difficult to reuse logic in different contexts (e.g., API server)

**Instead:** Keep commands thin, delegate to service layer.

**Example:**

```go
// BAD - logic in command
RunE: func(cmd *cobra.Command, args []string) error {
    cfg, _ := loadConfig("m2cv.yml")
    baseCVContent, _ := os.ReadFile(cfg.BaseCVPath)
    jobDesc, _ := os.ReadFile("applications/" + args[0] + "/job-description.md")
    prompt := fmt.Sprintf("Optimize this CV:\n%s\nFor this job:\n%s", baseCVContent, jobDesc)

    claudeCmd := exec.Command("claude", "-p")
    claudeCmd.Stdin = strings.NewReader(prompt)
    output, _ := claudeCmd.Output()

    os.WriteFile("applications/"+args[0]+"/optimized-cv-1.md", output, 0644)
    return nil
}

// GOOD - delegate to service
RunE: func(cmd *cobra.Command, args []string) error {
    outputFile, err := svc.OptimizeCV(ctx, ".", args[0], atsMode)
    if err != nil {
        return fmt.Errorf("optimize failed: %w", err)
    }
    fmt.Printf("Created: applications/%s/%s\n", args[0], outputFile)
    return nil
}
```

### Anti-Pattern 3: Tight Coupling to exec.Command

**What goes wrong:** Calling `exec.Command` directly in service layer.

**Why bad:**
- Services can't be tested without running actual processes
- No way to mock or stub external commands
- Difficult to add cross-cutting concerns (logging, retries, timeouts)

**Instead:** Use executor interfaces.

**Example:**

```go
// BAD - direct exec.Command
func (s *OptimizeService) OptimizeCV(appName string) error {
    cmd := exec.Command("claude", "-p", "--output-format", "text")
    cmd.Stdin = strings.NewReader(prompt)
    output, err := cmd.Output()
    // ...
}

// GOOD - executor interface
func (s *OptimizeService) OptimizeCV(appName string) error {
    output, err := s.claudeExec.Execute(ctx, prompt, executor.WithModel(cfg.DefaultModel))
    // ...
}

// Test with mock
type mockClaudeExecutor struct {
    output string
    err    error
}

func (m *mockClaudeExecutor) Execute(ctx context.Context, prompt string, opts ...Option) (string, error) {
    return m.output, m.err
}
```

### Anti-Pattern 4: Embedded Assets as String Literals

**What goes wrong:** Storing prompts or schemas as Go string literals in code.

**Why bad:**
- Harder to edit (need to escape quotes, newlines)
- IDE doesn't provide syntax highlighting for prompt content
- Can't validate schema files separately
- Mixing prompt content with Go code

**Instead:** Use embed.FS with separate files.

**Example:**

```go
// BAD - string literal
const optimizePrompt = `You are a CV optimization assistant.

Base CV:
{{.BaseCv}}

Job Description:
{{.JobDesc}}

Optimize the CV for this role.`

// GOOD - embedded file
// internal/assets/prompts/optimize.txt
//
// You are a CV optimization assistant.
//
// Base CV:
// {{.BaseCv}}
//
// Job Description:
// {{.JobDesc}}
//
// Optimize the CV for this role.

//go:embed prompts/*.txt
var promptFS embed.FS
```

### Anti-Pattern 5: Silent Failures in Process Execution

**What goes wrong:** Ignoring stderr or exit codes from external processes.

**Why bad:**
- Claude might fail but you use empty output
- npm install might fail but you assume success
- Users get confusing downstream errors instead of root cause

**Instead:** Capture and surface stderr, check exit codes.

**Example:**

```go
// BAD - ignores stderr
cmd := exec.Command("claude", "-p")
output, err := cmd.Output()
if err != nil {
    return err // Generic error, no context
}

// GOOD - captures stderr
cmd := exec.Command("claude", "-p")
var stderr bytes.Buffer
cmd.Stderr = &stderr

output, err := cmd.Output()
if err != nil {
    return fmt.Errorf("claude failed: %w\nstderr: %s", err, stderr.String())
}
```

### Anti-Pattern 6: Hardcoded Paths

**What goes wrong:** Hardcoding paths like `"applications/"` or `"m2cv.yml"` throughout codebase.

**Why bad:**
- Can't run tests in isolation (all write to same directory)
- Can't support custom project locations
- Brittle when project structure changes

**Instead:** Accept projectDir as parameter, use path construction functions.

**Example:**

```go
// BAD - hardcoded
func ReadJobDescription(appName string) (string, error) {
    return os.ReadFile("applications/" + appName + "/job-description.md")
}

// GOOD - parameterized
func (r *FileRepository) ReadJobDescription(ctx context.Context, projectDir, appName string) (string, error) {
    path := filepath.Join(projectDir, "applications", appName, "job-description.md")
    return os.ReadFile(path)
}
```

## Build Order and Dependencies

### Phase 1: Foundation (No Dependencies)

**Components:**
1. Package structure (`internal/executor`, `internal/persistence`, etc.)
2. Interface definitions (Executor, Repository interfaces)
3. Embedded assets structure (`internal/assets/prompts/`, `internal/assets/schema/`)

**Rationale:** Establishes contracts before implementations.

### Phase 2: Core Utilities (Depends on Phase 1)

**Components:**
1. Template renderer (`internal/template/`)
2. Config repository implementation (`internal/persistence/config.go`)
3. File repository implementation (`internal/persistence/file.go`)

**Rationale:** Persistence layer needed before executors and services can store results.

### Phase 3: Executors (Depends on Phase 1)

**Components:**
1. Claude executor (`internal/executor/claude.go`)
2. NPM executor (`internal/executor/npm.go`)

**Rationale:** Can be developed independently once interfaces are defined. Needed by service layer.

### Phase 4: Services (Depends on Phase 2 & 3)

**Components:**
1. Init service (`internal/service/init.go`)
2. Apply service (`internal/service/apply.go`)
3. Optimize service (`internal/service/optimize.go`)
4. Generate service (`internal/service/generate.go`)

**Rationale:** Orchestrates executors and repositories. Must come after both are implemented.

**Suggested order:**
- Init service first (simplest, sets up project)
- Apply service second (depends on Claude executor for name extraction)
- Optimize service third (builds on Apply pattern)
- Generate service last (most complex, uses JSON validation)

### Phase 5: CLI Commands (Depends on Phase 4)

**Components:**
1. Root command with preflight checks (`cmd/root.go`)
2. Init command (`cmd/init.go`)
3. Apply command (`cmd/apply.go`)
4. Optimize command (`cmd/optimize.go`)
5. Generate command (`cmd/generate.go`)

**Rationale:** Thin layer over services. Should be implemented after services are testable.

### Phase 6: Integration (Depends on Phase 5)

**Components:**
1. Main entry point (`main.go`)
2. Dependency wiring
3. End-to-end testing

**Rationale:** Wires everything together. Only possible when all components exist.

### Dependency Graph

```
Phase 1: Interfaces & Structure
    ↓
Phase 2: Persistence ← → Phase 3: Executors
    ↓                       ↓
    └────→ Phase 4: Services ←────┘
              ↓
         Phase 5: CLI Commands
              ↓
         Phase 6: Integration
```

**Critical path:** Interfaces → Persistence → Services → Commands → Integration

**Parallelizable:** Executors can be built alongside Persistence (both depend only on Phase 1).

## Testing Strategy

### Unit Testing Layers

| Layer | Test Strategy | Mocking |
|-------|--------------|---------|
| Executors | Mock os/exec with test helpers | Use `exec.CommandContext` with custom path to test binary |
| Persistence | In-memory implementations | Use afero filesystem abstraction |
| Services | Mock executors + repositories | Inject mock implementations |
| Commands | Mock services | Inject mock service layer |

### Integration Testing

**Approach:** Use real executors with test fixtures.

**Setup:**
```go
// testdata/fixtures/sample-project/
//   ├── m2cv.yml
//   ├── base-cv.md
//   └── applications/
//       └── test-app/
//           └── job-description.md

func TestIntegration_OptimizeCommand(t *testing.T) {
    // Use real executors but with test project dir
    claudeExec := executor.NewClaudeExecutor("claude")
    // ... initialize other dependencies

    svc := service.NewOptimizeService(claudeExec, fileRepo, configRepo, templates)

    // Run service
    outputFile, err := svc.OptimizeCV(ctx, "testdata/fixtures/sample-project", "test-app", false)

    // Verify output exists
    assert.NoError(t, err)
    assert.FileExists(t, filepath.Join("testdata/fixtures/sample-project", "applications", "test-app", outputFile))
}
```

## Scalability Considerations

| Concern | At 1 User | At 10 Users | At 1000 Users |
|---------|-----------|-------------|---------------|
| **Process Execution** | Sequential claude calls | Add `--concurrency` flag for parallel optimizations | Consider rate limiting, queue system |
| **File I/O** | Direct filesystem access | Same (local CLI tool) | Same (local CLI tool) |
| **Config Management** | Single m2cv.yml | Support multiple projects via `--project-dir` flag | Same |
| **NPM Dependencies** | Local node_modules | Same | Consider global cache for themes |

**Note:** m2cv is a local CLI tool. Scalability is primarily about supporting multiple projects and parallelizing independent operations (e.g., optimizing multiple applications concurrently).

## Future Architecture Considerations

### Potential Extensions (Not in MVP)

1. **Plugin System:** Allow custom executors (e.g., OpenAI instead of Claude)
   - Interface: Already designed with executor abstraction
   - Implementation: Load executors from `.m2cv/plugins/`

2. **Watch Mode:** Auto-regenerate PDF when markdown changes
   - Component: New `WatchService` using `fsnotify`
   - Integration: Add `m2cv watch <application-name>` command

3. **Multi-CV Support:** Support multiple base CVs (frontend vs backend)
   - Config: Change `base_cv_path` to `base_cvs: map[string]string`
   - Command: Add `--base-cv` flag to optimize command

4. **API Server Mode:** Expose functionality via HTTP API
   - Architecture: Same service layer, new HTTP handler layer
   - Use: `m2cv serve --port 8080`

5. **Batch Operations:** Optimize all applications at once
   - Service: Add `OptimizeAll()` method
   - Executor: Add concurrency control with worker pool

## Sources

**HIGH Confidence** (based on established Go patterns):
- Standard Go project layout: https://github.com/golang-standards/project-layout
- Cobra CLI patterns: https://github.com/spf13/cobra
- Dependency injection in Go: https://github.com/google/wire (pattern, not specific tool)
- Go embed directive: Go 1.16+ standard library
- Interface-based testing: Standard Go testing practices

**MEDIUM Confidence** (common patterns, not formally documented):
- Repository pattern in Go CLIs
- Service layer orchestration patterns
- Executor abstraction for external processes

**Notes:**
- Architecture is derived from successful Go CLI tools: kubectl, gh, docker CLI
- Patterns are widely used but not always formally documented
- Specific to m2cv's requirements (process orchestration, config management, versioning)
