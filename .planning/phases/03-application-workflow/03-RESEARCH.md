# Phase 3: Application Workflow - Research

**Researched:** 2026-02-03
**Domain:** Go filesystem operations, subprocess integration, path handling
**Confidence:** HIGH

## Summary

This phase implements the `apply` command which creates application folders with AI-extracted names. The standard approach involves using Go's `os` and `path/filepath` packages for filesystem operations, executing Claude CLI via subprocess to extract structured data (company + role), and implementing proper path sanitization and error handling.

**Key technical challenges:**
1. Subprocess execution with JSON output parsing (Claude CLI integration)
2. Filename sanitization from AI-extracted names
3. File copying and directory creation with proper error handling
4. Testing filesystem operations without touching disk

**Primary recommendation:** Use standard library `os.MkdirAll` for directory creation, `io.Copy` for file operations, `encoding/json` for parsing Claude output, and `testing/fstest` for filesystem mocking in tests. Implement the repository pattern (matching existing codebase) for filesystem operations to support test isolation.

## Standard Stack

The established libraries/tools for this domain:

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| os | stdlib | Directory creation, file operations | Native filesystem operations, cross-platform |
| path/filepath | stdlib | Path manipulation, joining, cleaning | OS-aware path handling (Windows/Unix) |
| encoding/json | stdlib | JSON parsing from Claude output | Built-in structured data parsing |
| io | stdlib | File copying operations | Efficient streaming file copies |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| testing/fstest | stdlib (Go 1.16+) | In-memory filesystem for tests | Unit testing filesystem operations |
| github.com/spf13/afero | latest | Filesystem abstraction | If complex filesystem mocking needed |
| github.com/kennygrant/sanitize | latest | Filename sanitization | If complex sanitization rules required |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| os.MkdirAll | os.Mkdir | MkdirAll creates parent dirs, more robust |
| io.Copy | os.ReadFile + os.WriteFile | io.Copy streams, better for large files |
| testing/fstest.MapFS | afero.MemMapFs | MapFS is stdlib, afero has more features |
| Custom sanitization | kennygrant/sanitize | Custom gives control, library handles edge cases |

**Installation:**
No external dependencies needed for core functionality. Standard library sufficient.

## Architecture Patterns

### Recommended Project Structure
```
cmd/
├── apply.go              # apply subcommand
internal/
├── filesystem/
│   ├── operations.go     # Repository interface
│   ├── operations_test.go
├── extractor/
│   ├── folder_name.go    # Claude integration for name extraction
│   ├── folder_name_test.go
├── sanitize/
│   ├── filename.go       # Path/filename sanitization
│   ├── filename_test.go
```

### Pattern 1: Repository Pattern for Filesystem Operations
**What:** Abstract filesystem operations behind an interface for testability
**When to use:** All file I/O operations
**Example:**
```go
// Source: Existing codebase pattern (config/config.go)
type FilesystemOperations interface {
    CreateDir(path string, perm os.FileMode) error
    CopyFile(src, dst string) error
    ReadFile(path string) ([]byte, error)
}

type osFilesystem struct{}

func (f *osFilesystem) CreateDir(path string, perm os.FileMode) error {
    return os.MkdirAll(path, perm)
}

func (f *osFilesystem) CopyFile(src, dst string) error {
    source, err := os.Open(src)
    if err != nil {
        return fmt.Errorf("open source: %w", err)
    }
    defer source.Close()

    dest, err := os.Create(dst)
    if err != nil {
        return fmt.Errorf("create dest: %w", err)
    }
    defer dest.Close()

    _, err = io.Copy(dest, source)
    return err
}
```

### Pattern 2: Subprocess with JSON Output Parsing
**What:** Execute Claude CLI with structured output format
**When to use:** Extracting structured data from LLM (company + role names)
**Example:**
```go
// Source: https://pkg.go.dev/encoding/json
type FolderNameResult struct {
    Company string `json:"company"`
    Role    string `json:"role"`
}

func ExtractFolderName(ctx context.Context, executor executor.ClaudeExecutor, jobDesc string) (string, error) {
    prompt := fmt.Sprintf(
        "Extract company name and role from this job description. Return JSON: {\"company\": \"...\", \"role\": \"...\"}\n\n%s",
        jobDesc,
    )

    output, err := executor.Execute(ctx, prompt,
        executor.WithOutputFormat("json"),
    )
    if err != nil {
        return "", fmt.Errorf("claude execution: %w", err)
    }

    var result FolderNameResult
    if err := json.Unmarshal([]byte(output), &result); err != nil {
        return "", fmt.Errorf("parse json: %w", err)
    }

    // Sanitize and format
    return sanitize.PathName(result.Company + "-" + result.Role), nil
}
```

