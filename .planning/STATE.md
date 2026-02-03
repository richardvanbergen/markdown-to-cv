# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-03)

**Core value:** Given a job description, produce a tailored, professional PDF resume in one pipeline
**Current focus:** Phase 3 - Application Workflow (COMPLETE)

## Current Position

Phase: 3 of 5 (Application Workflow) - Complete
Plan: 3 of 3 in current phase (all plans completed)
Status: Phase 3 complete, verified, ready for Phase 4
Last activity: 2026-02-03 - Completed Phase 3 execution and verification

Progress: [██████░░░░] 60%

## Performance Metrics

**Velocity:**
- Total plans completed: 9
- Average duration: 4.2 min
- Total execution time: 38 min

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 1. Foundation | 4 | 20 min | 5 min |
| 2. Project Init | 2 | 8 min | 4 min |
| 3. Application | 3 | 10 min | 3.3 min |

**Recent Trend:**
- Last 5 plans: 02-01 (6 min), 02-02 (2 min), 03-01 (3 min), 03-02 (4 min), 03-03 (3 min)
- Trend: Stable (consistent fast execution)

*Updated after each plan completion*

## Accumulated Context

### Decisions

Decisions are logged in PROJECT.md Key Decisions table.
Recent decisions affecting current work:

- Go for implementation (single binary, good CLI ecosystem)
- `resumed` over `resume-cli` (actively maintained)
- Project-based workflow with application folders (keeps jobs organized)
- Claude CLI via subprocess (no AI dependencies in binary)
- Versioned optimize output (supports iteration)
- ATS flag on optimize not generate (content strategy is writing concern)
- Auto-name application folders via Claude (reduces friction)
- Repository interface pattern for config (testability)
- go:embed for prompts and schema (single binary)
- Walk-up config discovery (mimics git pattern)
- Functional options for executor configuration (clean API)
- CheckInstalled uses filesystem not npm command (faster)
- CheckClaude only in PersistentPreRunE (all functional commands need Claude)
- CheckResumed is command-specific (only generate needs it)
- FindOptions with SkipSystemPaths for test isolation
- WithFindOptions NPMOption for passing FindOptions to executor
- huh v0.8.0 for interactive theme selection (latest stable)
- 8 curated themes: even, stackoverflow, elegant, actual, class, flat, kendall, macchiato
- ErrAlreadyInitialized sentinel error for clear error handling
- Service layer with constructor injection pattern
- Init uses local --base-cv flag separate from root's persistent flag
- Default Claude model set to claude-sonnet-4-20250514 for new projects
- Force flag removes existing m2cv.yml before calling service
- Repository interface for filesystem operations (matches config pattern)
- Max filename length 50 chars with word-boundary truncation
- Mock executor in test file for Claude integration testing
- Sanitize --name flag input through extractor.SanitizeFilename
- Disable preflight checks in tests using PersistentPreRunE = nil
- filepath.Glob for version pattern matching (cleaner than manual dir reading)
- Empty directories return empty slice not error (graceful first-version handling)

### Pending Todos

None yet.

### Blockers/Concerns

None yet.

## Session Continuity

Last session: 2026-02-03T07:52:00Z
Stopped at: Completed Phase 3 execution and verification
Resume file: None
