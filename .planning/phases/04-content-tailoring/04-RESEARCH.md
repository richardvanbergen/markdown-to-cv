# Phase 4: Content Tailoring - Research

**Researched:** 2026-02-03
**Domain:** Go CLI command implementation with AI subprocess integration
**Confidence:** HIGH

## Summary

Phase 4 implements the `m2cv optimize` command which reads a base CV and job description, calls Claude CLI to generate a tailored CV, and writes versioned output. This phase leverages existing infrastructure (ClaudeExecutor, versioning utilities, config repository) and primarily needs command structure, file I/O, and prompt template handling.

The standard approach for this phase uses Go's built-in text/template for prompt rendering (with simple placeholders), os.ReadFile for reading input files (base CV and job description are typically small), and the existing versioning.NextVersionPath pattern for auto-incrementing output files. The ClaudeExecutor is already established with proper subprocess handling via stdin piping.

**Primary recommendation:** Use strings.Replace for simple prompt variable substitution ({{.BaseCV}}, {{.JobDescription}}) following the existing pattern in extractor/folder_name.go. This is faster and simpler than text/template for the two-variable case. Focus testing on flag combinations, file reading errors, and versioning edge cases.

## Standard Stack

The established libraries/tools for this domain:

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| spf13/cobra | v1.10.2 | CLI framework | Already in use, proven pattern in codebase |
| os/filepath | stdlib | Path manipulation | Standard for cross-platform file paths |
| os | stdlib | File I/O operations | Standard for reading/writing files |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| gopkg.in/yaml.v3 | v3.0.1 | YAML parsing | Already in use for config, may be needed for CV frontmatter validation |
| context | stdlib | Timeout/cancellation | Already used in ClaudeExecutor for subprocess control |
| strings | stdlib | Simple template substitution | For {{.Variable}} replacement in prompts |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| strings.Replace | text/template | Template is 2x slower but more maintainable for >3 variables; use strings.Replace for 2 variables |
| os.ReadFile | bufio.Scanner | Scanner is better for large files (>10MB); CVs/job descriptions are typically <100KB |
| filepath.Join | string concatenation | String concat breaks on Windows; always use filepath.Join |

**Installation:**
All packages are in stdlib or already in go.mod - no new dependencies needed.

## Architecture Patterns

### Recommended Project Structure
```
cmd/
├── optimize.go           # Command definition and flag parsing
└── optimize_test.go      # Command integration tests

internal/
├── application/
│   └── versioning.go     # Already exists - version detection
├── executor/
│   └── claude.go         # Already exists - subprocess execution
├── config/
│   └── config.go         # Already exists - load m2cv.yml
└── assets/
    └── prompts/
        ├── optimize.txt     # Already exists - standard prompt
        └── optimize-ats.txt # Already exists - ATS mode prompt
```

### Pattern 1: Cobra Command with Flag Binding
**What:** Bind flags to variables, validate in PreRunE, execute in RunE
**When to use:** All Cobra commands in this codebase
**Example:**
```go
// Source: cmd/apply.go (existing pattern)
func newOptimizeCommand() *cobra.Command {
    var (
        model string
        atsMode bool
    )

    cmd := &cobra.Command{
        Use:   "optimize <application-name>",
        Short: "Tailor CV to job description with AI",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            return runOptimize(cmd.Context(), args[0], model, atsMode)
        },
    }

    cmd.Flags().StringVarP(&model, "model", "m", "", "override Claude model")
    cmd.Flags().BoolVar(&atsMode, "ats", false, "optimize for ATS systems")

    return cmd
}
```

### Pattern 2: Simple Template Substitution with strings.Replace
**What:** Use strings.Replace for prompts with 1-3 variables
**When to use:** Simple {{.Variable}} substitution without complex logic
**Example:**
```go
// Source: internal/extractor/folder_name.go (existing pattern)
prompt := strings.ReplaceAll(promptTemplate, "{{.JobDescription}}", jobDesc)
prompt = strings.ReplaceAll(prompt, "{{.BaseCV}}", baseCV)
```
**Why this pattern:** 2x faster than text/template, simpler code, sufficient for fixed placeholders. Performance benchmarks show strings.Replace at ~289 ns/op vs text/template at ~6373 ns/op for simple substitution.

### Pattern 3: File Reading with os.ReadFile
**What:** Read entire file into memory with os.ReadFile
**When to use:** Files <1MB (typical CVs and job descriptions)
**Example:**
```go
// Source: cmd/apply.go line 54 (existing pattern)
content, err := os.ReadFile(jobFile)
if err != nil {
    return fmt.Errorf("failed to read job description: %w", err)
}
```
**Why this pattern:** Simple, efficient for small files. bufio.Scanner only needed for >10MB files.

