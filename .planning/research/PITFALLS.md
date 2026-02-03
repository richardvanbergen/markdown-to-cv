# Domain Pitfalls: Go CLI + LLM Integration + NPM Ecosystem

**Domain:** Go CLI tool that shells out to LLMs and npm/npx
**Researched:** 2026-02-03
**Confidence:** HIGH (based on well-documented Go subprocess patterns, LLM integration challenges, and cross-ecosystem tooling)

## Critical Pitfalls

These mistakes cause rewrites, data loss, or production failures.

### Pitfall 1: Unbuffered stdout/stderr Reading from Subprocesses

**What goes wrong:** Using `cmd.Output()` or naive `cmd.Run()` with LLM subprocesses that produce large outputs causes deadlocks or truncated responses. The Go process blocks waiting for the subprocess to finish, but the subprocess blocks waiting for its stdout buffer to be consumed.

**Why it happens:** Claude responses can be kilobytes of text. If stdout buffer fills (typically 64KB on Unix), the subprocess blocks. Meanwhile, Go's `cmd.Wait()` waits for the process to exit. Deadlock.

**Consequences:**
- Hangs that require SIGKILL
- Truncated LLM responses without error
- Timeout logic triggers but doesn't fix root cause

**Prevention:**
```go
// BAD: Will deadlock on large output
cmd := exec.Command("claude", "-p", promptFile)
output, err := cmd.Output() // Blocks until process exits, but process blocks on full buffer

// GOOD: Stream output concurrently
cmd := exec.Command("claude", "-p", promptFile)
var stdout, stderr bytes.Buffer
cmd.Stdout = &stdout
cmd.Stderr = &stderr

if err := cmd.Start(); err != nil {
    return err
}

// Buffers are being filled concurrently as process runs
if err := cmd.Wait(); err != nil {
    return fmt.Errorf("claude failed: %w\nstderr: %s", err, stderr.String())
}

result := stdout.String()
```

**Detection:**
- CLI hangs during "Waiting for Claude..." step
- Works with short prompts, hangs with long responses
- `strace` shows subprocess blocked in `write()` syscall

**Phase impact:** Phase 1 (AI Integration) - Get this right from the start

---

### Pitfall 2: No JSON Schema Validation for LLM Output

**What goes wrong:** Parsing LLM output with `json.Unmarshal()` directly without validation. Claude returns JSON-like text with markdown code fences, extra commentary, or slightly malformed structure. Your parser explodes.

**Why it happens:** LLMs are non-deterministic. Even with "return JSON" prompts, Claude might:
- Wrap JSON in ```json...```
- Add explanatory text before/after
- Generate valid JSON with wrong field names
- Return partial JSON on truncation
- Include comments (not valid JSON)

**Consequences:**
- Random failures in production (works 95% of time)
- Silent data corruption (parses but wrong fields)
- User sees "invalid character '<' looking for beginning of value"
- No way to debug what Claude actually returned

**Prevention:**
```go
// BAD: Assumes Claude returns pure JSON
var resume JSONResume
if err := json.Unmarshal(claudeOutput, &resume); err != nil {
    return err // Cryptic error, no recovery
}

// GOOD: Extract and validate
func ParseClaudeJSON(output []byte, schema *jsonschema.Schema) (json.RawMessage, error) {
    // 1. Strip markdown fences
    content := stripMarkdownFences(output)

    // 2. Try to find JSON boundaries if mixed with text
    jsonStart := bytes.IndexByte(content, '{')
    jsonEnd := bytes.LastIndexByte(content, '}')
    if jsonStart == -1 || jsonEnd == -1 {
        return nil, fmt.Errorf("no JSON object found in output:\n%s", output)
    }

    extracted := content[jsonStart:jsonEnd+1]

    // 3. Validate it's parseable
    var raw json.RawMessage
    if err := json.Unmarshal(extracted, &raw); err != nil {
        return nil, fmt.Errorf("invalid JSON: %w\nExtracted:\n%s", err, extracted)
    }

    // 4. Validate against schema if provided
    if schema != nil {
        if err := schema.Validate(bytes.NewReader(raw)); err != nil {
            return nil, fmt.Errorf("schema validation failed: %w", err)
        }
    }

    return raw, nil
}
```