### Pattern 3: Atomic File Operations (Avoid TOCTOU)
**What:** Combine check and action into single operation
**When to use:** File existence checks before operations
**Example:**
```go
// Source: https://blog.stackademic.com/how-to-check-for-file-existence-in-go-are-there-potential-race-conditions-78af23ecc456
// DON'T: Check then act (race condition)
if _, err := os.Stat(path); err == nil {
    return fmt.Errorf("folder already exists")
}
os.MkdirAll(path, 0755)

// DO: Let the operation fail naturally
err := os.MkdirAll(path, 0755)
if err != nil {
    if os.IsExist(err) {
        return fmt.Errorf("folder already exists: %w", err)
    }
    return err
}
```

### Pattern 4: Path Manipulation with filepath
**What:** Use filepath.Join and filepath.Clean for safe path handling
**When to use:** Building paths, handling user input
**Example:**
```go
// Source: https://pkg.go.dev/path/filepath
// Always use filepath.Join, never string concatenation
appFolder := filepath.Join("applications", folderName)
jobDescDest := filepath.Join(appFolder, filepath.Base(jobDescPath))

// Clean paths before use
cleanPath := filepath.Clean(userInput)
```

### Anti-Patterns to Avoid
- **Check-then-act filesystem operations:** Creates race conditions (TOCTOU)
- **Hardcoded path separators:** `"path/to/file"` breaks on Windows, use `filepath.Join`
- **Using cmd.Output() for subprocess:** Can deadlock with large output, use bytes.Buffer
- **Ignoring stderr on subprocess failure:** Loses diagnostic information
- **Custom path sanitization without Unicode handling:** Breaks on non-ASCII characters

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| File copying | Manual read/write loops | `io.Copy(dst, src)` | Handles buffering, errors, streaming efficiently |
| Path joining | String concatenation | `filepath.Join(...)` | Cross-platform separator handling |
| Filename sanitization | Regex replacements | `path.Base()` + whitelist chars | Unicode normalization, OS-specific rules |
| Directory walk-up | Loop with filepath.Dir | Existing `config.Find()` pattern | Already tested, handles edge cases |
| Subprocess JSON output | String parsing | `json.Unmarshal()` | Handles escaping, types, validation |
| Temp directories | Random name generation | `t.TempDir()` in tests | Auto-cleanup, unique names, race-free |

**Key insight:** Go's standard library is comprehensive for filesystem operations. The main complexity is in proper error handling and testing, not in the operations themselves.

## Common Pitfalls

### Pitfall 1: TOCTOU Race Conditions
**What goes wrong:** Checking if a file exists before operating on it creates a time-of-check-time-of-use race condition
**Why it happens:** Developer instinct is to check first, then act
**How to avoid:** Let operations fail naturally and handle errors
**Warning signs:** Code pattern `if os.Stat(...); err == nil { os.MkdirAll(...) }`
**Reference:** https://blog.stackademic.com/how-to-check-for-file-existence-in-go-are-there-potential-race-conditions-78af23ecc456

### Pitfall 2: Subprocess Deadlock with cmd.Output()
**What goes wrong:** Using `cmd.Output()` or `cmd.CombinedOutput()` can deadlock when stderr/stdout buffers fill
**Why it happens:** These methods don't start reading until process completes
**How to avoid:** Use `bytes.Buffer` with `cmd.Stdout/Stderr`, then `cmd.Start()` + `cmd.Wait()`
**Warning signs:** Hangs when processing large output
**Reference:** https://gopheradvent.com/calendar/2021/gotchas-in-exec-errors/

