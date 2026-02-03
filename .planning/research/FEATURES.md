# Feature Landscape: CLI Resume/CV Management Tools

**Domain:** CLI-based resume/CV management and JSON Resume ecosystem
**Researched:** 2026-02-03
**Confidence:** MEDIUM (based on training data through January 2025, ecosystem knowledge)

## Executive Summary

The CLI resume/CV tool ecosystem has two main categories:
1. **JSON Resume tools** (resume-cli, resumed, jsonresume.org) — Format converters and exporters
2. **Full-featured resume managers** (HackMyResume, FluentCV) — Multi-format support, validation, optimization

Most tools focus on **format conversion and export**. Few offer **content optimization** or **job application workflow management**. The m2cv project occupies a unique position: combining AI-powered content tailoring with project-based application management.

## Table Stakes

Features users expect in CLI resume tools. Missing these = product feels incomplete or broken.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| **Export to PDF** | Primary output format for job applications | Low | Via puppeteer/chromium or JSON Resume themes |
| **Multiple output formats** | Users need HTML, PDF, sometimes DOCX | Medium | JSON Resume ecosystem provides HTML + PDF; DOCX harder |
| **Theme support** | Visual customization without code | Low | JSON Resume has ~400 themes on npm |
| **Schema validation** | Catch errors before export | Low | JSON Schema validation libraries available |
| **CLI flags for common options** | Output path, theme selection, format choice | Low | Standard CLI UX via cobra/commander |
| **Config file support** | Avoid repeating flags for every command | Low | YAML/JSON config with CLI flag overrides |
| **Human-editable source format** | Must edit resume without learning JSON | Medium | Markdown, YAML, or other structured text |
| **Version control friendly** | Text-based formats that diff well | Low | JSON/YAML/Markdown all work with git |
| **Single binary distribution** | No npm/pip/gem dependency hell for users | Medium | Go achieves this; Node.js tools struggle |
| **Error messages** | Clear feedback when validation/export fails | Low | Good error handling is hygiene |

## Differentiators

Features that set products apart. Not expected, but create competitive advantage.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| **AI-powered content tailoring** | Automatically optimize resume for job description | High | m2cv's core differentiator; requires LLM integration |
| **ATS optimization mode** | Generate ATS-friendly content (keyword density, format) | Medium | Content strategy, not just formatting |
| **Job application workflow** | Manage multiple applications with folder structure | Medium | m2cv's project-based approach is unique |
| **Multi-format input** | Accept JSON Resume, FRESH, LinkedIn export, etc. | High | HackMyResume does this; high complexity |
| **Live preview server** | See changes in browser as you edit | Medium | resume-cli has this (`resume serve`) |
| **Analytics/insights** | Keyword analysis, ATS score prediction | High | Few tools do this; ML/NLP required |
| **Version management** | Track resume iterations per application | Low-Medium | m2cv's versioned optimize output (optimized-cv-1.md, etc.) |
| **Diff visualization** | Show changes between base CV and tailored version | Low | Git diff is free, but in-tool diff is better UX |
| **Auto-naming from job description** | Extract company + role from JD for folder names | Medium | m2cv feature; requires parsing or LLM |
| **Interactive init wizard** | Guided setup with theme preview | Medium | Better onboarding than "read the docs" |
| **Template library** | Pre-built resume structures for different roles | Medium | Reduces cold-start problem |
| **Spell check / grammar check** | Catch typos before export | Low-Medium | Can shell out to external tools |
| **LinkedIn profile import** | Bootstrap resume from LinkedIn data | High | Requires scraping or API; often blocked |

## Anti-Features

Features to explicitly NOT build. Common mistakes or scope creep in this domain.