**Detection:**
- Intermittent "invalid character" errors
- Different errors on retry with same input
- Logs show Claude returned text + JSON
- JSON parses but validation fails downstream

**Phase impact:** Phase 1 (AI Integration) - Critical for CV generation, folder naming

---

### Pitfall 3: PATH Dependency Hell for npx/npm

**What goes wrong:** CLI works on dev machine, fails in CI/prod with "npx: command not found" or wrong npm version. Go's `exec.Command("npx")` doesn't use shell PATH resolution the same way your terminal does.

**Why it happens:**
- Go's `exec.LookPath()` uses different PATH than interactive shell
- `~/.bashrc`, `~/.zshrc` not loaded for non-interactive shells
- nvm/asdf/volta modify PATH in shell init scripts
- Different users (CI vs dev) have different PATH

**Consequences:**
- "works on my machine" syndrome
- CI builds fail mysteriously
- Users with different node setups can't run tool
- No clear error message (just "executable file not found")

**Prevention:**
```go
// BAD: Assumes npx is in PATH
cmd := exec.Command("npx", "resumed", "render", "...")

// GOOD: Explicit PATH management
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
        "/usr/local/bin/" + name,
        "/opt/homebrew/bin/" + name, // Apple Silicon
    }

    for _, candidate := range candidates {
        if _, err := os.Stat(candidate); err == nil {
            return candidate, nil
        }
    }

    return "", fmt.Errorf("%s not found in PATH or common locations. Please ensure Node.js is installed.", name)
}

// Usage
npxPath, err := findNodeExecutable("npx")
if err != nil {
    return fmt.Errorf("dependency check failed: %w", err)
}
cmd := exec.Command(npxPath, "resumed", "render", ...)
```

**Detection:**
- Works locally, fails in CI
- Fails for some users, not others
- `which npx` works in shell, but CLI says "not found"
- Different behavior when run via `sudo`

**Phase impact:** Phase 1 (Setup) - Block deployment without this

---

### Pitfall 4: Race Conditions in Temp File Cleanup

**What goes wrong:** Using `defer os.Remove(tempFile)` when multiple goroutines or subprocesses access the same temp file. File gets deleted while subprocess is reading it, causing "no such file or directory" errors.

**Why it happens:** Classic sequence:
1. Go creates temp prompt file
2. Go spawns `claude -p /tmp/prompt123.txt`
3. `defer` triggers, deletes file
4. Claude process tries to read file - GONE

Or in concurrent scenarios:
1. Goroutine A creates temp file
2. Goroutine B starts using it
3. Goroutine A's defer fires, deletes file
4. Goroutine B's subprocess fails

**Consequences:**
- Intermittent "file not found" errors
- More common under load (concurrent job processing)
- Hard to reproduce (timing-dependent)
- Temp directory fills up if cleanup never happens

**Prevention:**
```go
// BAD: Immediate defer
func runClaude(prompt string) error {
    tmpFile, err := os.CreateTemp("", "prompt-*.txt")
    if err != nil {
        return err
    }
    defer os.Remove(tmpFile.Name()) // WRONG: Fires before subprocess finishes

    tmpFile.WriteString(prompt)
    tmpFile.Close()

    cmd := exec.Command("claude", "-p", tmpFile.Name())
    return cmd.Run() // File deleted while Claude might still be reading
}

// GOOD: Wait for subprocess, then cleanup
func runClaude(prompt string) (string, error) {
    tmpFile, err := os.CreateTemp("", "prompt-*.txt")
    if err != nil {
        return "", err
    }
    tmpPath := tmpFile.Name()

    // Write and close file
    if _, err := tmpFile.WriteString(prompt); err != nil {
        os.Remove(tmpPath)
        return "", err
    }
    tmpFile.Close()

    // Run subprocess to completion
    cmd := exec.Command("claude", "-p", tmpPath)
    var stdout bytes.Buffer
    cmd.Stdout = &stdout

    err = cmd.Run()

    // NOW we can safely cleanup
    os.Remove(tmpPath)

    if err != nil {
        return "", err
    }

    return stdout.String(), nil
}
```