### Pattern 4: Config Discovery with Walk-Up
**What:** Use config.FindWithOverrides to locate m2cv.yml
**When to use:** Commands that need base_cv_path or default_model from config
**Example:**
```go
// Source: internal/config/config.go (existing pattern)
configPath, err := config.FindWithOverrides(cfgFile, ".")
if err != nil {
    return fmt.Errorf("m2cv.yml not found: %w", err)
}

cfg, err := configRepo.Load(configPath)
if err != nil {
    return fmt.Errorf("failed to load config: %w", err)
}
```

### Pattern 5: Version Auto-Increment
**What:** Use application.NextVersionPath for output filename
**When to use:** All optimize command executions
**Example:**
```go
// Source: internal/application/versioning.go (existing utility)
outputPath, err := application.NextVersionPath(appDir)
if err != nil {
    return fmt.Errorf("failed to determine version: %w", err)
}
// outputPath will be "appDir/optimized-cv-N.md" where N auto-increments
```

### Anti-Patterns to Avoid
- **Don't parse templates on every execution:** Load prompt templates once via assets.GetPrompt, not on each command run (though with embed.FS this is fast, it's still wasteful)
- **Don't use filepath.Abs unnecessarily:** Only use when you need to verify paths; filepath.Join handles relative paths correctly
- **Don't forget context:** Always pass cmd.Context() to ClaudeExecutor.Execute for proper cancellation support
- **Don't hard-code model names:** Load default_model from config, allow -m flag to override

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Version detection | Manual regex on directory listing | application.ListVersions, NextVersionPath | Already handles edge cases (malformed names, gaps, sorting) |
| Config file discovery | Recursive directory walk | config.FindWithOverrides | Handles flag override, env var, walk-up discovery order |
| Subprocess execution | exec.Command with .Run() | executor.ClaudeExecutor | Already handles stdin piping, buffer deadlock prevention, error messages |
| Filename sanitization | Custom string cleaning | extractor.SanitizeFilename | Already handles special chars, length limits, word boundaries |
| Path security | String prefix checks | filepath.IsLocal (Go 1.20+) | Prevents path traversal attacks; pure lexical check |

**Key insight:** The codebase already has utilities for versioning, config, and execution. The optimize command is primarily orchestration of existing components.

## Common Pitfalls

### Pitfall 1: Buffer Deadlock with Subprocess Output
**What goes wrong:** Using cmd.Run() or cmd.Output() with large Claude responses causes deadlock when stdout buffer fills
**Why it happens:** cmd.Output() reads stdout after process completes; if stdout buffer fills before process finishes, deadlock occurs
**How to avoid:** Use the existing ClaudeExecutor which uses cmd.Start() + cmd.Wait() with bytes.Buffer (Pattern 1: streaming subprocess execution in executor/claude.go)
**Warning signs:** Command hangs when Claude produces >64KB output; process never completes

### Pitfall 2: Missing Job Description File in Application Folder
**What goes wrong:** User runs `m2cv optimize app-name` but no job description exists in applications/app-name/
**Why it happens:** Job description is copied during `m2cv apply` but user may delete it or use manual folder creation
**How to avoid:** Check for job-description.txt or *.txt files in folder; error with helpful message if none found
**Warning signs:** Empty prompt sent to Claude; optimization output is generic

### Pitfall 3: Model Override Not Passed to Executor
**What goes wrong:** User specifies -m flag but default model from config is used instead
**Why it happens:** Flag variable not passed to executor.WithModel() option
**How to avoid:** Check if flag is non-empty before deciding between flag and config model
**Warning signs:** Tests with -m flag don't actually change which model is called

### Pitfall 4: Path Traversal in Application Name
**What goes wrong:** User runs `m2cv optimize ../../etc/passwd` and command accesses files outside applications/
**Why it happens:** Application name is concatenated directly to applications/ without validation
**How to avoid:** Use filepath.Join and validate result with filepath.IsLocal or check path doesn't start with ".."
**Warning signs:** Security scanner flags path concatenation; could read arbitrary files

