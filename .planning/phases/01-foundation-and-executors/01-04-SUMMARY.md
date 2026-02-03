---
phase: 01-foundation-and-executors
plan: 04
subsystem: build, testing
tags: [makefile, ldflags, go, test-isolation]

# Dependency graph
requires:
  - phase: 01-foundation-and-executors
    provides: CLI structure with version vars, executor package
provides:
  - Makefile with ldflags for version injection
  - FindNodeExecutableWithOptions for isolated testing
  - WithFindOptions NPMOption for test isolation
affects: [all-phases]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Makefile build with ldflags for version injection"
    - "Injectable FindOptions for test isolation"

key-files:
  created:
    - Makefile
  modified:
    - internal/executor/find.go
    - internal/executor/find_test.go
    - internal/executor/npm.go
    - internal/executor/npm_test.go

key-decisions:
  - "FindOptions struct with SkipSystemPaths for test isolation"
  - "WithFindOptions NPMOption to pass through to FindNodeExecutableWithOptions"

patterns-established:
  - "Use FindNodeExecutableWithOptions with SkipSystemPaths: true in tests to avoid host system pollution"
  - "Build via `make build` to get proper version info"

# Metrics
duration: 5min
completed: 2026-02-03
---

# Phase 1 Plan 4: Gap Closure Summary

**Makefile with ldflags version injection plus FindNodeExecutableWithOptions for test isolation**

## Performance

- **Duration:** 5 min
- **Started:** 2026-02-03T06:03:23Z
- **Completed:** 2026-02-03T06:08:23Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments
- Makefile with build/install/test/clean targets using ldflags to inject version/commit/date
- FindOptions struct and FindNodeExecutableWithOptions function for controllable fallback path behavior
- WithFindOptions NPMOption to allow test isolation in NewNPMExecutor
- Updated NotFound tests to use SkipSystemPaths for host-independent testing

## Task Commits

Note: Due to container environment limitations (no git access), commits need to be made manually or by orchestrator:

1. **Task 1: Create Makefile with ldflags for version injection** - pending commit (feat)
2. **Task 2: Fix executor test isolation with injectable fallback paths** - pending commit (fix)

## Files Created/Modified
- `Makefile` - Build targets with ldflags for version injection (-X cmd.version, cmd.commit, cmd.date)
- `internal/executor/find.go` - Added FindOptions struct and FindNodeExecutableWithOptions function
- `internal/executor/find_test.go` - Updated TestFindNodeExecutable_NotFound to use SkipSystemPaths
- `internal/executor/npm.go` - Added WithFindOptions NPMOption and updated NewNPMExecutor
- `internal/executor/npm_test.go` - Updated TestNPMExecutor_NotFound to use WithFindOptions

## Decisions Made
- **FindOptions struct approach:** Chose to add a new struct `FindOptions` with `SkipSystemPaths` boolean rather than modifying function signature, maintaining backward compatibility
- **WithFindOptions NPMOption:** Follows established functional options pattern from Phase 1, allowing test isolation without breaking production usage
- **Wrapper function pattern:** FindNodeExecutable wraps FindNodeExecutableWithOptions(name, nil) for backward compatibility

## Deviations from Plan
None - plan executed exactly as written.

## Issues Encountered
- Container environment lacks `make` and `go` commands, so verification of `make build` and `./m2cv version` must be done on host
- Tests cannot be run in container for same reason

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Makefile ready for use on host system with `make build` to get proper version info
- Test isolation pattern established for all future executor tests
- Recommend running `make test` on host to verify all tests pass with new isolation

---
*Phase: 01-foundation-and-executors*
*Completed: 2026-02-03*
