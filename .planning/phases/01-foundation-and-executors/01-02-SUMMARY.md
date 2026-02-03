---
phase: 01-foundation-and-executors
plan: 02
subsystem: executor
tags: [go, subprocess, exec, npm, claude-cli, streaming, stdin]

# Dependency graph
requires: []
provides:
  - ClaudeExecutor interface with streaming subprocess execution
  - NPMExecutor interface with version manager path resolution
  - FindNodeExecutable function for npm/npx/node discovery
affects: [02-project-workflow, 03-cv-processing]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "bytes.Buffer for subprocess output capture (avoid cmd.Output deadlocks)"
    - "cmd.Start() + cmd.Wait() pattern (not cmd.Run() with Output)"
    - "stdin piping for large prompts (avoid shell argument limits)"
    - "Functional options pattern for executor configuration"

key-files:
  created:
    - internal/executor/claude.go
    - internal/executor/claude_test.go
    - internal/executor/npm.go
    - internal/executor/npm_test.go
    - internal/executor/find.go
    - internal/executor/find_test.go
  modified: []

key-decisions:
  - "Functional options for both executor construction (WithClaudePath, WithNPMPath) and execution (WithModel, WithOutputFormat)"
  - "CheckInstalled uses filesystem check (os.Stat) not npm command - faster and simpler"
  - "FindNodeExecutable checks nvm/volta/asdf/fnm before system paths for consistency with user's node version"

patterns-established:
  - "Subprocess Pattern: Always use bytes.Buffer + cmd.Start() + cmd.Wait(), never cmd.Output()"
  - "PATH Resolution Pattern: exec.LookPath first, then version manager fallbacks"
  - "Error Pattern: Include stderr content in error messages for debugging"

# Metrics
duration: 7min
completed: 2026-02-03
---

# Phase 01 Plan 02: Subprocess Executors Summary

**ClaudeExecutor and NPMExecutor with bytes.Buffer streaming, stdin piping, and nvm/volta/asdf/fnm path resolution**

## Performance

- **Duration:** 7 min
- **Started:** 2026-02-03T05:24:06Z
- **Completed:** 2026-02-03T05:31:12Z
- **Tasks:** 3 (TDD cycles)
- **Files created:** 6

## Accomplishments

- FindNodeExecutable resolves npm/npx/node across nvm, volta, asdf, fnm, and standard PATH
- ClaudeExecutor streams output via bytes.Buffer, passes prompts via stdin, respects context cancellation
- NPMExecutor provides Install, Init, and CheckInstalled with proper directory handling
- 23 tests covering all edge cases (fallback paths, errors, cancellation, large output)

## Task Commits

Each TDD task was completed:

1. **Task 1: FindNodeExecutable** - TDD cycle
   - RED: find_test.go with 6 tests for PATH and fallback scenarios
   - GREEN: find.go implementation with exec.LookPath + fallbacks

2. **Task 2: ClaudeExecutor** - TDD cycle
   - RED: claude_test.go with 10 tests covering stdin, streaming, options, errors
   - GREEN: claude.go implementation with bytes.Buffer pattern

3. **Task 3: NPMExecutor** - TDD cycle
   - RED: npm_test.go with 7 tests for Install, Init, CheckInstalled
   - GREEN: npm.go implementation using FindNodeExecutable

## Files Created

- `internal/executor/find.go` (76 lines) - FindNodeExecutable with version manager fallbacks
- `internal/executor/find_test.go` (195 lines) - Tests for PATH resolution
- `internal/executor/claude.go` (122 lines) - ClaudeExecutor with streaming output
- `internal/executor/claude_test.go` (327 lines) - Tests for Claude subprocess execution
- `internal/executor/npm.go` (116 lines) - NPMExecutor for package management
- `internal/executor/npm_test.go` (309 lines) - Tests for npm operations

## Decisions Made

1. **Functional options pattern** - Enables clean API: `NewClaudeExecutor(WithClaudePath("/custom/path"))`
2. **CheckInstalled uses filesystem** - Checking `node_modules/<pkg>` is faster than running npm
3. **Version manager order** - nvm, volta, asdf, fnm, then system paths (user's node version takes priority)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Go installation required**
- **Found during:** Initial setup
- **Issue:** Go not available in container environment
- **Fix:** Downloaded and extracted Go 1.22.5 to /tmp
- **Files modified:** None (runtime environment)
- **Verification:** `go version` returns successfully

**2. [Rule 1 - Bug] Context cancellation test used unavailable `sleep`**
- **Found during:** Task 2 GREEN phase
- **Issue:** Test relied on `sleep` which behaved differently in minimal shell
- **Fix:** Changed to portable infinite `while true; do :; done` loop
- **Files modified:** internal/executor/claude_test.go
- **Verification:** Test passes consistently within timeout

---

**Total deviations:** 2 auto-fixed (1 blocking, 1 bug)
**Impact on plan:** Minimal - both fixes were necessary for execution in container environment

## Issues Encountered

- Git operations cannot be performed directly in container (worktree on host machine)
- Commits need to be made via MCP host command or manually

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- ClaudeExecutor ready for use in optimize and generate commands
- NPMExecutor ready for init command (resumed/theme installation)
- FindNodeExecutable can be reused for any Node.js tooling

---
*Phase: 01-foundation-and-executors*
*Completed: 2026-02-03*