| Anti-Feature | Why Avoid | What to Do Instead |
|--------------|-----------|-------------------|
| **Built-in LLM/AI** | Ties tool to specific AI provider; bloats binary; API keys/billing | Shell out to `claude` CLI or other AI tools |
| **Resume hosting/publishing** | Becomes a SaaS platform; security/privacy concerns | Focus on local file generation; users host elsewhere |
| **Web UI or GUI** | Different product category; complex to maintain | Stay CLI-only; users can use their own editors |
| **Theme creation tooling** | Massive scope creep; JSON Resume ecosystem already provides this | Use existing jsonresume-theme-* packages |
| **Rich text editing** | Wrong abstraction; users have preferred editors | Markdown/YAML editing happens in user's editor |
| **Built-in spell check (complex)** | Better tools exist (grammarly, languagetool CLI) | Provide hooks for external tools if needed |
| **Social media integration** | Privacy concerns; API rate limits; maintenance burden | Users manually update their profiles |
| **Application tracking** | Becomes a job search CRM; scope creep | Stay focused on resume generation |
| **Email/submission integration** | Requires email credentials; security liability | Users submit resumes via their own email |
| **Cover letter generation** | Different document type; different workflow | Could be separate tool or future extension |
| **Multi-user/team features** | Not a use case for personal resume tools | Single-user CLI only |
| **Database storage** | Adds dependency; overkill for file-based workflow | Files in folders, version control with git |
| **Plugin system** | Maintenance nightmare for CLI tools | Keep core simple, provide clear subprocess hooks |

## Feature Dependencies

Dependencies show which features must be built in what order.

```
Core foundation:
  Config file (m2cv.yml)
    └─> Project init (m2cv init)
          └─> Theme installation
          └─> Base CV setup

Application workflow:
  Application creation (m2cv apply)
    └─> Content optimization (m2cv optimize)
          └─> Version management (optimized-cv-N.md)
          └─> ATS mode (--ats flag)
    └─> Export pipeline (m2cv generate)
          └─> JSON Resume conversion
          └─> Schema validation
          └─> PDF export (via resumed)
          └─> Theme selection

Prerequisites (external):
  - claude CLI (must be in PATH)
  - resumed + themes (installed by m2cv init)
```

**Critical path for MVP:**
1. Config file support (foundation for everything)
2. Init command (sets up project + installs dependencies)
3. Apply command (creates application folder)
4. Optimize command (core AI integration)
5. Generate command (export pipeline)

**Can defer:**
- Live preview (nice-to-have)
- Diff visualization (git diff works)
- Interactive wizards (docs first, UX later)
- Analytics/insights (post-MVP)

## Feature Categories by Complexity

### Low Complexity (< 1 day)
- CLI flags and argument parsing
- Config file loading (YAML)
- Theme selection/switching
- Error message formatting
- Version numbering for optimize output
- Preflight checks (claude in PATH, resumed installed)

### Medium Complexity (1-3 days)
- Project init workflow (npm install resumed + themes)
- Application folder creation + naming
- JSON Resume schema validation
- Markdown ↔ JSON Resume conversion
- ATS mode prompt variations
- Multi-format export coordination

### High Complexity (> 3 days)
- AI-powered content optimization (prompt engineering + testing)
- Multi-format input support (JSON Resume + FRESH + LinkedIn)
- Analytics/insights (NLP/keyword analysis)
- Live preview server (file watching + hot reload)
- Interactive wizards with theme preview

## Domain-Specific Feature Patterns

### Pattern 1: Format Conversion Pipeline
**What:** Input format → normalized data → output format(s)
**Why it matters:** Resume tools must support multiple formats
**m2cv approach:** Markdown (input) → JSON Resume (intermediate) → PDF/HTML (output)

**Key insight:** Intermediate representation (JSON Resume) enables:
- Schema validation
- Theme ecosystem compatibility
- Multiple output formats from single conversion

### Pattern 2: Content vs Presentation Separation
**What:** Resume content lives separately from styling/themes
**Why it matters:** Users iterate on content more than themes
**m2cv approach:** Markdown CV (content) + JSON Resume themes (presentation)

**Key insight:** Don't bake styling into content format. Use theme ecosystem.

