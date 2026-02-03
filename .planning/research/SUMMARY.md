# Project Research Summary

**Project:** m2cv — Markdown to CV CLI Tool
**Domain:** CLI resume/CV management with AI-powered content optimization
**Researched:** 2026-02-03
**Confidence:** MEDIUM-HIGH

## Executive Summary

m2cv is a Go CLI tool that orchestrates AI-powered CV tailoring through subprocess integration. It occupies a unique market position: combining Claude AI content optimization with project-based job application workflow management. The tool bridges two distinct ecosystems—Go (for CLI orchestration, validation, file management) and npm (for PDF theming via resumed and JSON Resume).

The recommended architectural approach follows the Command-Executor-Service pattern: thin cobra commands delegate to service layer business logic, which coordinates interface-based executors (ClaudeExecutor, NPMExecutor) and repository abstractions (ConfigRepository, FileRepository). Embedded prompts and schema enable single-binary distribution. The hybrid Go+npm stack is constrained by requirements (Go language, cobra CLI, resumed for PDF export) but well-suited to the domain.

Critical risks center on subprocess management: unbuffered stdout/stderr will cause deadlocks with large Claude responses, PATH resolution differences between dev/prod will break npm integration, and temp file cleanup races will corrupt concurrent operations. These pitfalls must be addressed in Phase 1 foundation. Success depends on treating LLM calls as unreliable network services (retry logic, validation, progress indicators) rather than deterministic functions.

## Key Findings

### Recommended Stack

m2cv uses a hybrid architecture: Go 1.22+ for the CLI layer and Node.js 18+ LTS as the runtime for resumed PDF export. The Go stack is largely constrained by project requirements but represents industry-standard choices.

**Core technologies:**
- **Go 1.22+ with cobra v1.8+**: Single binary CLI with excellent subprocess management; cobra is the de facto standard for Go CLIs (kubectl, gh, hugo)
- **gopkg.in/yaml.v3**: YAML config parsing; simpler than viper for file-based configuration without env var merging complexity
- **github.com/santhosh-tekuri/jsonschema/v6**: JSON Resume schema validation; supports draft-07 used by JSON Resume standard
- **resumed 4.0+ (npm)**: PDF/HTML export from JSON Resume; actively maintained fork of resume-cli with 400+ theme ecosystem
- **embed (stdlib)**: Compile prompts and schema into binary for zero-dependency distribution
- **os/exec (stdlib)**: Subprocess management for claude and npm; sufficient without external process libraries

**Key architectural decision:** Custom markdown parsing instead of goldmark/gomarkdown. The CV markdown format is constrained and documented; full AST parsers are overkill for extracting YAML frontmatter and structured sections. This reduces dependencies but requires careful implementation and comprehensive testing.

**Confidence:** HIGH on core Go stack (project constraints + industry standard), MEDIUM on custom markdown parsing (implementation risk, but architecturally sound)

### Expected Features

The CLI resume tool ecosystem splits between format converters (resume-cli, resumed) and full-featured managers (HackMyResume, FluentCV). m2cv differentiates through AI-powered content tailoring and job application workflow management—features missing from existing tools.

**Must have (table stakes):**
- Export to PDF via JSON Resume themes (primary output format for job applications)
- Theme support (visual customization; 400+ themes available on npm)
- Schema validation (catch errors before export)
- CLI flags for common options (output path, theme selection, format choice)
- Config file support (m2cv.yml to avoid repeating flags)
- Human-editable source format (markdown for base CV)
- Single binary distribution (no dependency hell for end users)
- Clear error messages

**Must have (differentiators):**
- AI-powered content tailoring (optimize command; core value proposition)
- Job application workflow (apply command; project-based folder structure)
- Version management (optimized-cv-1.md, optimized-cv-2.md, etc.; enables iteration)
- ATS optimization mode (--ats flag; keyword density, format strategy)
- Auto-naming from job description (extract company + role for folder naming)