**Detection:**
- Errors mentioning temp file paths
- More failures under concurrent load
- Strace shows `open()` returning ENOENT
- Works in isolation, fails when parallelized

**Phase impact:** Phase 1 (AI Integration) - Will bite immediately with concurrent folder processing

---

### Pitfall 5: No Validation of npm Package Existence Before npx

**What goes wrong:** Running `npx resumed render ...` without checking if `resumed` package exists or is reachable. npx silently falls back to prompting for install (in interactive mode) or fails cryptically (in non-interactive mode).

**Why it happens:**
- npx tries to install missing packages on-the-fly
- In CI/non-TTY environments, this fails
- Network issues cause timeouts (30s+ hangs)
- Package name typos don't error immediately

**Consequences:**
- 30-second hangs waiting for npm registry
- Cryptic "ENOTTY" errors in CI
- Random failures when npm registry is down
- Users confused by "resumed@latest not found" messages

**Prevention:**
```go
// BAD: Just run npx and hope
cmd := exec.Command("npx", "resumed", "render", ...)

// GOOD: Pre-flight check with timeout
func validateNPMPackage(pkg string) error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Check if package exists in registry
    cmd := exec.CommandContext(ctx, "npm", "view", pkg, "version")
    if err := cmd.Run(); err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            return fmt.Errorf("npm registry unreachable (timeout checking %s)", pkg)
        }
        return fmt.Errorf("npm package '%s' not found. Is it published?", pkg)
    }

    return nil
}

// Run once at startup or cache result
if err := validateNPMPackage("resumed"); err != nil {
    return fmt.Errorf("dependency check failed: %w\nPlease ensure npm is configured and network is available", err)
}
```

**Detection:**
- Long hangs before error
- "TTY required" errors in CI
- Works with package cached, fails on fresh install
- Network-dependent failures

**Phase impact:** Phase 1 (Setup/Validation) - Should be in startup health checks

---

## Moderate Pitfalls

These cause delays, technical debt, or poor UX.

### Pitfall 6: Hardcoded Config File Paths

**What goes wrong:** Looking for `m2cv.yml` only in current directory. Fails when user runs command from subdirectory or wants global config.

**Why it happens:** Simple implementation: `config, err := os.ReadFile("m2cv.yml")`

**Consequences:**
- User must be in exact directory
- Can't have workspace-level settings
- No user-level defaults
- Frustrating UX ("but I have a config file!")

**Prevention:**
```go
func findConfig() (string, error) {
    // 1. Explicit flag (--config)
    if configFlag != "" {
        return configFlag, nil
    }

    // 2. Environment variable
    if envConfig := os.Getenv("M2CV_CONFIG"); envConfig != "" {
        return envConfig, nil
    }

    // 3. Walk up directory tree (like git)
    dir, _ := os.Getwd()
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

    // 4. User home directory
    home, _ := os.UserHomeDir()
    userConfig := filepath.Join(home, ".config", "m2cv", "config.yml")
    if _, err := os.Stat(userConfig); err == nil {
        return userConfig, nil
    }

    return "", fmt.Errorf("no config file found (searched: ./m2cv.yml, ~/.config/m2cv/config.yml)")
}
```

**Detection:**
- Users complain about "file not found" from subdirectories
- Common support question: "where do I put the config?"