### Pattern 3: Versioned Iterations
**What:** Keep history of tailored resumes per application
**Why it matters:** Users iterate on AI-generated content, need undo
**m2cv approach:** `optimized-cv-1.md`, `optimized-cv-2.md`, etc.

**Key insight:** Don't overwrite; increment version. Disk is cheap, lost work is expensive.

### Pattern 4: Subprocess Integration
**What:** Shell out to external tools instead of embedding
**Why it matters:** Keeps binary small, leverages existing tools
**m2cv approach:** `claude -p` for AI, `resumed` for PDF export

**Key insight:** Treat specialized tools as dependencies, not as code to rewrite.

## MVP Recommendation

For MVP, prioritize:

### Must Have (Table Stakes)
1. Export to PDF via resumed
2. Theme support (JSON Resume themes)
3. Schema validation
4. CLI flags for common options
5. Config file (m2cv.yml)
6. Clear error messages

### Must Have (Differentiators)
1. AI-powered content tailoring (optimize command)
2. Job application workflow (apply command)
3. Version management (optimized-cv-N.md)
4. ATS optimization mode (--ats flag)
5. Auto-naming from job description

### Can Defer to Post-MVP
- Live preview server (use `resumed serve` manually if needed)
- Diff visualization (git diff works)
- Interactive init wizard (docs + simple prompts OK for MVP)
- Analytics/insights (need user feedback first)
- Multi-format input (JSON Resume only for MVP)
- Spell check integration (users have their own tools)
- Template library (one good example CV in docs is enough)

## Feature Prioritization Matrix

| Feature | User Value | Complexity | Priority |
|---------|------------|------------|----------|
| AI content tailoring | Very High | High | P0 (core differentiator) |
| Application workflow | High | Medium | P0 (core differentiator) |
| PDF export | Very High | Low | P0 (table stakes) |
| Version management | High | Low | P0 (low cost, high value) |
| ATS mode | High | Medium | P0 (key use case) |
| Theme support | High | Low | P0 (table stakes) |
| Schema validation | Medium | Low | P0 (prevents errors) |
| Config file | Medium | Low | P0 (UX improvement) |
| Auto-naming | Medium | Medium | P1 (nice QoL) |
| Live preview | Medium | Medium | P2 (workaround exists) |
| Diff visualization | Low | Low | P2 (git diff works) |
| Analytics/insights | Low | High | P3 (unproven value) |
| Multi-format input | Low | High | P3 (YAGNI) |

## Competitive Positioning

**m2cv's unique position:**
- Only tool combining AI content optimization with workflow management
- Project-based approach (application folders) is novel
- Subprocess model (claude CLI) avoids vendor lock-in
- Go binary distribution (no npm install for users)

**Table stakes covered by JSON Resume ecosystem:**
- Theme variety (400+ themes on npm)
- Export formats (HTML, PDF)
- Schema validation

**Gap in market:**
- No tool offers job-specific CV optimization + workflow tracking
- Existing tools focus on format conversion, not content strategy
- m2cv bridges the gap between "resume formatter" and "application manager"

## Sources

**Confidence level: MEDIUM**

This research is based on:
- Training data knowledge of JSON Resume ecosystem (jsonresume.org)
- Familiarity with resume-cli, resumed, HackMyResume, FluentCV
- Understanding of CLI tool design patterns
- Analysis of m2cv's PROJECT.md and README.md

**Limitations:**
- Did not verify current state of resume-cli vs resumed (used project context)
- Did not check for new tools launched after January 2025
- Theme count (~400) is approximate based on npm search patterns
- Feature comparisons based on training data, not live verification

**Verification needed:**
- Current JSON Resume theme count and popular themes
- Resumed feature set vs resume-cli (project context says resumed is active fork)
- Existence of competitors launched 2025-2026
- ATS optimization approaches in existing tools

---

**Ready for downstream use:** Requirements definition can now categorize features clearly as table stakes, differentiators, or anti-features. Phase structure should prioritize P0 features for MVP.
