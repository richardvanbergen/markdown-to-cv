# Phase 1: Foundation & Executors - Research

**Researched:** 2026-02-03
**Domain:** Go subprocess execution, embedded assets, config management, preflight checks
**Confidence:** HIGH

## Summary

Phase 1 establishes the foundational infrastructure for m2cv: reliable subprocess execution (ClaudeExecutor, NPMExecutor), configuration management (m2cv.yml discovery), embedded asset handling (prompts + JSON Resume schema), and preflight checks (claude/resumed availability). This phase addresses critical subprocess patterns that prevent buffer deadlocks, PATH resolution issues across node version managers, and config file discovery following git-like walk-up patterns.

The research confirms that Go 1.25+ (current stable as of February 2026) provides all necessary capabilities through stdlib (`os/exec`, `embed`, `filepath`) with minimal external dependencies (cobra for CLI, yaml.v3 for config, jsonschema/v6 for validation).

**Primary recommendation:** Implement ClaudeExecutor with streaming stdout/stderr from day one (not `cmd.Output()`), use stdin piping for prompts to avoid shell argument limits, and create a robust PATH resolution system for npm/npx that checks common node version manager locations.

## Standard Stack

### Core Libraries

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| **Go** | 1.25+ | Implementation language | Current stable (Feb 2026); 1.24-1.25 both actively maintained. 1.25.6 latest patch. |
| **github.com/spf13/cobra** | v1.9.1+ | CLI framework | Industry standard (kubectl, gh, hugo). 1126 code snippets in Context7. Excellent support for subcommands, persistent flags, and PersistentPreRunE hooks. |
| **gopkg.in/yaml.v3** | v3.0+ | YAML config parsing | Most mature Go YAML library. Maintains insertion order, better error messages than v2. |
| **github.com/santhosh-tekuri/jsonschema** | v6.0+ | JSON Schema validation | Supports draft-07 (used by JSON Resume). Fast, mature API, good error messages. |

**Confidence:** HIGH - All libraries verified via Context7 and official sources. Versions confirmed current as of Feb 2026.

### External Dependencies (User-Installed)

| Tool | Version | Purpose | Detection Method |
|------|---------|---------|------------------|
| **claude** | latest | AI prompt execution | `exec.LookPath("claude")` in PersistentPreRunE |
| **node/npm** | 18+ LTS | Run resumed, install themes | `exec.LookPath("npm")` with fallback search |
| **npx** | (bundled with npm) | Run resumed without global install | `exec.LookPath("npx")` with fallback search |
| **resumed** | 6.1.0+ | PDF export from JSON Resume | Check project node_modules or global install |

**Notes:**
- **resumed 6.1.0** is latest as of October 2025 (4 months before Feb 2026). Actively maintained, ~400 JSON Resume themes available.
- **Claude CLI** requires Claude Pro subscription. Uses `-p` flag for print mode (non-interactive stdin processing).
- **Node 18+ LTS** ensures stability. Users manage their own npm environment.

**Confidence:** HIGH - resumed version verified from npm registry (Feb 2026), Claude CLI usage verified from official docs.

### Installation

```bash
# Go dependencies
go get github.com/spf13/cobra@latest
go get gopkg.in/yaml.v3@latest
go get github.com/santhosh-tekuri/jsonschema/v6@latest

# External tools (user-installed, verified in preflight)
# - claude CLI (from Anthropic, requires Pro subscription)
# - node/npm 18+ LTS
# - resumed installed via npm in project
```

## Architecture Patterns

### Pattern 1: Streaming Subprocess Execution (CRITICAL)

**What:** Always stream stdout/stderr from subprocesses using goroutines and buffers. Never use `cmd.Output()` for Claude calls.

**When to use:** All subprocess execution, especially Claude (produces kilobyte responses).

**Why critical:** Buffer deadlock is the #1 subprocess pitfall in Go. If stdout buffer fills (typically 64KB on Unix), subprocess blocks on write while Go blocks on `cmd.Wait()`.

**Example:**

```go
// BAD: Will deadlock on large output
cmd := exec.Command("claude", "-p")
cmd.Stdin = strings.NewReader(prompt)
output, err := cmd.Output() // Blocks until process exits, but process blocks on full buffer

// GOOD: Stream output concurrently
cmd := exec.Command("claude", "-p")
cmd.Stdin = strings.NewReader(prompt)

var stdout, stderr bytes.Buffer
cmd.Stdout = &stdout
cmd.Stderr = &stderr

if err := cmd.Start(); err != nil {
    return "", err
}

// Buffers are being filled concurrently as process runs
if err := cmd.Wait(); err != nil {
    return "", fmt.Errorf("claude failed: %w\nstderr: %s", err, stderr.String())
}

result := stdout.String()
```