**Defer (v2+):**
- Live preview server (resumed serve exists; users can run manually)
- Diff visualization (git diff works; in-tool UX is better but not essential)
- Interactive init wizard (docs + simple prompts sufficient for MVP)
- Analytics/insights (keyword analysis, ATS scoring; need user feedback first)
- Multi-format input support (JSON Resume only for MVP; FRESH, LinkedIn import later)
- Template library (one good example CV in docs is enough)

**Anti-features to explicitly avoid:**
- Built-in LLM/AI (bloats binary; shell out to claude CLI instead)
- Resume hosting/publishing (becomes SaaS; security/privacy concerns)
- Web UI or GUI (different product category; stay CLI-focused)
- Theme creation tooling (massive scope creep; use existing jsonresume-theme-* packages)
- Application tracking CRM (scope creep; focus on resume generation)
- Database storage (overkill; files in folders work fine)

### Architecture Approach

The Command-Executor-Service pattern separates concerns into four layers: CLI (cobra commands), Service (business logic), Executor (subprocess abstraction), and Persistence (filesystem + config). This enables testability through dependency injection while maintaining cobra's flag binding ergonomics.

**Major components:**
1. **CLI Layer (cmd/)**: Thin cobra commands handling flag parsing, user interaction, and output formatting. Root command includes preflight checks (claude in PATH). Commands delegate to service layer immediately after validation.
2. **Service Layer (internal/service/)**: InitService, ApplyService, OptimizeService, GenerateService orchestrate executors, repositories, and templates to implement business logic. Services are constructor-injected with dependencies (no global state).
3. **Executor Layer (internal/executor/)**: ClaudeExecutor and NPMExecutor interfaces abstract subprocess invocation. Implementations handle stdout/stderr streaming, exit code checking, and stderr capture. Interface-based design enables mocking for unit tests.
4. **Persistence Layer (internal/persistence/)**: ConfigRepository (m2cv.yml), FileRepository (application folders, versioned files), JSONValidator (schema validation). Repository pattern abstracts filesystem for testability.
5. **Template Engine (internal/template/)**: TemplateRenderer loads embedded prompts (prompts/*.txt) and schema (schema/*.json) via embed.FS. Go text/template for variable substitution.

**Data flow example (optimize command):**
1. cobra command parses --ats flag and application name
2. OptimizeService.OptimizeCV() loads config, reads base CV and job description
3. TemplateRenderer selects optimize.txt or optimize-ats.txt prompt template
4. ClaudeExecutor pipes prompt to `claude -p`, streams stdout/stderr concurrently
5. FileRepository determines next version number, writes optimized-cv-N.md
6. Command outputs success message with file path

**Key patterns:**
- Interface-based executors for subprocess abstraction and testing
- Repository pattern for filesystem operations (enables in-memory testing)
- Embedded assets (embed.FS) for single binary distribution
- Dependency injection via constructors (no global state)
- PersistentPreRunE for preflight checks (claude in PATH)
- Streaming stdout/stderr to avoid buffer deadlocks
- Config file discovery (walk up directory tree like git)

### Critical Pitfalls

**1. Unbuffered stdout/stderr Reading from Subprocesses (CRITICAL)**
Claude responses can be kilobytes of text. Using `cmd.Output()` directly causes deadlocks when stdout buffer fills (typically 64KB on Unix). The subprocess blocks waiting for buffer consumption while Go's `cmd.Wait()` blocks waiting for process exit. Prevention: Use concurrent stdout/stderr streaming with `bytes.Buffer` for all subprocess calls. Phase 1 blocker.

**2. No JSON Schema Validation for LLM Output (CRITICAL)**
LLMs are non-deterministic. Claude may wrap JSON in markdown fences, add commentary, use wrong field names, or return partial output. Direct `json.Unmarshal()` fails randomly. Prevention: Implement `ParseClaudeJSON()` helper that strips markdown fences, finds JSON boundaries, validates parseability, and runs schema validation before returning. Phase 1 blocker.

**3. PATH Dependency Hell for npx/npm (CRITICAL)**
Go's `exec.LookPath()` uses different PATH than interactive shells. nvm/asdf/volta modify PATH in shell init scripts that aren't loaded for non-interactive execution. CLI works on dev machine, fails in CI/prod with "npx: command not found". Prevention: Implement `findNodeExecutable()` that checks LookPath first, then common node manager locations (~/.nvm/current/bin, ~/.volta/bin, ~/.asdf/shims, /usr/local/bin, /opt/homebrew/bin). Phase 1 blocker.

**4. Race Conditions in Temp File Cleanup (CRITICAL)**
Using `defer os.Remove(tempFile)` before subprocess completes. Sequence: Go creates temp prompt file → spawns `claude -p /tmp/prompt123.txt` → defer fires → file deleted → Claude tries to read → ENOENT. Prevention: Only cleanup temp files after `cmd.Wait()` completes. Phase 1 blocker.

**5. No Retry Logic for LLM Calls (MODERATE)**
Treating Claude like a deterministic function instead of a network service. Transient failures (rate limits, network blips, timeouts) cause entire operation to fail. Prevention: Implement exponential backoff retry (3 attempts max) with non-retryable error detection (auth failures, invalid prompts). Phase 2 reliability enhancement.

## Implications for Roadmap

Based on research, suggested phase structure organized by architectural layer dependencies and pitfall avoidance:

### Phase 1: Foundation & Core Utilities
**Rationale:** Establish interfaces, embedded assets, and persistence layer before executors and services depend on them. Critical pitfall avoidance (subprocess buffering, PATH resolution, temp file lifecycle) must be designed in from the start—retrofitting is expensive.

**Delivers:**
- Package structure (internal/executor, internal/persistence, internal/service, internal/template)
- Interface definitions (Executor, Repository)
- Embedded assets structure with prompts and JSON Resume schema
- Config repository (m2cv.yml loading/saving with directory tree walk-up)
- File repository (application folder creation, versioned file handling)
- Template renderer (embed.FS prompt loading with text/template)

**Addresses:**
- Config file support (table stakes)
- Version management foundation (differentiator)
- Hardcoded config paths pitfall (#6) through directory walk-up

**Avoids:**
- Designing subprocess handling correctly from start prevents Pitfall #1, #4 rework

**Research needs:** None (standard Go patterns)

### Phase 2: Subprocess Executors
**Rationale:** Executors depend only on Phase 1 interfaces. Can be developed/tested in parallel with services. Must implement streaming stdout/stderr, PATH resolution, and stderr capture from the start to avoid critical pitfalls.

**Delivers:**
- ClaudeExecutor with streaming output, context timeout, retry logic, stderr capture
- NPMExecutor with PATH resolution, package existence validation, environment setup
- Subprocess helper utilities (findNodeExecutable, ParseClaudeJSON, CommandError wrapper)
- Debug mode (--debug flag to log subprocess invocations and output)

**Addresses:**
- AI-powered content tailoring foundation (differentiator)
- Clear error messages (table stakes)

**Avoids:**
- Pitfall #1: Streaming stdout/stderr prevents deadlocks
- Pitfall #3: PATH resolution prevents dev/prod inconsistency
- Pitfall #4: Proper temp file lifecycle prevents race conditions
- Pitfall #11: CommandError wrapper standardizes error formatting
- Pitfall #12: Debug mode enables troubleshooting

**Research needs:** Verify current claude CLI flags (claude --help), resumed API, and npm package version constraints

### Phase 3: Init & Apply Services
**Rationale:** Simplest services to implement. Init sets up project structure (required for all other commands). Apply introduces Claude integration with simpler use case (name extraction) before complex CV optimization.

**Delivers:**
- InitService: Project initialization, npm install resumed + theme, m2cv.yml creation
- ApplyService: Job description parsing, Claude-powered name extraction, folder creation
- Input validation (markdown structure, file size, required sections)
- Preflight checks (claude in PATH, resumed installed)

**Addresses:**
- Application workflow (differentiator)
- Auto-naming from job description (differentiator)
- Schema validation foundation (table stakes)

**Avoids:**
- Pitfall #2: JSON extraction layer tested with simpler name extraction before complex CV parsing
- Pitfall #5: npm package validation before npx invocation
- Pitfall #9: Input validation prevents wasted API calls

**Research needs:** None (uses executors from Phase 2)

### Phase 4: Optimize Service
**Rationale:** Builds on Apply pattern but adds complexity: base CV reading, ATS mode prompt selection, version management. Core differentiator feature.

**Delivers:**
- OptimizeService: CV tailoring orchestration with prompt template selection
- ATS mode (--ats flag) with specialized prompt
- Version numbering (optimized-cv-N.md)
- Progress indicators for long-running Claude calls

**Addresses:**
- AI-powered content tailoring (core differentiator)
- ATS optimization mode (differentiator)
- Version management (differentiator)

**Avoids:**
- Pitfall #7: Retry logic with exponential backoff
- Pitfall #8: Progress indicators prevent "CLI is hung" perception

**Research needs:** None (uses executors from Phase 2)

### Phase 5: Generate Service
**Rationale:** Most complex service. Depends on JSON Resume schema validation, markdown-to-JSON conversion, resumed integration. Final piece of the workflow.

**Delivers:**
- GenerateService: Markdown → JSON Resume conversion, schema validation, PDF export
- Markdown parser (YAML frontmatter + section extraction)
- JSON Resume schema validation
- Multi-format output coordination (JSON, HTML, PDF)
- Theme selection

**Addresses:**
- Export to PDF (table stakes)
- Theme support (table stakes)
- Schema validation (table stakes)

**Avoids:**
- Pitfall #2: Full JSON schema validation before PDF export
- Pitfall #14: Version pinning for resumed package

**Research needs:** JSON Resume schema structure, markdown CV format conventions

### Phase 6: CLI Commands & Integration
**Rationale:** Thin layer over services. Implemented after services are fully tested. Wires everything together with cobra command structure.

**Delivers:**
- Root command with preflight checks (PersistentPreRunE)
- init, apply, optimize, generate commands
- Flag definitions and help text
- Main entry point with dependency injection
- End-to-end integration testing

**Addresses:**
- CLI flags for common options (table stakes)
- Single binary distribution (table stakes)

**Avoids:**
- Pitfall #10: Context cancellation with process group signal handling

**Research needs:** None (assembles existing components)

### Phase Ordering Rationale

**Critical path:** Interfaces (Phase 1) → Executors (Phase 2) → Services (Phase 3-5) → Commands (Phase 6)

**Rationale for this order:**
1. Phase 1 establishes contracts; all other phases depend on these interfaces
2. Phase 2 executors can be developed in parallel with Phase 1 utilities (both depend only on interfaces)
3. Phase 3-5 services build in complexity: Init (simplest) → Apply (introduces Claude) → Optimize (core feature) → Generate (most complex)
4. Phase 6 CLI wiring happens last when all services are testable

**Pitfall avoidance strategy:**
- Critical pitfalls (#1-4) addressed in Phase 1-2 foundation—no retrofitting needed
- Moderate pitfalls (#6-10) addressed incrementally as relevant services are built
- Minor pitfalls (#11-14) handled as polish in final phases

**Dependency insights from architecture research:**
- Persistence layer (Phase 1) must exist before services can store results
- Executors (Phase 2) must be interface-based for service testability
- Services (Phase 3-5) orchestrate executors + repositories; can't exist before them
- CLI commands (Phase 6) are thin wrappers; implement after business logic is tested

### Research Flags

**Phases needing deeper research during planning:**
- **Phase 5 (Generate):** JSON Resume schema structure, markdown CV format conventions, resumed API surface area. Niche domain with less community knowledge than general Go CLI patterns.

**Phases with standard patterns (skip research-phase):**
- **Phase 1 (Foundation):** Well-documented Go patterns (interfaces, embed.FS, repository pattern)
- **Phase 2 (Executors):** Standard subprocess management; pitfalls are known and researched
- **Phase 3 (Init/Apply):** Straightforward service orchestration
- **Phase 4 (Optimize):** Builds on Apply pattern with prompt variations
- **Phase 6 (CLI):** cobra framework has excellent documentation and examples

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Core Go stack constrained by project requirements; versions need official source verification but choices are sound |
| Features | MEDIUM | Based on training data knowledge of JSON Resume ecosystem; should verify resumed current feature set and theme count |
| Architecture | HIGH | Command-Executor-Service pattern is established for Go CLI tools that orchestrate subprocesses; supported by kubectl, gh, docker CLI examples |
| Pitfalls | HIGH | Subprocess pitfalls (#1, #3, #4) are well-documented Go gotchas; LLM parsing pitfall (#2) is based on Anthropic Claude API patterns |

**Overall confidence:** MEDIUM-HIGH

Stack and architecture recommendations are HIGH confidence (established patterns, project constraints). Feature categorization is MEDIUM (based on training data, not live market research). Pitfall avoidance is HIGH (well-documented issues in the domain).

### Gaps to Address

**Version verification needed:**
- Go 1.22 release status (is it released, or should we target 1.21?)
- cobra version (is v1.8 current?)
- resumed version (is 4.0+ current, or has it evolved?)
- santhosh-tekuri/jsonschema version (does v6 exist, or is v5 latest?)

**Domain knowledge gaps:**
- Exact JSON Resume schema structure (need schema file for validation)
- Markdown CV format conventions (need documented example or specification)
- Current ATS optimization best practices (keyword density targets, format requirements)
- Resumed theme installation and configuration API

**How to handle:**
- Phase 1: Verify Go, cobra, jsonschema versions against official sources (golang.org, GitHub releases, npm registry)
- Phase 5: Research JSON Resume schema and markdown CV format before implementation
- Phase 4: Consult ATS optimization resources for prompt engineering guidance

**Validation during implementation:**
- Custom markdown parser (MEDIUM confidence decision): Comprehensive test suite with sample CVs, edge cases, malformed input
- Claude CLI flags and options: Verify with `claude --help` before executor implementation
- Resumed API: Check npm package documentation and examples before PDF export integration

## Sources

### Primary (HIGH confidence)
- **Go stdlib documentation** (os/exec, embed, encoding/json, text/template): Official Go documentation for subprocess management, embedded assets, JSON handling
- **cobra documentation** (github.com/spf13/cobra): Official CLI framework documentation and examples
- **Standard Go project layout** (github.com/golang-standards/project-layout): Community-maintained Go project structure conventions
- **Go testing patterns**: Standard library testing package documentation and interface-based mocking patterns
- **Unix subprocess behavior**: Buffer deadlocks, PATH resolution, signal handling are well-documented Unix/Linux patterns

### Secondary (MEDIUM confidence)
- **JSON Resume ecosystem** (jsonresume.org): Training data knowledge of schema, themes, tools (resume-cli, resumed)
- **npm/npx behavior**: npm documentation on package installation and npx execution in non-interactive mode
- **Node version manager PATH modifications**: nvm, volta, asdf documentation on shell integration
- **CLI resume tool landscape**: Training data knowledge of HackMyResume, FluentCV, resume-cli feature sets
- **Anthropic Claude API patterns**: LLM output parsing challenges based on API behavior as of training cutoff (January 2025)

### Tertiary (LOW confidence - needs validation)
- **Resumed package version and features**: Training data suggests it's active fork of resume-cli; should verify current npm registry state
- **JSON Resume theme count**: ~400 themes approximate based on npm search patterns; should verify current ecosystem size
- **Claude CLI current flags**: Should verify with `claude --help` before implementation
- **ATS optimization strategies**: Training data knowledge; should validate with current best practices

---
*Research completed: 2026-02-03*
*Ready for roadmap: yes*