### Pitfall 5: Config Loading Fails Silently
**What goes wrong:** m2cv.yml not found but command continues with empty config; base CV path is empty string
**Why it happens:** Config loading error is ignored or default values are used
**How to avoid:** Require m2cv.yml to exist for optimize command (unlike apply which doesn't need config); return clear error
**Warning signs:** os.ReadFile("") returns "file not found" instead of "config not found"

### Pitfall 6: Forgetting Context for Cancellation
**What goes wrong:** User hits Ctrl+C but Claude process keeps running for minutes
**Why it happens:** context.Background() used instead of cmd.Context() when calling executor
**How to avoid:** Always pass cmd.Context() to ClaudeExecutor.Execute
**Warning signs:** SIGINT doesn't stop the command; zombie Claude processes after cancel

## Code Examples

Verified patterns from official sources:

### Reading Base CV from Config Path
```go
// Source: internal/config/config.go + Go filepath docs
configPath, err := config.FindWithOverrides(cfgFile, ".")
if err != nil {
    return fmt.Errorf("m2cv.yml not found: %w. Run 'm2cv init' first", err)
}

cfg, err := configRepo.Load(configPath)
if err != nil {
    return fmt.Errorf("failed to load config: %w", err)
}

// Resolve base CV path (may be relative to config)
baseCVPath := cfg.BaseCVPath
if !filepath.IsAbs(baseCVPath) {
    configDir := filepath.Dir(configPath)
    baseCVPath = filepath.Join(configDir, baseCVPath)
}

baseCV, err := os.ReadFile(baseCVPath)
if err != nil {
    return fmt.Errorf("failed to read base CV at %s: %w", baseCVPath, err)
}
```

### Determining Application Folder Path
```go
// Source: cmd/apply.go pattern + filepath.Join behavior
appDir := filepath.Join("applications", applicationName)

// Validate app folder exists
if _, err := os.Stat(appDir); os.IsNotExist(err) {
    return fmt.Errorf("application folder not found: %s. Run 'm2cv apply' first", appDir)
}
```

### Finding Job Description in Application Folder
```go
// Source: Go filepath.Glob documentation
pattern := filepath.Join(appDir, "*.txt")
matches, err := filepath.Glob(pattern)
if err != nil {
    return fmt.Errorf("glob pattern error: %w", err)
}

if len(matches) == 0 {
    return fmt.Errorf("no .txt file found in %s. Job description required", appDir)
}

// Use first match (or could list all and ask user)
jobDescPath := matches[0]
jobDesc, err := os.ReadFile(jobDescPath)
if err != nil {
    return fmt.Errorf("failed to read job description: %w", err)
}
```

### Building Prompt with Variable Substitution
```go
// Source: internal/extractor/folder_name.go lines 79-89
promptTemplate, err := assets.GetPrompt("optimize") // or "optimize-ats"
if err != nil {
    return fmt.Errorf("failed to load prompt template: %w", err)
}

prompt := strings.ReplaceAll(promptTemplate, "{{.BaseCV}}", string(baseCV))
prompt = strings.ReplaceAll(prompt, "{{.JobDescription}}", string(jobDesc))
```

### Executing Claude with Model Override
```go
// Source: internal/executor/claude.go lines 56-61, 79-86
exec := executor.NewClaudeExecutor()

// Determine model: flag takes precedence over config
model := cfg.DefaultModel
if modelFlag != "" {
    model = modelFlag
}

var opts []executor.ExecuteOption
if model != "" {
    opts = append(opts, executor.WithModel(model))
}

result, err := exec.Execute(ctx, prompt, opts...)
if err != nil {
    return fmt.Errorf("Claude execution failed: %w", err)
}
```

### Writing Versioned Output
```go
// Source: internal/application/versioning.go lines 76-91
outputPath, err := application.NextVersionPath(appDir)
if err != nil {
    return fmt.Errorf("failed to determine version: %w", err)
}

if err := os.WriteFile(outputPath, []byte(result), 0644); err != nil {
    return fmt.Errorf("failed to write optimized CV: %w", err)
}

fmt.Printf("Optimized CV written to: %s\n", outputPath)
```

### Testing with Mocked Executor
```go
// Source: cmd/apply_test.go pattern (disable preflight) + testing best practices
func TestOptimizeCommand(t *testing.T) {
    t.Parallel()

    // Setup test directory structure
    tmpDir := t.TempDir()
    appDir := filepath.Join(tmpDir, "applications", "test-app")
    os.MkdirAll(appDir, 0755)

    // Create test files
    os.WriteFile(filepath.Join(appDir, "job.txt"), []byte("job desc"), 0644)

    rootCmd := NewRootCommand()
    rootCmd.AddCommand(newOptimizeCommand())
    rootCmd.SetArgs([]string{"optimize", "test-app", "--config", configPath})

    // Disable preflight checks for testing
    rootCmd.PersistentPreRunE = nil

    if err := rootCmd.Execute(); err != nil {
        t.Errorf("optimize command failed: %v", err)
    }

    // Verify output file exists
    outputPath := filepath.Join(appDir, "optimized-cv-1.md")
    if _, err := os.Stat(outputPath); os.IsNotExist(err) {
        t.Errorf("output file not created at %s", outputPath)
    }
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| text/template for all templates | strings.Replace for simple substitution | Go 1.0+ (always valid) | 2x faster for 1-3 variables |
| cmd.Output() for subprocess | cmd.Start() + bytes.Buffer | Pattern established in Phase 1 | Prevents deadlock on large output |
| HasPrefix for path checks | filepath.IsLocal | Go 1.20 (Feb 2023) | Secure path validation |
| Manual directory traversal | config.FindWithOverrides | Pattern established in Phase 2 | Consistent config discovery |
| filepath.Glob with manual parsing | application.ListVersions | Pattern established in Phase 3 | Handles edge cases correctly |

**Deprecated/outdated:**
- **ioutil package**: Deprecated in Go 1.16 (Feb 2021); use os.ReadFile, os.WriteFile instead
- **filepath.HasPrefix**: Not safe for path validation; use filepath.IsLocal or filepath.Rel checks
- **cmd.Run() for long output**: Works for small output but deadlocks on large; always use Start() + Wait() pattern

## Open Questions

Things that couldn't be fully resolved:

1. **Job Description Filename Convention**
   - What we know: apply command copies job file with original name (e.g., job.txt, posting.md)
   - What's unclear: Should optimize search for all *.txt files, specific name (job-description.txt), or first match?
   - Recommendation: Use filepath.Glob("*.txt") and take first match; this is most flexible

2. **Multiple Job Descriptions in One Folder**
   - What we know: One application folder may contain multiple job postings
   - What's unclear: Should optimize use all of them (concatenated) or just first match?
   - Recommendation: Use first match for initial implementation; document in help text

3. **Base CV Frontmatter Preservation**
   - What we know: Prompts say "maintain markdown format with YAML frontmatter"
   - What's unclear: Does Claude reliably preserve frontmatter, or should we parse and merge?
   - Recommendation: Trust Claude output initially; add frontmatter validation in verification if needed

4. **ATS Flag Behavior with Model Override**
   - What we know: ATS mode uses different prompt (optimize-ats.txt vs optimize.txt)
   - What's unclear: Should ATS mode force a specific model (e.g., sonnet for keyword extraction)?
   - Recommendation: Let user override model regardless of ATS flag; prompt handles optimization strategy

## Sources

### Primary (HIGH confidence)
- [Go text/template Package](https://pkg.go.dev/text/template) - Official documentation for template API
- [Go path/filepath Package](https://pkg.go.dev/path/filepath) - Official documentation for path manipulation
- [Cobra Command Framework](https://pkg.go.dev/github.com/spf13/cobra) - Official docs for v1.10.2 (in use)
- Existing codebase patterns:
  - cmd/apply.go - Command structure pattern
  - internal/executor/claude.go - Subprocess execution pattern
  - internal/application/versioning.go - Version detection pattern
  - internal/extractor/folder_name.go - Prompt substitution pattern

### Secondary (MEDIUM confidence)
- [Go Template Performance Comparison](https://blog.logrocket.com/golang-template-libraries-performance-comparison/) - Benchmarks showing strings.Replace 2x faster
- [Cobra Testing Best Practices](https://gianarb.it/blog/golang-mockmania-cli-command-with-cobra) - Testing patterns for CLI commands
- [Go Context Usage Guide](https://pkg.go.dev/context) - Official context documentation for cancellation
- [File Reading in Go Comparison](https://dev.to/moseeh_52/efficient-file-reading-in-go-mastering-bufionewscanner-vs-osreadfile-4h05) - os.ReadFile vs bufio benchmarks

### Tertiary (LOW confidence - marked for validation)
- [CV Tailoring Prompt Strategies](https://blog.theinterviewguys.com/claude-resume-prompts/) - General prompt engineering (not code-specific)
- [ATS Resume Optimization Guide](https://resume.io/resume-templates/ats) - ATS formatting requirements (not Go-specific)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All packages are stdlib or already in use in the codebase
- Architecture: HIGH - Patterns directly from existing code (cmd/apply.go, internal/executor/claude.go)
- Pitfalls: MEDIUM - Based on existing error handling patterns and common subprocess/file I/O issues
- Prompt engineering: MEDIUM - Existing prompts in assets/prompts/ verified; optimization strategies from web sources

**Research date:** 2026-02-03
**Valid until:** 2026-03-03 (30 days - stable patterns, stdlib APIs don't change frequently)
