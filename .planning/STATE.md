# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-03)

**Core value:** Given a job description, produce a tailored, professional PDF resume in one pipeline
**Current focus:** Phase 3 - Application Workflow (IN PROGRESS)

## Current Position

Phase: 3 of 5 (Application Workflow)
Plan: 1 of N in current phase
Status: In progress - Plan 03-01 complete
Last activity: 2026-02-03 - Completed 03-01-PLAN.md (filesystem & extractor)

Progress: [███░░░░░░░] 30%

## Performance Metrics

**Velocity:**
- Total plans completed: 5
- Average duration: 4.6 min
- Total execution time: 23 min

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 1. Foundation | 4 | 20 min | 5 min |
| 3. Application | 1 | 3 min | 3 min |

**Recent Trend:**
- Last 5 plans: 01-02 (7 min), 01-03 (3 min), 01-04 (5 min), 03-01 (3 min)
- Trend: Improving (faster execution)

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
- Repository interface for filesystem operations (matches config pattern)
- Max filename length 50 chars with word-boundary truncation
- Mock executor in test file for Claude integration testing

### Pending Todos

None yet.

### Blockers/Concerns

None yet.

## Session Continuity

Last session: 2026-02-03T07:39:49Z
Stopped at: Completed 03-01-PLAN.md (filesystem & extractor)
Resume file: None
