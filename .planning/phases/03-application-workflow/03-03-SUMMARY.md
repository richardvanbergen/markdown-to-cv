---
phase: 03-application-workflow
plan: 03
subsystem: application
tags: [versioning, filepath, glob, filesystem]

# Dependency graph
requires:
  - phase: 03-01
    provides: Filesystem operations and folder name extraction patterns
provides:
  - ListVersions function for finding optimized CV versions
  - LatestVersionPath for getting most recent version
  - NextVersionPath for calculating next version number
  - Constants for optimized-cv filename pattern
affects: [04-optimize-command, optimize, versioned-output]

# Tech tracking
tech-stack:
  added: []
  patterns: [filepath.Glob for pattern matching, table-driven tests with t.TempDir]

key-files:
  created:
    - internal/application/versioning.go
    - internal/application/versioning_test.go
  modified: []

key-decisions:
  - "Use filepath.Glob over manual directory reading for cleaner pattern matching"
  - "Return empty slice (not error) for empty directories - enables graceful first-version handling"
  - "Ignore malformed filenames silently rather than error - defensive parsing"
  - "Version numbers must be positive integers (> 0)"

patterns-established:
  - "Table-driven tests with parallel execution using tt := tt capture"
  - "Use t.TempDir() for filesystem test isolation"
  - "Graceful handling of missing/empty directories as non-error conditions"

# Metrics
duration: 3min
completed: 2026-02-03
---

# Phase 03 Plan 03: Versioning Utilities Summary

**Optimized CV versioning with ListVersions/LatestVersionPath/NextVersionPath using filepath.Glob pattern matching**

## Performance

- **Duration:** 3 min
- **Started:** 2026-02-03T07:42:12Z
- **Completed:** 2026-02-03T07:45:00Z
- **Tasks:** 1
- **Files created:** 2

## Accomplishments
- Created internal/application package for application folder utilities
- Implemented ListVersions to find and sort all optimized-cv-N.md versions
- Implemented LatestVersionPath to get the highest version path
- Implemented NextVersionPath to calculate next version (max + 1, or 1 if empty)
- Comprehensive table-driven tests covering all edge cases (358 lines)

## Task Commits

Each task was committed atomically:

1. **Task 1: Versioning Utilities** - `pending` (feat)

**Plan metadata:** `pending` (docs: complete plan)

_Note: Commits pending - git commands require host execution via MCP_

## Files Created/Modified
- `internal/application/versioning.go` - Version number management for optimized CVs (91 lines)
  - ListVersions: finds optimized-cv-*.md files, parses versions, returns sorted
  - LatestVersionPath: returns path to highest version
  - NextVersionPath: returns path for next version
- `internal/application/versioning_test.go` - Comprehensive tests (358 lines)
  - TestListVersions: 9 test cases covering empty, single, multiple, malformed, non-existent
  - TestLatestVersionPath: 5 test cases
  - TestNextVersionPath: 6 test cases
  - TestVersioningIntegration: end-to-end workflow test
  - TestConstants: verifies constant values

## Decisions Made
- **filepath.Glob over manual dir reading**: Glob provides cleaner pattern matching with wildcard support
- **Empty = graceful, not error**: Empty directories return empty slice with nil error, enabling clean first-version handling
- **Silent malformed file handling**: Files like optimized-cv-abc.md are silently ignored rather than causing errors
- **Positive integers only**: Version numbers must be > 0 (rejects 0, -1)

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- **Git worktree issue**: The git repository is configured as a worktree pointing to a host filesystem path (`/Users/rich/Code/markdown-to-cv/.git/worktrees/...`) which is not accessible from within the container. Commits will need to be performed via host MCP tools.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Versioning utilities ready for Phase 4 optimize command
- Functions exported and tested: NextVersionPath, LatestVersionPath, ListVersions
- Pattern constants exported: OptimizedCVPrefix, OptimizedCVSuffix
- No blockers for Phase 4

---
*Phase: 03-application-workflow*
*Completed: 2026-02-03*
