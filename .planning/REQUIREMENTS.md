# Requirements: m2cv

**Defined:** 2026-02-03
**Core Value:** Given a job description, produce a tailored, professional PDF resume in one pipeline

## v1 Requirements

### Project Setup

- [ ] **INIT-01**: `m2cv init` creates `m2cv.yml` config file with base CV path, default theme, installed themes list, and default Claude model
- [ ] **INIT-02**: `m2cv init` installs `resumed` and user-selected themes via npm in the project directory
- [ ] **INIT-03**: `m2cv init` presents available themes interactively for user selection
- [ ] **INIT-04**: Preflight checks verify `claude` CLI is in PATH before any AI command runs
- [ ] **INIT-05**: Preflight checks verify `resumed` is installed before generate runs

### Application Workflow

- [ ] **WORK-01**: `m2cv apply <job-desc-file>` creates an application folder under `applications/`
- [ ] **WORK-02**: `m2cv apply` auto-names the folder by shelling out to Claude to extract company + role from the job description
- [ ] **WORK-03**: `m2cv apply` copies the job description into the application folder
- [ ] **WORK-04**: `m2cv optimize` writes versioned output files (optimized-cv-1.md, optimized-cv-2.md, etc.)
- [ ] **WORK-05**: Base CV path comes from `m2cv.yml`, overridable with a CLI flag

### Content Optimization

- [ ] **OPT-01**: `m2cv optimize <application-name>` reads base CV + job description, shells out to `claude -p` to produce a tailored CV markdown
- [ ] **OPT-02**: `m2cv optimize` supports `-m` flag to select Claude model

### Export Pipeline

- [ ] **GEN-01**: `m2cv generate <application-name>` converts latest optimized markdown to JSON Resume format via `claude -p`
- [ ] **GEN-02**: `m2cv generate` validates output JSON against JSON Resume schema before export
- [ ] **GEN-03**: `m2cv generate` exports PDF via `resumed` with the configured or specified theme
- [ ] **GEN-04**: `m2cv generate` supports `--theme <name>` flag to select JSON Resume theme

## v2 Requirements

### Content Optimization

- **ATS-01**: `m2cv optimize --ats` adjusts Claude prompt to produce ATS-friendly content (keyword density, parseable formatting)

### Developer Experience

- **DX-01**: Live preview server for resume in browser
- **DX-02**: Diff visualization between base CV and tailored version
- **DX-03**: Analytics/insights on keyword matching and ATS score prediction

## Out of Scope

| Feature | Reason |
|---------|--------|
| Built-in AI/LLM integration | All AI happens via `claude -p` subprocess; no SDK or API keys |
| Web UI or GUI | CLI-only tool; users edit markdown in their own editor |
| Theme creation tooling | Uses existing JSON Resume theme ecosystem |
| Application tracking/CRM | Scope creep; this is a resume generator, not a job search manager |
| Cover letter generation | Different document type; different workflow |
| Resume hosting/publishing | Focus on local file generation |
| Multi-user/team features | Single-user CLI tool |
| Plugin system | Keep core simple |
| Database storage | File-based workflow with git |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| INIT-01 | Phase 2 | Complete |
| INIT-02 | Phase 2 | Complete |
| INIT-03 | Phase 2 | Complete |
| INIT-04 | Phase 1 | Complete |
| INIT-05 | Phase 1 | Complete |
| WORK-01 | Phase 3 | Pending |
| WORK-02 | Phase 3 | Pending |
| WORK-03 | Phase 3 | Pending |
| WORK-04 | Phase 3 | Pending |
| WORK-05 | Phase 1 | Complete |
| OPT-01 | Phase 4 | Pending |
| OPT-02 | Phase 4 | Pending |
| GEN-01 | Phase 5 | Pending |
| GEN-02 | Phase 5 | Pending |
| GEN-03 | Phase 5 | Pending |
| GEN-04 | Phase 5 | Pending |

**Coverage:**
- v1 requirements: 16 total
- Mapped to phases: 16
- Unmapped: 0

---
*Requirements defined: 2026-02-03*
*Last updated: 2026-02-03 after Phase 2 completion*