### Pitfall 3: Path Separator Hardcoding
**What goes wrong:** Using `/` or `\` directly in paths breaks cross-platform compatibility
**Why it happens:** Developer works on one OS
**How to avoid:** Always use `filepath.Join()`, never string concatenation
**Warning signs:** Paths like `"applications/" + name` or `"applications\\" + name`
**Reference:** https://pkg.go.dev/path/filepath

### Pitfall 4: Ignoring Subprocess stderr
**What goes wrong:** Subprocess fails with cryptic error, actual error message was in stderr
**Why it happens:** Only capturing stdout by default
**How to avoid:** Capture both stdout and stderr, include stderr in error messages
**Warning signs:** Error messages like "exit status 1" without explanation
**Reference:** Existing codebase pattern in `internal/executor/claude.go:114-117`

### Pitfall 5: Unicode in Filenames
**What goes wrong:** Filename sanitization breaks on company names with non-ASCII characters (e.g., "Müller & Co")
**Why it happens:** Simple character whitelisting removes valid Unicode
**How to avoid:** Use libraries that handle Unicode normalization, or keep valid Unicode ranges
**Warning signs:** Company names get mangled or empty strings
**Reference:** https://github.com/subosito/gozaru

### Pitfall 6: Parallel Test Race Conditions
**What goes wrong:** Table-driven tests with `t.Parallel()` share loop variable
**Why it happens:** Loop variable `tt` is shared across all iterations
**How to avoid:** Capture loop variable: `tt := tt` before `t.Run()`
**Warning signs:** Random test failures when run with `-race` flag
**Reference:** https://www.glukhov.org/post/2025/12/parallel-table-driven-tests-in-go/

## Code Examples

Verified patterns from official sources:

### Directory Creation
```go
// Source: https://pkg.go.dev/os
// MkdirAll creates parent directories, idempotent if exists
appDir := filepath.Join("applications", folderName)
if err := os.MkdirAll(appDir, 0755); err != nil {
    return fmt.Errorf("create application directory: %w", err)
}
```

### File Copying (Streaming)
```go
// Source: https://opensource.com/article/18/6/copying-files-go
func CopyFile(src, dst string) error {
    source, err := os.Open(src)
    if err != nil {
        return fmt.Errorf("open source: %w", err)
    }
    defer source.Close()

    dest, err := os.Create(dst)
    if err != nil {
        return fmt.Errorf("create dest: %w", err)
    }
    defer dest.Close()

    if _, err := io.Copy(dest, source); err != nil {
        return fmt.Errorf("copy data: %w", err)
    }

    // Sync to ensure data is written
    return dest.Sync()
}
```

### JSON Parsing from Subprocess
```go
// Source: https://pkg.go.dev/encoding/json
type ExtractedData struct {
    Company string `json:"company"`
    Role    string `json:"role"`
}

