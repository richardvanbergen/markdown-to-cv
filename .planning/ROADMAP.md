# Roadmap: m2cv — Markdown to CV

## Overview

m2cv delivers a streamlined pipeline from job description to tailored PDF resume through five phases. Starting with core infrastructure and subprocess handling, we build project initialization, then layer on the complete workflow: application creation, AI-powered CV optimization, and professional PDF export. Each phase delivers a working capability that builds toward the complete tool.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [x] **Phase 1: Foundation & Executors** - Core infrastructure, subprocess handling, embedded assets
- [x] **Phase 2: Project Initialization** - Project setup with config, npm integration, theme selection
- [ ] **Phase 3: Application Workflow** - Job application creation with AI-powered folder naming
- [ ] **Phase 4: Content Tailoring** - AI-powered CV optimization with versioning and ATS mode
- [ ] **Phase 5: Export Pipeline** - JSON Resume conversion, schema validation, themed PDF export

## Phase Details

### Phase 1: Foundation & Executors
**Goal**: Core infrastructure enables reliable subprocess execution and asset management
**Depends on**: Nothing (first phase)
**Requirements**: INIT-04, INIT-05, WORK-05
**Success Criteria** (what must be TRUE):
  1. ClaudeExecutor can invoke claude CLI and stream output without buffer deadlocks
  2. NPMExecutor can find npm/npx in PATH across different node version managers
  3. Config repository can load m2cv.yml from current directory or parent directories
  4. Embedded prompts and JSON Resume schema compile into binary
  5. Preflight checks detect missing claude or resumed before commands fail
**Plans**: 4 plans

Plans:
- [x] 01-01-PLAN.md — Project scaffolding, config repository, embedded assets
- [x] 01-02-PLAN.md — ClaudeExecutor and NPMExecutor (TDD)
- [x] 01-03-PLAN.md — CLI skeleton with cobra and preflight checks
- [x] 01-04-PLAN.md — Gap closure: Makefile + test isolation

### Phase 2: Project Initialization
**Goal**: Users can initialize m2cv projects with config, themes, and dependencies
**Depends on**: Phase 1
**Requirements**: INIT-01, INIT-02, INIT-03
**Success Criteria** (what must be TRUE):
  1. User runs `m2cv init` and gets valid m2cv.yml with base CV path, default theme, default model
  2. `resumed` package is installed in project via npm
  3. User can select from available JSON Resume themes interactively during init
  4. Selected theme is installed via npm and recorded in m2cv.yml
**Plans**: 2 plans

Plans:
- [x] 02-01-PLAN.md — Init service and theme selector (foundation)
- [x] 02-02-PLAN.md — Init command with cobra integration

### Phase 3: Application Workflow
**Goal**: Users can create organized job applications with AI-extracted folder names
**Depends on**: Phase 2
**Requirements**: WORK-01, WORK-02, WORK-03, WORK-04
**Success Criteria** (what must be TRUE):
  1. User runs `m2cv apply <job-desc.txt>` and application folder is created under applications/
  2. Folder name is auto-generated from job description (company-role format) via Claude
  3. Job description file is copied into the application folder
  4. Application folder structure supports versioned CV files (optimized-cv-N.md pattern)
**Plans**: TBD

Plans:
- [ ] TBD (determined during planning)

### Phase 4: Content Tailoring
**Goal**: Users can tailor CVs to job descriptions with AI optimization and version management
**Depends on**: Phase 3
**Requirements**: OPT-01, OPT-02
**Success Criteria** (what must be TRUE):
  1. User runs `m2cv optimize <app-name>` and gets tailored CV markdown in versioned file
  2. Optimizer reads base CV from configured path and job description from application folder
  3. Claude receives appropriate prompt (standard or ATS mode) with CV + job description context
  4. Output writes to optimized-cv-N.md with auto-incrementing version numbers
  5. User can override Claude model with -m flag
**Plans**: TBD

Plans:
- [ ] TBD (determined during planning)

### Phase 5: Export Pipeline
**Goal**: Users can convert tailored CVs to professional themed PDFs via JSON Resume
**Depends on**: Phase 4
**Requirements**: GEN-01, GEN-02, GEN-03, GEN-04
**Success Criteria** (what must be TRUE):
  1. User runs `m2cv generate <app-name>` and gets PDF output
  2. Markdown CV (YAML frontmatter + sections) converts to valid JSON Resume format via Claude
  3. JSON output validates against JSON Resume schema before PDF generation
  4. PDF exports via resumed with configured or specified theme
  5. User can override theme with --theme flag
**Plans**: TBD

Plans:
- [ ] TBD (determined during planning)

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → 4 → 5

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Foundation & Executors | 4/4 | Complete | 2026-02-03 |
| 2. Project Initialization | 2/2 | Complete | 2026-02-03 |
| 3. Application Workflow | 0/TBD | Not started | - |
| 4. Content Tailoring | 0/TBD | Not started | - |
| 5. Export Pipeline | 0/TBD | Not started | - |