**Sources:**
- [Go os/exec buffer deadlock issue #16787](https://github.com/golang/go/issues/16787)
- [Advanced command execution in Go with os/exec](https://blog.kowalczyk.info/article/wOYk/advanced-command-execution-in-go-with-osexec.html)

**Confidence:** HIGH - Well-documented Go subprocess pattern.

---

### Pattern 2: stdin Piping for Large Prompts

**What:** Pass prompts to Claude via stdin pipe instead of command arguments to avoid shell argument limits.

**When to use:** All Claude invocations (prompts can be several KB with CV + job description).

**Why:** Shell argument limits are typically 128KB-2MB depending on system. stdin has no such limit.

**Example:**

```go
// Use strings.NewReader or bytes.NewReader for stdin
cmd := exec.Command("claude", "-p", "--output-format", "text")
cmd.Stdin = strings.NewReader(largePrompt) // No argument limit

var stdout bytes.Buffer
cmd.Stdout = &stdout
cmd.Stderr = os.Stderr // Passthrough for debugging

if err := cmd.Run(); err != nil {
    return "", fmt.Errorf("claude execution failed: %w", err)
}

return stdout.String(), nil
```

**Alternative for file-based prompts:**

```go
// Write prompt to temp file, pass path to Claude
tmpFile, err := os.CreateTemp("", "prompt-*.txt")
if err != nil {
    return "", err
}
tmpPath := tmpFile.Name()

tmpFile.WriteString(prompt)
tmpFile.Close()

cmd := exec.Command("claude", "-p", tmpPath)
// ... execute ...

os.Remove(tmpPath) // Cleanup AFTER cmd.Run() completes
```

**Sources:**
- [Go os/exec StdinPipe documentation](https://pkg.go.dev/os/exec)
- [Proxying to a subcommand with Go](https://kevin.burke.dev/kevin/proxying-to-a-subcommand-with-go/)

**Confidence:** HIGH - Standard Go pattern for large subprocess input.

---

### Pattern 3: PATH Resolution with Node Version Manager Fallbacks

**What:** Custom executable search that checks `exec.LookPath()` first, then common node version manager locations.

**When to use:** Finding npm/npx executables before subprocess invocation.

**Why:** Go's `exec.LookPath()` uses different PATH than interactive shells. `~/.bashrc`/`~/.zshrc` aren't loaded for non-interactive shells. nvm/asdf/volta modify PATH in shell init scripts.

**Example:**

```go
func findNodeExecutable(name string) (string, error) {
    // 1. Try standard PATH first
    if path, err := exec.LookPath(name); err == nil {
        return path, nil
    }

    // 2. Check common node manager locations
    home, _ := os.UserHomeDir()
    candidates := []string{
        filepath.Join(home, ".nvm/current/bin", name),
        filepath.Join(home, ".volta/bin", name),
        filepath.Join(home, ".asdf/shims", name),
        filepath.Join(home, ".fnm/current/bin", name),
        "/usr/local/bin/" + name,
        "/opt/homebrew/bin/" + name, // Apple Silicon Mac
    }

    for _, candidate := range candidates {
        if _, err := os.Stat(candidate); err == nil {
            return candidate, nil
        }
    }

    return "", fmt.Errorf("%s not found in PATH or common locations. Please ensure Node.js is installed.", name)
}

// Usage in NPMExecutor
type NPMExecutor struct {
    npmPath string
    npxPath string
}

func NewNPMExecutor() (*NPMExecutor, error) {
    npmPath, err := findNodeExecutable("npm")
    if err != nil {
        return nil, err
    }

    npxPath, err := findNodeExecutable("npx")
    if err != nil {
        return nil, err
    }

    return &NPMExecutor{
        npmPath: npmPath,
        npxPath: npxPath,
    }, nil
}

func (e *NPMExecutor) Install(ctx context.Context, dir string, packages ...string) error {
    args := append([]string{"install"}, packages...)
    cmd := exec.CommandContext(ctx, e.npmPath, args...)
    cmd.Dir = dir
    return cmd.Run()
}
```

**Sources:**
- [Go exec.LookPath documentation](https://pkg.go.dev/os/exec)
- [Go Command PATH security](https://go.dev/blog/path-security)

**Confidence:** HIGH - Addresses known PATH resolution issues.

---

### Pattern 4: Config File Discovery (Walk-Up Directory Tree)

**What:** Search for `m2cv.yml` by walking up directory tree from current working directory, like git searches for `.git`.

**When to use:** Loading project configuration.

**Why:** Users run commands from subdirectories. Config should be discoverable without exact path.

**Example:**

```go
func findConfig() (string, error) {
    // 1. Explicit flag (--config path/to/config.yml)
    if configFlag != "" {
        return configFlag, nil
    }

    // 2. Environment variable
    if envConfig := os.Getenv("M2CV_CONFIG"); envConfig != "" {
        return envConfig, nil
    }

    // 3. Walk up directory tree (like git)
    dir, err := os.Getwd()
    if err != nil {
        return "", err
    }

    for {
        candidate := filepath.Join(dir, "m2cv.yml")
        if _, err := os.Stat(candidate); err == nil {
            return candidate, nil
        }

        parent := filepath.Dir(dir)
        if parent == dir {
            break // Reached filesystem root
        }
        dir = parent
    }

    // 4. User home directory fallback
    home, err := os.UserHomeDir()
    if err == nil {
        userConfig := filepath.Join(home, ".config", "m2cv", "config.yml")
        if _, err := os.Stat(userConfig); err == nil {
            return userConfig, nil
        }
    }

    return "", fmt.Errorf("no m2cv.yml found (searched: current dir → root, ~/.config/m2cv/)")
}
```

**Search order:**
1. `--config` flag (explicit override)
2. `M2CV_CONFIG` environment variable
3. Walk up from current directory to filesystem root
4. `~/.config/m2cv/config.yml` (user-level defaults)

**Sources:**
- [Go filepath package](https://pkg.go.dev/path/filepath)
- [Go filepath.Walk usage patterns](https://reintech.io/blog/guide-to-gos-path-filepath-package)

**Confidence:** HIGH - Standard pattern used by git, docker, etc.

---

### Pattern 5: Embedded Assets with embed.FS

**What:** Use Go 1.16+ `//go:embed` directive to bundle prompts and JSON Resume schema into binary.

**When to use:** Static assets needed at runtime (prompts, schema files).

**Why:** Single binary distribution, no external file dependencies, immutable assets.

**Example:**

```go
// internal/assets/assets.go
package assets

import (
    "embed"
    "io/fs"
)

//go:embed prompts/*.txt
var promptFS embed.FS

//go:embed schema/*.json
var schemaFS embed.FS

// GetPrompt reads an embedded prompt template
func GetPrompt(name string) ([]byte, error) {
    return promptFS.ReadFile("prompts/" + name + ".txt")
}

// GetSchema reads the embedded JSON Resume schema
func GetSchema(name string) ([]byte, error) {
    return schemaFS.ReadFile("schema/" + name)
}

// PromptFS returns the embedded filesystem for advanced usage
func PromptFS() fs.FS {
    sub, _ := fs.Sub(promptFS, "prompts")
    return sub
}
```

**Directory structure:**

```
internal/assets/
├── assets.go
├── prompts/
│   ├── extract-name.txt
│   ├── optimize.txt
│   ├── optimize-ats.txt
│   └── md-to-json-resume.txt
└── schema/
    └── resume.schema.json
```

**Best practices:**
- ✅ Use separate files (not string literals) for syntax highlighting and validation
- ✅ Implement `io/fs.FS` interface for compatibility with `text/template`, `net/http`
- ✅ embed.FS is read-only and goroutine-safe
- ⚠️ Files become part of binary (can be extracted) - don't embed secrets

**Sources:**
- [Go embed package documentation](https://pkg.go.dev/embed)
- [Embedded File Systems: Using embed.FS in Production](https://dev.to/rezmoss/embedded-file-systems-using-embedfs-in-production-89-2fpa)
- [Go Embedding Series: Advanced Usage of embed.FS](https://ehewen.com/en/blog/go-embedfs/)

**Confidence:** HIGH - Standard library feature since Go 1.16.

---

### Pattern 6: Cobra PersistentPreRunE for Preflight Checks

**What:** Use `PersistentPreRunE` hook in root command to verify external dependencies before any command runs.

**When to use:** Checking `claude` CLI availability before AI commands, `resumed` before generate.

**Why:** Fail fast with clear error messages instead of cryptic exec errors mid-operation.

**Example:**

```go
// cmd/root.go
package cmd

import (
    "fmt"
    "os/exec"

    "github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "m2cv",
        Short: "Markdown to CV - AI-powered resume tailoring",
        PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
            // Check claude CLI is in PATH
            if _, err := exec.LookPath("claude"); err != nil {
                return fmt.Errorf(`claude CLI not found in PATH

Install from: https://claude.ai/download
Requires: Claude Pro subscription

After installing, verify with: claude --version`)
            }

            return nil
        },
    }

    return cmd
}

// cmd/generate.go - additional check for specific command
func NewGenerateCommand(svc *service.GenerateService) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "generate <application-name>",
        Short: "Generate PDF resume from optimized CV",
        PreRunE: func(cmd *cobra.Command, args []string) error {
            // Check resumed is installed
            projectDir := "." // or from flag
            if err := checkResumedInstalled(projectDir); err != nil {
                return fmt.Errorf(`resumed not installed

Run 'm2cv init' to set up project and install resumed.
Or manually: npm install resumed`)
            }
            return nil
        },
        RunE: func(cmd *cobra.Command, args []string) error {
            // ... implementation
        },
    }

    return cmd
}

func checkResumedInstalled(projectDir string) error {
    // Check node_modules/resumed exists
    resumedPath := filepath.Join(projectDir, "node_modules", "resumed")
    if _, err := os.Stat(resumedPath); err == nil {
        return nil // Found in project
    }

    // Check if globally installed
    if _, err := exec.LookPath("resumed"); err == nil {
        return nil // Found globally
    }

    return fmt.Errorf("resumed not found")
}
```

**Execution order (from Cobra docs):**
1. `PersistentPreRun` (parent command)
2. `PreRun` (current command)
3. `Run` (current command)
4. `PostRun` (current command)
5. `PersistentPostRun` (parent command)

**Key insight:** `PersistentPreRunE` runs for ALL subcommands, making it perfect for global dependency checks.

**Sources:**
- [Cobra PreRun and PostRun Hooks](https://github.com/spf13/cobra/blob/main/site/content/user_guide.md)
- Context7: /spf13/cobra

**Confidence:** HIGH - Standard Cobra pattern for preflight checks.

---

### Pattern 7: Repository Interfaces for Testability

**What:** Abstract filesystem and config operations behind interfaces.

**When to use:** All file I/O, config loading, versioned file management.

**Why:** Enables testing with in-memory implementations, centralizes path logic.

**Example:**

```go
// internal/persistence/interfaces.go
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

type FileRepository interface {
    CreateAppFolder(ctx context.Context, projectDir, appName string) error
    WriteFile(ctx context.Context, path, content string) error
    ReadFile(ctx context.Context, path string) (string, error)
    GetNextVersion(ctx context.Context, projectDir, appName, prefix string) (int, error)
}

// internal/persistence/config.go
type yamlConfigRepo struct{}

func NewYAMLConfigRepository() ConfigRepository {
    return &yamlConfigRepo{}
}

func (r *yamlConfigRepo) Load(ctx context.Context, projectDir string) (*Config, error) {
    // Use findConfig() pattern to locate m2cv.yml
    configPath, err := findConfigInDir(projectDir)
    if err != nil {
        return nil, err
    }

    data, err := os.ReadFile(configPath)
    if err != nil {
        return nil, fmt.Errorf("read config: %w", err)
    }

    var cfg Config
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("parse config: %w", err)
    }

    return &cfg, nil
}
```

**Confidence:** HIGH - Standard Go repository pattern.

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| **CLI framework** | Custom flag parsing + subcommand routing | cobra | Handles subcommands, persistent flags, help generation, shell completion. 1100+ code examples. |
| **YAML parsing** | Custom config parser | yaml.v3 | Handles edge cases (multi-line, anchors, types). Insertion order preservation. |
| **JSON Schema validation** | Custom validation logic | jsonschema/v6 | Supports draft-07, comprehensive error messages, handles all schema features. |
| **Subprocess streaming** | Custom buffer management | stdlib patterns (bytes.Buffer + goroutines) | Prevents deadlocks, handles stderr correctly. |
| **PDF export from JSON Resume** | Custom PDF generator | resumed + JSON Resume themes | 400+ themes, actively maintained, handles all JSON Resume features. |

**Key insight:** Go stdlib (`os/exec`, `embed`, `filepath`) + cobra + YAML is sufficient. Avoid adding heavy dependencies (viper, etc.) for simple config use case.

## Common Pitfalls

### Pitfall 1: Unbuffered Subprocess Output (CRITICAL)

**What goes wrong:** Using `cmd.Output()` with Claude subprocess causes deadlocks on large responses (>64KB).

**Why it happens:** stdout buffer fills, subprocess blocks on write, Go blocks on `cmd.Wait()`. Deadlock.

**How to avoid:**
```go
// ALWAYS use streaming pattern
var stdout, stderr bytes.Buffer
cmd.Stdout = &stdout
cmd.Stderr = &stderr
cmd.Start()
cmd.Wait() // Buffers fill concurrently
```

**Warning signs:**
- CLI hangs during "Waiting for Claude..." step
- Works with short prompts, hangs with long responses

**Sources:** [Go issue #16787](https://github.com/golang/go/issues/16787)

**Confidence:** HIGH

---

### Pitfall 2: PATH Resolution Failures with Node Version Managers

**What goes wrong:** `exec.LookPath("npx")` fails even though `which npx` works in shell.

**Why it happens:** Go doesn't load `~/.bashrc`/`~/.zshrc`. nvm/volta/asdf modify PATH in shell init.

**How to avoid:** Implement `findNodeExecutable()` with fallback search (Pattern 3 above).

**Warning signs:**
- Works locally, fails in CI
- "npx: command not found" but `which npx` succeeds

**Confidence:** HIGH - Known Go subprocess issue.

---

### Pitfall 3: Temp File Cleanup Race Condition

**What goes wrong:** Using `defer os.Remove(tempFile)` immediately after creating temp prompt file. File gets deleted while Claude is reading it.

**Why it happens:** `defer` fires when function returns, but subprocess might still be running.

**How to avoid:**
```go
// BAD: Defer fires too early
tmpFile, _ := os.CreateTemp("", "prompt-*.txt")
defer os.Remove(tmpFile.Name()) // WRONG
tmpFile.WriteString(prompt)
tmpFile.Close()
cmd := exec.Command("claude", "-p", tmpFile.Name())
return cmd.Run() // defer fires here, file deleted while Claude reading

// GOOD: Cleanup after subprocess completes
tmpFile, _ := os.CreateTemp("", "prompt-*.txt")
tmpPath := tmpFile.Name()
tmpFile.WriteString(prompt)
tmpFile.Close()

cmd := exec.Command("claude", "-p", tmpPath)
err := cmd.Run() // Wait for completion

os.Remove(tmpPath) // NOW safe to delete
return err
```

**Warning signs:**
- Intermittent "file not found" errors
- More failures under concurrent load

**Confidence:** HIGH - Classic Go timing issue.

---

### Pitfall 4: Hardcoded Config Paths

**What goes wrong:** Only looking for `m2cv.yml` in current directory.

**Why it happens:** Simple implementation: `os.ReadFile("m2cv.yml")`

**How to avoid:** Use walk-up directory tree pattern (Pattern 4 above).

**Warning signs:**
- Users complain about "file not found" from subdirectories

**Confidence:** HIGH

---

### Pitfall 5: No JSON Extraction from Claude Output

**What goes wrong:** Directly unmarshaling Claude output with `json.Unmarshal()`. Claude wraps JSON in markdown fences, adds commentary.

**Why it happens:** LLMs are non-deterministic. Claude might return:
- `\`\`\`json\n{...}\n\`\`\`` (markdown fences)
- Text + JSON mixed
- Comments (not valid JSON)

**How to avoid:**

```go
func extractJSON(claudeOutput []byte) (json.RawMessage, error) {
    // 1. Strip markdown fences
    content := stripMarkdownCodeFences(claudeOutput)

    // 2. Find JSON boundaries
    start := bytes.IndexByte(content, '{')
    end := bytes.LastIndexByte(content, '}')
    if start == -1 || end == -1 {
        return nil, fmt.Errorf("no JSON object found in:\n%s", claudeOutput)
    }

    extracted := content[start:end+1]

    // 3. Validate it's parseable
    var raw json.RawMessage
    if err := json.Unmarshal(extracted, &raw); err != nil {
        return nil, fmt.Errorf("invalid JSON: %w\nExtracted:\n%s", err, extracted)
    }

    return raw, nil
}

func stripMarkdownCodeFences(data []byte) []byte {
    // Remove ```json...``` fences
    re := regexp.MustCompile("(?s)```(?:json)?\n(.*?)\n```")
    matches := re.FindSubmatch(data)
    if matches != nil {
        return matches[1]
    }
    return data
}
```

**Warning signs:**
- Intermittent "invalid character" errors
- Different errors on retry with same input

**Confidence:** HIGH - Known LLM integration challenge.

---

### Pitfall 6: Missing stderr Passthrough

**What goes wrong:** Only capturing stdout, ignoring stderr. Claude errors are invisible to user.

**Why it happens:** Focusing only on success case.

**How to avoid:**

```go
var stdout, stderr bytes.Buffer
cmd.Stdout = &stdout
cmd.Stderr = &stderr

if err := cmd.Run(); err != nil {
    return "", fmt.Errorf("claude failed: %w\nstderr: %s", err, stderr.String())
}
```

**Confidence:** HIGH

## Code Examples

### Complete ClaudeExecutor Implementation

```go
// internal/executor/claude.go
package executor

import (
    "bytes"
    "context"
    "fmt"
    "os/exec"
    "strings"
)

type ClaudeExecutor interface {
    Execute(ctx context.Context, prompt string, opts ...Option) (string, error)
}

type claudeExecutor struct {
    binaryPath string
}

type Option func(*executeOptions)

type executeOptions struct {
    Model        string
    OutputFormat string
}

func WithModel(model string) Option {
    return func(o *executeOptions) { o.Model = model }
}

func NewClaudeExecutor() (ClaudeExecutor, error) {
    // Verify claude is in PATH
    path, err := exec.LookPath("claude")
    if err != nil {
        return nil, fmt.Errorf("claude CLI not found in PATH")
    }

    return &claudeExecutor{binaryPath: path}, nil
}

func (e *claudeExecutor) Execute(ctx context.Context, prompt string, opts ...Option) (string, error) {
    options := &executeOptions{
        OutputFormat: "text",
    }
    for _, opt := range opts {
        opt(options)
    }

    args := []string{"-p", "--output-format", options.OutputFormat}
    if options.Model != "" {
        args = append(args, "-m", options.Model)
    }

    cmd := exec.CommandContext(ctx, e.binaryPath, args...)
    cmd.Stdin = strings.NewReader(prompt)

    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    if err := cmd.Start(); err != nil {
        return "", fmt.Errorf("start claude: %w", err)
    }

    if err := cmd.Wait(); err != nil {
        return "", fmt.Errorf("claude failed: %w\nstderr: %s", err, stderr.String())
    }

    return stdout.String(), nil
}
```

**Source:** Synthesized from Go os/exec patterns and Claude CLI documentation.

---

### Complete NPMExecutor Implementation

```go
// internal/executor/npm.go
package executor

import (
    "context"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
)

type NPMExecutor interface {
    Init(ctx context.Context, dir string) error
    Install(ctx context.Context, dir string, packages ...string) error
    CheckInstalled(ctx context.Context, dir string, pkg string) (bool, error)
}

type npmExecutor struct {
    npmPath string
    npxPath string
}

func NewNPMExecutor() (NPMExecutor, error) {
    npmPath, err := findNodeExecutable("npm")
    if err != nil {
        return nil, err
    }

    npxPath, err := findNodeExecutable("npx")
    if err != nil {
        return nil, err
    }

    return &npmExecutor{
        npmPath: npmPath,
        npxPath: npxPath,
    }, nil
}

func (e *npmExecutor) Init(ctx context.Context, dir string) error {
    cmd := exec.CommandContext(ctx, e.npmPath, "init", "-y")
    cmd.Dir = dir
    return cmd.Run()
}

func (e *npmExecutor) Install(ctx context.Context, dir string, packages ...string) error {
    args := append([]string{"install"}, packages...)
    cmd := exec.CommandContext(ctx, e.npmPath, args...)
    cmd.Dir = dir

    var stderr bytes.Buffer
    cmd.Stderr = &stderr

    if err := cmd.Run(); err != nil {
        return fmt.Errorf("npm install failed: %w\nstderr: %s", err, stderr.String())
    }

    return nil
}

func (e *npmExecutor) CheckInstalled(ctx context.Context, dir string, pkg string) (bool, error) {
    pkgPath := filepath.Join(dir, "node_modules", pkg)
    _, err := os.Stat(pkgPath)
    if err == nil {
        return true, nil
    }
    if os.IsNotExist(err) {
        return false, nil
    }
    return false, err
}

func findNodeExecutable(name string) (string, error) {
    // 1. Try LookPath first
    if path, err := exec.LookPath(name); err == nil {
        return path, nil
    }

    // 2. Check common node manager locations
    home, _ := os.UserHomeDir()
    candidates := []string{
        filepath.Join(home, ".nvm/current/bin", name),
        filepath.Join(home, ".volta/bin", name),
        filepath.Join(home, ".asdf/shims", name),
        filepath.Join(home, ".fnm/current/bin", name),
        "/usr/local/bin/" + name,
        "/opt/homebrew/bin/" + name, // Apple Silicon
    }

    for _, candidate := range candidates {
        if _, err := os.Stat(candidate); err == nil {
            return candidate, nil
        }
    }

    return "", fmt.Errorf("%s not found in PATH or common locations", name)
}
```

---

### ConfigRepository Implementation

```go
// internal/persistence/config.go
package persistence

import (
    "context"
    "fmt"
    "os"
    "path/filepath"

    "gopkg.in/yaml.v3"
)

type Config struct {
    BaseCVPath   string   `yaml:"base_cv_path"`
    DefaultTheme string   `yaml:"default_theme"`
    Themes       []string `yaml:"themes"`
    DefaultModel string   `yaml:"default_model"`
}

type ConfigRepository interface {
    Load(ctx context.Context, projectDir string) (*Config, error)
    Save(ctx context.Context, projectDir string, cfg *Config) error
    Find() (string, error)
}

type yamlConfigRepo struct{}

func NewYAMLConfigRepository() ConfigRepository {
    return &yamlConfigRepo{}
}

func (r *yamlConfigRepo) Find() (string, error) {
    // Walk up directory tree
    dir, err := os.Getwd()
    if err != nil {
        return "", err
    }

    for {
        candidate := filepath.Join(dir, "m2cv.yml")
        if _, err := os.Stat(candidate); err == nil {
            return candidate, nil
        }

        parent := filepath.Dir(dir)
        if parent == dir {
            break // Reached root
        }
        dir = parent
    }

    return "", fmt.Errorf("m2cv.yml not found (searched from %s to root)", dir)
}

func (r *yamlConfigRepo) Load(ctx context.Context, projectDir string) (*Config, error) {
    configPath := filepath.Join(projectDir, "m2cv.yml")
    data, err := os.ReadFile(configPath)
    if err != nil {
        return nil, fmt.Errorf("read config: %w", err)
    }

    var cfg Config
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("parse config: %w", err)
    }

    return &cfg, nil
}

func (r *yamlConfigRepo) Save(ctx context.Context, projectDir string, cfg *Config) error {
    configPath := filepath.Join(projectDir, "m2cv.yml")
    data, err := yaml.Marshal(cfg)
    if err != nil {
        return fmt.Errorf("marshal config: %w", err)
    }

    if err := os.WriteFile(configPath, data, 0644); err != nil {
        return fmt.Errorf("write config: %w", err)
    }

    return nil
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `cmd.Output()` for subprocesses | Stream with `bytes.Buffer` + goroutines | Always been best practice | Prevents buffer deadlocks |
| String literals for embedded assets | `//go:embed` directive | Go 1.16 (Feb 2021) | Single binary, cleaner code |
| Custom flag parsing | cobra framework | N/A (cobra standard since ~2015) | Shell completion, help generation |
| `filepath.Walk()` | `filepath.WalkDir()` | Go 1.16 | Faster directory traversal |
| resume-cli | resumed | ~2020 | Active maintenance, faster |

**Deprecated/outdated:**
- **resume-cli**: Unmaintained since 2017. Use resumed instead.
- **cmd.Output() for large output**: Causes deadlocks. Always stream.

**Confidence:** HIGH

## Open Questions

### 1. Claude CLI Version/Flags Stability

**What we know:** Claude CLI uses `-p` flag for print mode, `--output-format text`, `-m` for model selection (as of Feb 2026 docs).

**What's unclear:** Long-term API stability. Will flags change in future Claude CLI versions?

**Recommendation:**
- Document supported Claude CLI version in README
- Add version check in preflight if possible
- Wrap Claude invocation in executor interface for easy adaptation

**Confidence:** MEDIUM - CLI is actively developed, flags may evolve.

---

### 2. resumed Package Stability

**What we know:** resumed 6.1.0 is latest (Oct 2025). Actively maintained fork of resume-cli.

**What's unclear:** API stability for `resumed export` command.

**Recommendation:**
- Pin resumed version in documentation
- Test with latest version during Phase 5 (Export Pipeline)
- Consider version check in preflight

**Confidence:** MEDIUM - npm package ecosystem changes.

---

### 3. JSON Resume Schema Version

**What we know:** JSON Resume uses JSON Schema draft-07. jsonschema/v6 supports draft-07.

**What's unclear:** Current JSON Resume schema location and exact version.

**Recommendation:**
- Fetch latest schema from https://github.com/jsonresume/resume-schema during Phase 1
- Embed it in binary via `//go:embed`
- Update schema periodically

**Confidence:** MEDIUM - Need to verify current schema URL.

---

### 4. Optimal Config Discovery Strategy

**What we know:** Walk-up pattern works (git, docker use it).

**What's unclear:** Should we support XDG_CONFIG_HOME? Multiple config locations with merge?

**Recommendation:**
- Start simple: walk-up tree only
- Add environment variable override (`M2CV_CONFIG`)
- Defer XDG_CONFIG_HOME to future iteration if users request

**Confidence:** HIGH - Simple approach sufficient for v1.

## Sources

### Primary (HIGH confidence)

**Libraries & Frameworks:**
- [Cobra v1.9.1 - Context7](https://github.com/spf13/cobra) - 1126 code snippets, High reputation
- [Go embed package](https://pkg.go.dev/embed) - Official Go documentation
- [Go os/exec package](https://pkg.go.dev/os/exec) - Official Go documentation
- [jsonschema/v6](https://github.com/santhosh-tekuri/jsonschema) - GitHub, High reputation (Context7)
- [yaml.v3](https://pkg.go.dev/gopkg.in/yaml.v3) - Official package documentation

**Go Version:**
- [Go Release History](https://go.dev/doc/devel/release) - Official Go releases
- [Go 1.24 Release Notes](https://go.dev/blog/go1.24) - Official blog
- [Go 1.25 Release Notes](https://go.dev/doc/go1.24) - Confirmed 1.25.6 current (Jan 2026)

**Subprocess Patterns:**
- [Go issue #16787: io/ioutil hangs with too big output](https://github.com/golang/go/issues/16787)
- [Advanced command execution in Go with os/exec](https://blog.kowalczyk.info/article/wOYk/advanced-command-execution-in-go-with-osexec.html)
- [Go Command PATH security](https://go.dev/blog/path-security)

**embed.FS:**
- [Embedded File Systems: Using embed.FS in Production](https://dev.to/rezmoss/embedded-file-systems-using-embedfs-in-production-89-2fpa)
- [Go Embedding Series: Advanced Usage](https://ehewen.com/en/blog/go-embedfs/)

**Claude CLI:**
- [Claude CLI reference](https://code.claude.com/docs/en/cli-reference)
- [Claude Code CLI Cheatsheet](https://shipyard.build/blog/claude-code-cheat-sheet/)
- [Claude CLI stdin bug #7263](https://github.com/anthropics/claude-code/issues/7263)

**resumed:**
- [resumed on npm](https://www.npmjs.com/package/resumed) - 6.1.0 (Oct 2025)
- [resumed GitHub](https://github.com/rbardini/resumed) - Official repository

### Secondary (MEDIUM confidence)

- [filepath package usage](https://reintech.io/blog/guide-to-gos-path-filepath-package)
- [Proxying to subcommand with Go](https://kevin.burke.dev/kevin/proxying-to-a-subcommand-with-go/)

### Tertiary (LOW confidence)

- Custom markdown parsing approach (no authoritative source, architectural decision)

## Metadata

**Confidence breakdown:**
- Standard stack: **HIGH** - All libraries verified via Context7/official sources, versions confirmed current
- Architecture patterns: **HIGH** - Subprocess streaming, stdin piping, embed.FS are well-documented Go patterns
- Pitfalls: **HIGH** - Buffer deadlock, PATH resolution, temp file cleanup are known Go issues with established solutions
- Claude CLI integration: **MEDIUM** - CLI is active development, flags may evolve
- resumed integration: **MEDIUM** - npm package ecosystem, version pinning needed

**Research date:** 2026-02-03
**Valid until:** ~2026-05-03 (90 days - Go ecosystem stable, Claude CLI may change faster)

**Overall confidence:** HIGH

Phase 1 implementation can proceed with high confidence. All critical patterns (subprocess execution, config loading, embedded assets, preflight checks) are well-established Go practices with authoritative sources.
