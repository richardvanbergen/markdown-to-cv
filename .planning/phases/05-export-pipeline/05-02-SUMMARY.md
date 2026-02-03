---
phase: 05-export-pipeline
plan: 02
subsystem: cli
tags: [cobra, pdf-generation, json-resume, pipeline-orchestration]

# Dependency graph
requires:
  - phase: 05-01
    provides: generator service layer (ExtractJSON, Validator, Exporter)
  - phase: 04-02
    provides: optimize command pattern, versioned optimized CV output
  - phase: 03-03
    provides: apply command, application folder structure
provides:
  - Generate subcommand orchestrating full PDF pipeline
  - User-facing endpoint for m2cv workflow completion
  - JSON Resume output (resume.json) for debugging
  - PDF output (resume.pdf) as final deliverable
affects: [future-cli-enhancements, documentation]

# Tech tracking
tech-stack:
  added: []
  patterns: [command-specific PreRunE for preflight checks, pipeline orchestration in RunE]

key-files:
  created:
    - cmd/generate.go
    - cmd/generate_test.go
  modified:
    - cmd/root.go

key-decisions:
  - "Generate has command-specific PreRunE for CheckResumed (not in root PersistentPreRunE)"
  - "Fallback to 'even' theme if neither flag nor config specifies theme"
  - "Write resume.json intermediate output for debugging/troubleshooting"
  - "Tests must disable both root PersistentPreRunE and generate PreRunE for isolation"

patterns-established:
  - "Command-specific preflight: use PreRunE on individual command for dependency-specific checks"
  - "Test pattern: generateCmd.PreRunE = nil alongside rootCmd.PersistentPreRunE = nil"

# Metrics
duration: 7min
completed: 2026-02-03
---

# Phase 5 Plan 2: Generate Command Summary

**User-facing generate command orchestrating full pipeline: optimized CV to JSON Resume via Claude to PDF via resumed**

## Performance

- **Duration:** 7 min
- **Started:** 2026-02-03T09:02:23Z
- **Completed:** 2026-02-03T09:09:12Z
- **Tasks:** 2
- **Files modified:** 3 (cmd/generate.go, cmd/generate_test.go, cmd/root.go)

## Accomplishments
- Generate command completes the m2cv workflow (apply -> optimize -> generate)
- Full pipeline: read optimized CV, convert to JSON Resume via Claude, validate schema, export PDF
- Comprehensive tests covering error paths, flag bindings, and error ordering
- Actionable error messages guide users to correct commands (e.g., "Run 'm2cv optimize' first")

## Task Commits

Each task was committed atomically:

1. **Task 1: Generate Command Implementation** - `b3ab155` (feat)
2. **Task 2: Wire Generate to Root and Test** - `fad788d` (test)

## Files Created/Modified
- `cmd/generate.go` - newGenerateCommand() with --theme and -m flags, runGenerate() orchestrating full pipeline
- `cmd/generate_test.go` - 10 test functions covering structure, error paths, flag bindings
- `cmd/root.go` - Added newGenerateCommand() to Execute()

## Decisions Made
- Generate command has its own PreRunE for CheckResumed rather than adding to root's PersistentPreRunE, keeping the check command-specific (only generate needs resumed, not other commands like apply/optimize)
- Fallback default theme is "even" if neither --theme flag nor config.DefaultTheme is set
- resume.json is written to application folder as intermediate output for debugging
- In tests, must disable both root's PersistentPreRunE and generate's PreRunE to isolate command logic from preflight checks

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Fixed test failures from generate PreRunE**
- **Found during:** Task 2 (tests)
- **Issue:** Tests failing because generate command's PreRunE was checking for resumed even when testing error paths
- **Fix:** Added `generateCmd.PreRunE = nil` to all tests that needed to exercise RunE error handling
- **Files modified:** cmd/generate_test.go
- **Verification:** All tests pass
- **Committed in:** fad788d (Task 2 commit)

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** Test fix was necessary to exercise command logic in isolation. No scope creep.

## Issues Encountered
- MCP tools not directly available in function list, required HTTP JSON-RPC calls to cbox-host server - resolved by using curl to invoke run_command tool

## User Setup Required

None - no external service configuration required. (Users must have Claude CLI, npm, and resumed installed per existing setup requirements)

## Next Phase Readiness
- All m2cv commands implemented: init, apply, optimize, generate
- Full workflow now executable: job description -> tailored CV -> JSON Resume -> PDF
- Ready for documentation updates and any polish/refinement

---
*Phase: 05-export-pipeline*
*Completed: 2026-02-03*