**Phase impact:** Phase 1 (Config) - Design this upfront

---

### Pitfall 7: No Retry Logic for LLM Calls

**What goes wrong:** Single attempt at Claude invocation. Transient failures (rate limits, network blips, timeouts) cause entire operation to fail.

**Why it happens:** Treating LLM like a deterministic function instead of a network service.

**Consequences:**
- Brittle user experience
- Rate limit 429s kill entire batch job
- Network hiccups require manual retry
- No exponential backoff = thundering herd

**Prevention:**
```go
func callClaudeWithRetry(prompt string, maxRetries int) (string, error) {
    var lastErr error

    for attempt := 0; attempt < maxRetries; attempt++ {
        if attempt > 0 {
            backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
            log.Printf("Retry %d/%d after %v", attempt+1, maxRetries, backoff)
            time.Sleep(backoff)
        }

        result, err := callClaude(prompt)
        if err == nil {
            return result, nil
        }

        lastErr = err

        // Don't retry on non-transient errors
        if isNonRetryable(err) {
            return "", err
        }
    }

    return "", fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

func isNonRetryable(err error) bool {
    // Malformed prompt, auth failure, etc.
    return strings.Contains(err.Error(), "invalid prompt") ||
           strings.Contains(err.Error(), "authentication")
}
```

**Detection:**
- Users report intermittent failures
- Batch processing fails halfway through
- Errors during high-load periods

**Phase impact:** Phase 2 (Reliability) - Add after basic integration works

---

### Pitfall 8: Streaming vs Buffered Output Confusion

**What goes wrong:** Not showing progress for long-running LLM calls. User thinks CLI is hung. Or, trying to stream JSON output that must be complete before parsing.

**Why it happens:** Unclear requirements about when to stream vs buffer.

**Consequences:**
- Poor UX (no feedback during 10s+ Claude calls)
- Premature JSON parsing on partial output
- Can't cancel long-running operations

**Prevention:**
```go
// For operations that return JSON: buffer completely
func getStructuredOutput(prompt string) (*JSONResume, error) {
    spinner := startSpinner("Generating CV with Claude...")
    defer spinner.Stop()

    output, err := callClaude(prompt) // Fully buffered
    if err != nil {
        return nil, err
    }

    return parseJSON(output)
}

// For operations that show progress: stream with updates
func generateWithProgress(prompt string) (string, error) {
    cmd := exec.Command("claude", "-p", prompt)

    // Stream stderr for progress updates
    stderr, _ := cmd.StderrPipe()
    go func() {
        scanner := bufio.NewScanner(stderr)
        for scanner.Scan() {
            log.Printf("Claude: %s", scanner.Text())
        }
    }()

    var stdout bytes.Buffer
    cmd.Stdout = &stdout

    if err := cmd.Run(); err != nil {
        return "", err
    }

    return stdout.String(), nil
}
```

**Detection:**
- Users report CLI "hanging"
- No way to see progress on long operations
- Can't interrupt long-running commands

**Phase impact:** Phase 2 (UX Polish) - After core functionality works

---

### Pitfall 9: Missing Input Validation Before Expensive LLM Calls

**What goes wrong:** Sending malformed markdown or empty content to Claude, wasting API credits and time on guaranteed failures.

**Why it happens:** Optimistic coding - assume input is valid.

**Consequences:**
- Wasted API costs (Claude returns error after processing)
- Slow failure loop (user → Claude → error → user)
- No actionable error messages

