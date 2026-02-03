---
phase: 04-content-tailoring
plan: 01
subsystem: cli
tags: [cobra, claude-ai, cv-optimization, ats]

# Dependency graph
requires:
  - phase: 03-application-workflow
    provides: application folder structure, versioning utilities
  - phase: 01-foundation
    provides: executor pattern, config loading, assets embedding
provides:
  - optimize command with --ats and --model flags
  - Claude-based CV tailoring integration
  - versioned output to optimized-cv-N.md
affects: [05-pdf-generation]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - prompt template substitution with string.ReplaceAll
    - model override via ExecuteOption

key-files:
  created:
    - cmd/optimize.go
  modified:
    - cmd/root.go

key-decisions:
  - "Uses persistent flags from root.go (cfgFile, baseCVPath)"
  - "Model determined from config.DefaultModel with flag override"
  - "Job description found via filepath.Glob for *.txt files"

patterns-established:
  - "Command follows apply.go pattern for validation and error messages"
  - "Uses config.FindWithOverrides for config discovery"
  - "Uses application.NextVersionPath for versioned output"

# Metrics
duration: 8min
completed: 2026-02-03
---

# Phase 4 Plan 1: Optimize Command Summary

**Claude-based CV optimization command with standard and ATS modes, reading base CV from config and job description from application folder, writing versioned output**

## Performance

- **Duration:** 8 min
- **Started:** 2026-02-03T08:17:47Z
- **Completed:** 2026-02-03T08:25:59Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments
- Implemented `m2cv optimize <app-name>` command for AI-powered CV tailoring
- Added --ats flag for ATS-optimized output using different prompt template
- Added -m/--model flag to override Claude model from config default
- Integrated with existing versioning system for optimized-cv-N.md output

## Task Commits

Each task was committed atomically:

1. **Task 1: Create optimize command** - `334dfcc` (feat)
2. **Task 2: Wire optimize command to root** - `6927d01` (feat)

## Files Created/Modified
- `cmd/optimize.go` - New optimize subcommand with full implementation
- `cmd/root.go` - Added newOptimizeCommand() registration

## Decisions Made
- Used persistent flags (cfgFile, baseCVPath) from root.go package-level vars
- Model override takes precedence over config.DefaultModel
- Job description found via glob pattern (*.txt) in application folder
- Prompt placeholders use Go template syntax ({{.BaseCV}}, {{.JobDescription}})

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
- Go and git commands not available in container - resolved using Docker containers with host filesystem mounts
- Git worktree required mounting both main .git directory and worktree directory

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- optimize command ready for use
- Requires Claude CLI authentication (handled by preflight check)
- Ready for Phase 4 Plan 2 (if any) or Phase 5 PDF generation

---
*Phase: 04-content-tailoring*
*Completed: 2026-02-03*