var result ExtractedData
if err := json.Unmarshal([]byte(output), &result); err != nil {
    return fmt.Errorf("parse JSON output: %w", err)
}
```

### Filename Sanitization (Simple)
```go
// Source: https://pkg.go.dev/github.com/kennygrant/sanitize
// For simple cases, can use custom function:
func SanitizeFilename(s string) string {
    // Replace invalid chars with hyphens
    s = strings.Map(func(r rune) rune {
        if r == ' ' || r == '/' || r == '\\' {
            return '-'
        }
        if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '-' || r == '_' {
            return r
        }
        return -1 // Remove
    }, s)

    // Lowercase for consistency
    return strings.ToLower(s)
}
```

### Table-Driven Test Pattern
```go
// Source: https://go.dev/wiki/TableDrivenTests
func TestExtractFolderName(t *testing.T) {
    tests := []struct {
        name     string
        jobDesc  string
        want     string
        wantErr  bool
    }{
        {
            name: "valid job description",
            jobDesc: "Software Engineer at Google\n\nWe are looking for...",
            want: "google-software-engineer",
            wantErr: false,
        },
        {
            name: "missing company",
            jobDesc: "Looking for a developer\n\nNo company mentioned",
            want: "",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        tt := tt // Capture loop variable for parallel tests
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()

            got, err := ExtractFolderName(context.Background(), mockExecutor, tt.jobDesc)
            if (err != nil) != tt.wantErr {
                t.Errorf("ExtractFolderName() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("ExtractFolderName() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Filesystem Mock for Testing
```go
// Source: https://pkg.go.dev/testing/fstest
func TestApplyCommand(t *testing.T) {
    // Use interface for filesystem operations
    type fsOps interface {
        MkdirAll(path string, perm os.FileMode) error
        WriteFile(name string, data []byte, perm os.FileMode) error
    }

    // Test implementation
    type testFS struct {
        dirs  []string
        files map[string][]byte
    }

    func (f *testFS) MkdirAll(path string, perm os.FileMode) error {
        f.dirs = append(f.dirs, path)
        return nil
    }

    // Test uses mock, production uses real filesystem
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| os.Mkdir | os.MkdirAll | Always | MkdirAll creates parents, more robust |
| filepath.Walk | filepath.WalkDir | Go 1.16 | WalkDir is more efficient (uses DirEntry) |
| os.IsNotExist | errors.Is(err, fs.ErrNotExist) | Go 1.13 | New error handling with wrapping |
| ioutil.ReadFile | os.ReadFile | Go 1.16 | ioutil deprecated, use os directly |
| ioutil.WriteFile | os.WriteFile | Go 1.16 | ioutil deprecated, use os directly |

**Deprecated/outdated:**
- `ioutil` package: All functions moved to `os` and `io` packages in Go 1.16
- `os.IsNotExist()`: Still works but `errors.Is(err, fs.ErrNotExist)` is preferred for wrapped errors

## Open Questions

Things that couldn't be fully resolved:

1. **Filename sanitization complexity level**
   - What we know: Standard library has no built-in sanitization; kennygrant/sanitize exists
   - What's unclear: How aggressive sanitization needs to be for cross-platform compatibility
   - Recommendation: Start with simple custom function (lowercase, replace spaces/slashes), add library if edge cases appear

2. **Claude CLI JSON output reliability**
   - What we know: Claude supports `--output-format json` for structured output
   - What's unclear: Error handling when Claude returns malformed JSON or refuses to extract
   - Recommendation: Implement retry logic, fallback to sanitized user input if extraction fails

3. **Application folder collision handling**
   - What we know: os.MkdirAll succeeds if directory exists
   - What's unclear: Should command fail on existing folder, append number, or overwrite?
   - Recommendation: Start with error on collision, defer numbering scheme to WORK-04 versioning pattern

## Sources

### Primary (HIGH confidence)
- [os package - Go Packages](https://pkg.go.dev/os) - Directory operations, file I/O, error types
- [path/filepath package - Go Packages](https://pkg.go.dev/path/filepath) - Path manipulation functions
- [encoding/json package - Go Packages](https://pkg.go.dev/encoding/json) - JSON parsing patterns
- Existing codebase - internal/executor/claude.go (subprocess pattern), internal/config/config.go (repository pattern)

### Secondary (MEDIUM confidence)
- [3 ways to copy files in Go | Opensource.com](https://opensource.com/article/18/6/copying-files-go) - File copying best practices
- [Go Wiki: TableDrivenTests](https://go.dev/wiki/TableDrivenTests) - Table-driven test patterns
- [Gotchas when running failing commands in the Go os/exec package](https://gopheradvent.com/calendar/2021/gotchas-in-exec-errors/) - Subprocess error handling
- [How to Check for File Existence in Go: Are There Potential Race Conditions?](https://blog.stackademic.com/how-to-check-for-file-existence-in-go-are-there-potential-race-conditions-78af23ecc456) - TOCTOU patterns
- [Parallel Table-Driven Tests in Go](https://www.glukhov.org/post/2025/12/parallel-table-driven-tests-in-go/) - Test parallelization

### Tertiary (LOW confidence)
- [kennygrant/sanitize](https://github.com/kennygrant/sanitize) - Filename sanitization library (if needed)
- [Claude API Structured Output guides](https://platform.claude.com/docs/en/build-with-claude/structured-outputs) - Claude JSON mode documentation
- [testing/fstest package](https://pkg.go.dev/testing/fstest) - Filesystem testing utilities

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Standard library well-documented, existing codebase patterns established
- Architecture: HIGH - Repository pattern proven in existing code, subprocess pattern verified
- Pitfalls: HIGH - All pitfalls sourced from official docs, existing codebase, or authoritative articles

**Research date:** 2026-02-03
**Valid until:** 2026-03-03 (30 days - stable domain, standard library changes infrequently)