**Prevention:**
```go
func validateMarkdownForCV(mdPath string) error {
    content, err := os.ReadFile(mdPath)
    if err != nil {
        return err
    }

    // Check basic structure
    if len(content) == 0 {
        return fmt.Errorf("empty markdown file")
    }

    // Check for required sections (customize based on requirements)
    requiredSections := []string{"# Experience", "# Education"}
    for _, section := range requiredSections {
        if !bytes.Contains(content, []byte(section)) {
            return fmt.Errorf("missing required section: %s", section)
        }
    }

    // Check file size (Claude has token limits)
    if len(content) > 100000 { // ~25k tokens
        return fmt.Errorf("markdown file too large (>100KB). Please split or summarize.")
    }

    return nil
}

// Use before expensive operations
if err := validateMarkdownForCV(inputPath); err != nil {
    return fmt.Errorf("invalid input: %w", err)
}

output, err := callClaude(prompt) // Now confident this will work
```

**Detection:**
- High error rate on Claude calls
- Users confused by LLM errors
- API costs higher than expected

**Phase impact:** Phase 1 (Validation) - Prevents wasted API calls

---

### Pitfall 10: Context Leakage in Subprocess Signals

**What goes wrong:** Using `context.WithTimeout()` but not propagating signal to subprocess. Claude process keeps running after Go context cancels.

**Why it happens:** `exec.CommandContext()` sends SIGKILL on context cancel, but child processes might not handle it gracefully.

**Consequences:**
- Zombie Claude processes
- Resource leaks (file descriptors, memory)
- Background processes consuming API quota
- Can't truly cancel operations

**Prevention:**
```go
// BAD: Context timeout doesn't guarantee cleanup
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
cmd := exec.CommandContext(ctx, "claude", "-p", promptFile)
cmd.Run() // If timeout, SIGKILL sent but subprocess might not exit cleanly

// GOOD: Proper signal handling and cleanup verification
func runWithTimeout(ctx context.Context, name string, args ...string) error {
    cmd := exec.CommandContext(ctx, name, args...)

    // Set process group so we can kill entire tree
    cmd.SysProcAttr = &syscall.SysProcAttr{
        Setpgid: true,
    }

    if err := cmd.Start(); err != nil {
        return err
    }

    // Wait in goroutine
    done := make(chan error, 1)
    go func() {
        done <- cmd.Wait()
    }()

    select {
    case <-ctx.Done():
        // Kill process group (catches child processes too)
        pgid, _ := syscall.Getpgid(cmd.Process.Pid)
        syscall.Kill(-pgid, syscall.SIGTERM)

        // Wait briefly for graceful shutdown
        time.Sleep(100 * time.Millisecond)

        // Force kill if still running
        syscall.Kill(-pgid, syscall.SIGKILL)

        return ctx.Err()
    case err := <-done:
        return err
    }
}
```

**Detection:**
- `ps aux | grep claude` shows zombie processes
- File descriptors leak over time
- Memory usage grows with each timeout
- API quota consumed unexpectedly

**Phase impact:** Phase 2 (Reliability) - Important for production stability

---

## Minor Pitfalls

These cause annoyance but are easily fixed.

### Pitfall 11: Inconsistent Error Messages Across Subprocess Failures

**What goes wrong:** Different error formats for npm failures vs Claude failures vs validation failures. Hard to parse or guide users to solutions.

**Prevention:**
```go
type CommandError struct {
    Command string
    Args    []string
    Stdout  string
    Stderr  string
    Err     error
}

func (e *CommandError) Error() string {
    return fmt.Sprintf(
        "Command failed: %s %s\nError: %v\nStdout: %s\nStderr: %s",
        e.Command, strings.Join(e.Args, " "), e.Err, e.Stdout, e.Stderr,
    )
}

// Wrap all subprocess errors consistently
func runCommand(name string, args ...string) error {
    cmd := exec.Command(name, args...)
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    if err := cmd.Run(); err != nil {
        return &CommandError{
            Command: name,
            Args:    args,
            Stdout:  stdout.String(),
            Stderr:  stderr.String(),
            Err:     err,
        }
    }

    return nil
}
```

**Phase impact:** Phase 1 (Foundation) - Easier to standardize early

---

### Pitfall 12: No Debug Mode for Subprocess Inspection

