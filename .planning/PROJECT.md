# m2cv — Markdown to CV

## What This Is

A Go CLI tool for managing job applications. It takes a base CV in markdown format and a job description, uses Claude CLI to tailor the CV for that specific role, converts it to JSON Resume format, and exports a themed PDF via `resumed`. Each job application lives in its own folder within a project directory.

## Core Value

Given a job description, produce a tailored, professional PDF resume in one pipeline — no manual format wrangling, no copy-pasting between tools.

## Requirements

### Validated

(None yet — ship to validate)

### Active

- [ ] `m2cv init` initializes a project: creates `m2cv.yml` config, installs `resumed` and user-selected themes via npm, sets base CV path
- [ ] `m2cv apply <job-desc-file>` creates an application folder under `applications/`, auto-names it by shelling out to Claude to extract company + role from the job description, copies the job description into the folder
- [ ] `m2cv optimize [--ats] <application-name>` reads base CV + job description, shells out to `claude -p` to tailor the CV, writes versioned output (optimized-cv-1.md, optimized-cv-2.md, etc.)
- [ ] `m2cv generate [--theme <name>] <application-name>` converts the latest optimized markdown to JSON Resume via `claude -p`, validates against JSON Resume schema, exports PDF via `resumed`
- [ ] `--ats` flag on optimize adjusts the Claude prompt to produce ATS-friendly content
- [ ] Config file `m2cv.yml` stores: base CV path, default theme, installed themes list, default Claude model
- [ ] JSON Resume schema validation before PDF export (using `santhosh-tekuri/jsonschema/v6`)
- [ ] Preflight checks: `claude` CLI must be in PATH; `resumed` checked at generate time
- [ ] Base CV path configurable via `m2cv.yml`, overridable with a CLI flag

### Out of Scope

- Built-in AI/LLM integration — all AI happens via `claude -p` subprocess
- Web UI or GUI — CLI only
- Resume editing within the tool — user edits markdown in their own editor
- Theme creation — uses existing JSON Resume theme ecosystem
- Hosting or publishing resumes online

## Context

- The JSON Resume ecosystem provides structured resume data + themed PDF/HTML export
- `resume-cli` (original) is unmaintained; `resumed` (https://github.com/rbardini/resumed) is the active fork
- Themes are npm packages (`jsonresume-theme-*`), ~400 available on npm
- Default theme candidates from jsonresume.org/themes: data-driven, developer-mono, flat, consultant-polished, bold-header-statement — but user selects at init time
- The markdown CV format uses YAML frontmatter for contact/basics and heading conventions for sections (Experience, Education, Skills, etc.) that map to JSON Resume schema fields
- Claude CLI (`claude -p --output-format text`) accepts prompts via stdin, avoiding shell argument limits

## Constraints

- **Language**: Go — CLI tool, single binary distribution
- **AI integration**: Must use `claude -p` subprocess only, no SDK or API calls
- **PDF export**: Must use `resumed` + JSON Resume themes (npm ecosystem)
- **Config format**: YAML (`m2cv.yml`)
- **Dependencies**: `github.com/spf13/cobra` for CLI, `github.com/santhosh-tekuri/jsonschema/v6` for schema validation

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Go for implementation | Single binary, good CLI ecosystem (cobra) | — Pending |
| `resumed` over `resume-cli` | resume-cli is unmaintained, resumed is actively maintained | — Pending |
| Project-based workflow with application folders | Keeps each job application organized, supports versioning | — Pending |
| Claude CLI via subprocess | No AI dependencies in binary, user controls their own Claude setup | — Pending |
| Versioned optimize output | Supports iteration without losing previous versions | — Pending |
| ATS flag on optimize (not generate) | Content strategy is a writing concern, not a formatting concern | — Pending |
| Auto-name application folders via Claude | Reduces friction, consistent naming from job description content | — Pending |

---
*Last updated: 2026-02-03 after initialization*
