# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-03)

**Core value:** Given a job description, produce a tailored, professional PDF resume in one pipeline
**Current focus:** Phase 4 - Content Tailoring complete, ready for Phase 5

## Current Position

Phase: 4 of 5 (Content Tailoring)
Plan: 2 of 2 in current phase (all plans completed)
Status: Phase 4 complete, ready for Phase 5
Last activity: 2026-02-03 - Completed 04-02-PLAN.md (optimize tests)

Progress: [████████░░] 80%

## Performance Metrics

**Velocity:**
- Total plans completed: 11
- Average duration: 4.3 min
- Total execution time: 50 min

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 1. Foundation | 4 | 20 min | 5 min |
| 2. Project Init | 2 | 8 min | 4 min |
| 3. Application | 3 | 10 min | 3.3 min |
| 4. Content Tailoring | 2 | 12 min | 6 min |

**Recent Trend:**
- Last 5 plans: 03-01 (3 min), 03-02 (4 min), 03-03 (3 min), 04-01 (8 min), 04-02 (4 min)
- Trend: Stable

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
- Uses persistent flags from root.go (cfgFile, baseCVPath) for optimize command
- Model determined from config.DefaultModel with flag override
- Job description found via filepath.Glob for *.txt files
- Avoid t.Parallel() on tests that call NewRootCommand (writes to global vars)
- Use t.Skip() when test conditions can't be met on host system

### Pending Todos

None yet.

### Blockers/Concerns

None yet.

## Session Continuity

Last session: 2026-02-03T08:32:15Z
Stopped at: Completed 04-02-PLAN.md (optimize tests)
Resume file: None