**What goes wrong:** When things fail, user can't see what was actually sent to Claude or npm.

**Prevention:**
```go
var debugMode bool // From --debug flag

func runCommand(name string, args ...string) error {
    if debugMode {
        log.Printf("[DEBUG] Running: %s %s", name, strings.Join(args, " "))
        log.Printf("[DEBUG] Working directory: %s", workingDir)
        log.Printf("[DEBUG] Environment: %v", os.Environ())
    }

    cmd := exec.Command(name, args...)

    if debugMode {
        // Echo all output to console
        cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
        cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)
    }

    return cmd.Run()
}
```

**Phase impact:** Phase 1 (Tooling) - Essential for development and support

---

### Pitfall 13: Ignoring npm Peer Dependency Warnings

**What goes wrong:** `npx resumed` works but prints peer dependency warnings. Users think something is broken.

**Prevention:** Document expected warnings or suppress with `npm_config_loglevel=error` environment variable.

**Phase impact:** Phase 2 (UX Polish)

---

### Pitfall 14: No Version Pinning for npx Packages

**What goes wrong:** `npx resumed` uses latest version, which might break between runs.

**Prevention:**
```go
// Pin to specific version
cmd := exec.Command("npx", "resumed@2.1.0", "render", ...)
```

**Phase impact:** Phase 1 (Stability) - Should be configurable in m2cv.yml

---

## Phase-Specific Warnings

| Phase Topic | Likely Pitfall | Mitigation |
|-------------|---------------|------------|
| Initial Claude Integration | Unbuffered stdout/stderr (Pitfall #1) | Use streaming from day 1, not Output() |
| JSON Parsing | No extraction layer (Pitfall #2) | Write `ParseClaudeJSON()` helper first |
| NPM Integration | PATH issues (Pitfall #3) | Explicit PATH search on startup |
| Temp File Usage | Race conditions (Pitfall #4) | Cleanup after subprocess, not defer immediately |
| Batch Processing | No concurrency limits | Semaphore to limit parallel Claude calls |
| Config System | Hardcoded paths (Pitfall #6) | Walk-up directory tree pattern |
| Error Handling | Inconsistent formats (Pitfall #11) | Unified CommandError wrapper |
| Production Deployment | Zombie processes (Pitfall #10) | Process group signal handling |

---

## Sources

**HIGH Confidence:**
- Go `os/exec` package documentation (official)
- Subprocess buffer deadlock patterns (well-documented Go gotcha)
- JSON parsing best practices (Go standard library patterns)
- PATH resolution differences between interactive/non-interactive shells (Unix/Linux behavior)
- Temp file cleanup race conditions (classic OS concurrency issue)

**MEDIUM Confidence:**
- LLM output parsing challenges (based on Anthropic Claude API patterns as of training cutoff)
- npm/npx behavior in non-interactive mode (npm documentation)
- Node version manager PATH modifications (nvm, volta, asdf documentation)

**LOW Confidence (needs validation):**
- Specific resumed package behavior (would need to check current npm registry)
- Latest Claude CLI flags and options (should verify with `claude --help`)
- Exact token limits for Claude 2026 models (check current Anthropic documentation)

---

## Notes for Roadmap Creation

**Must address in Phase 1:**
- Subprocess output buffering (Pitfall #1)
- JSON extraction layer (Pitfall #2)
- PATH resolution (Pitfall #3)
- Temp file lifecycle (Pitfall #4)

**Can defer to Phase 2:**
- Retry logic (Pitfall #7)
- Progress indicators (Pitfall #8)
- Context cancellation cleanup (Pitfall #10)

**Quick wins for UX:**
- Config file discovery (Pitfall #6)
- Debug mode (Pitfall #12)
- Input validation (Pitfall #9)

**Research needed during implementation:**
- Exact resumed package version and API (Phase 1)
- Claude CLI current flags and options (Phase 1)
- JSON Resume schema validation approach (Phase 1)
