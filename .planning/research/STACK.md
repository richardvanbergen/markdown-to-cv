# Technology Stack

**Project:** m2cv — Markdown to CV CLI Tool
**Researched:** 2026-02-03
**Confidence:** MEDIUM (based on training data and project constraints)

## Executive Summary

m2cv is a Go CLI tool that orchestrates Claude CLI subprocess calls and integrates with the npm-based JSON Resume ecosystem. The stack is constrained by specific requirements: Go language, cobra CLI framework, claude subprocess only, resumed for PDF export, YAML config, and santhosh-tekuri/jsonschema for validation.

**Key architectural decision:** Hybrid Go + npm ecosystem. Go handles CLI orchestration, file management, and validation; npm/Node.js handles PDF theming via resumed.

## Recommended Stack

### Core Language & Runtime

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| Go | 1.22+ | CLI implementation, core logic | Single binary distribution, excellent CLI tooling, stdlib covers most needs. 1.22+ for improved type inference and stdlib additions. |
| Node.js | 18+ LTS | Runtime for resumed PDF export | Required by resumed and JSON Resume themes. LTS ensures stability. User installs; not bundled. |

**Rationale for Go 1.22+:**
- Modern generics support (if needed for internal utilities)
- Improved `os/exec` handling for subprocess management
- Enhanced error wrapping with `errors.Join`
- Project constraint: Go language mandated

**Confidence:** HIGH (project constraint, versions verified from Go release schedule)

### CLI Framework

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| github.com/spf13/cobra | v1.8+ | CLI command structure, flags, help | Industry standard for Go CLIs (kubectl, gh, hugo use it). Mature, well-documented, supports subcommands, persistent flags, and hooks. |
| github.com/spf13/pflag | (cobra dep) | POSIX-style flag parsing | Pulled in by cobra; better flag ergonomics than stdlib |

**Rationale:**
- cobra is the de facto standard for Go CLIs
- PersistentPreRunE hooks perfect for preflight checks (claude/resumed availability)
- Automatic help generation and shell completion
- Project constraint: cobra explicitly required

**Confidence:** HIGH (project constraint, widely used)

### Configuration Management

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| gopkg.in/yaml.v3 | v3.0+ | YAML parsing for m2cv.yml | Most mature Go YAML library. v3 has better error messages and maintains insertion order. |

**Alternatives considered:**
- **github.com/spf13/viper** (full config framework): Overkill for this use case. Viper is great for complex configs with env var merging, remote config, etc., but m2cv.yml is simple and file-based only. yaml.v3 + custom struct unmarshaling is cleaner and more transparent.
- **github.com/goccy/go-yaml**: Faster but less battle-tested than yaml.v3.

**Rationale:**
- Simple YAML config file doesn't need viper's complexity
- yaml.v3 is stdlib-like in maturity and usage
- Direct struct unmarshaling keeps config handling explicit

**Confidence:** MEDIUM (yaml.v3 is standard, but viper choice is opinionated)

### JSON Schema Validation

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| github.com/santhosh-tekuri/jsonschema/v6 | v6.0+ | Validate JSON Resume against schema | Fast, supports draft-07 (used by JSON Resume schema), mature API, good error messages. |

**Rationale:**
- JSON Resume uses JSON Schema draft-07
- v6 is the current major version (v5 was draft-07 focused, v6 adds draft 2019-09 and 2020-12)
- Widely used in Go ecosystem for schema validation
- Project constraint: explicitly required

**Confidence:** HIGH (project constraint, current stable version)

### Process Execution

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| os/exec (stdlib) | (Go 1.22+) | Shell out to claude, resumed, npm | Stdlib is sufficient. No need for external deps like go-cmd/cmd. Go 1.22 improved context handling. |

**Rationale:**
- Shelling out to `claude -p` and `resumed export`
- Stdin piping for prompt injection (avoids shell arg limits)
- stderr passthrough for debugging
- Stdlib is well-tested and sufficient

**Confidence:** HIGH (stdlib, no alternatives needed)

### File Embedding

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| embed (stdlib) | (Go 1.22+) | Embed prompts, JSON Resume schema | Stdlib since Go 1.16. Zero dependencies, compiled into binary. |

**Usage:**
```go
//go:embed assets/prompts/*.txt assets/schema/*.json
var assets embed.FS
```

**Confidence:** HIGH (stdlib, perfect fit)

### Markdown Parsing

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| gopkg.in/yaml.v3 | v3.0+ | Parse YAML frontmatter | Same lib as config; handles `---` delimited frontmatter. |
| (custom parser) | — | Parse markdown sections | Markdown structure is simple (# headings, ## subheadings, bullets). Regex + string parsing sufficient. No need for full markdown AST. |

**Alternatives considered:**
- **github.com/yuin/goldmark**: Full-featured markdown parser. Overkill. We're not rendering HTML; we're extracting structured data from a constrained format.
- **github.com/gomarkdown/markdown**: Same reasoning.

**Rationale:**
- Markdown CV format is constrained and documented
- Frontmatter: extract with regex `(?s)^---\n(.*?)\n---`, parse with yaml.v3
- Sections: split on `# ` headings, parse subsections with simple text processing
- Avoids heavy dependencies for a simple parsing task

**Confidence:** MEDIUM (custom parsing is simpler, but requires careful implementation)

### npm Integration (PDF Export)

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| resumed | 4.0+ | PDF/HTML export from JSON Resume | Actively maintained fork of resume-cli. Supports all JSON Resume themes. Faster, modern, TypeScript. |
| jsonresume-theme-* | (user choice) | PDF theming | 400+ themes on npm. User selects at init time. |

**Why resumed over resume-cli:**
- resume-cli is unmaintained (last update 2017)
- resumed is the active community fork (2020+)
- Better theme compatibility
- Faster rendering
- Project constraint: resumed explicitly required

**npm workflow:**
- `m2cv init` runs `npm init -y` in project dir, installs resumed + selected theme
- `m2cv.yml` tracks installed themes
- `m2cv generate` calls `resumed export resume.pdf --theme <name> --resume resume.json`

**Confidence:** HIGH (resumed is community standard, project constraint)

### External Dependencies (User-Installed)

| Tool | Version | Purpose | Preflight Check |
|------|---------|---------|-----------------|
| claude | latest | AI prompt execution | Hard error if missing (PersistentPreRunE) |
| node/npm | 18+ LTS | Run resumed, install themes | Soft warn at init, hard error at generate |
| resumed | 4.0+ | PDF export | Hard error if --pdf flag used |

**Preflight strategy:**
- Check `claude` availability in root command's `PersistentPreRunE`
- Check `resumed` availability only when `generate` command runs
- Provide helpful error messages with install instructions

**Confidence:** HIGH (clear from project requirements)

## Development & Tooling

| Tool | Purpose | Notes |
|------|---------|-------|
| go mod | Dependency management | Standard Go tooling |
| go build | Binary compilation | Single binary output |
| go test | Unit testing | Stdlib testing package sufficient |
| go install | User installation | Standard distribution method |

## Installation Steps

### For Development

```bash
# Clone repo
git clone https://github.com/richardvanbergen/markdown-to-cv.git
cd markdown-to-cv

# Initialize Go module
go mod init github.com/richardvanbergen/markdown-to-cv

# Install Go dependencies
go get github.com/spf13/cobra@latest
go get gopkg.in/yaml.v3@latest
go get github.com/santhosh-tekuri/jsonschema/v6@latest

# Build
go build -o m2cv .
```

### For Users

```bash
# Install from source
go install github.com/richardvanbergen/markdown-to-cv@latest

# Or download binary from releases
```

## Dependency Graph

```
m2cv (Go binary)
├── github.com/spf13/cobra (CLI framework)
│   └── github.com/spf13/pflag (flag parsing)
├── gopkg.in/yaml.v3 (YAML parsing)
├── github.com/santhosh-tekuri/jsonschema/v6 (schema validation)
└── (stdlib: os/exec, embed, encoding/json, etc.)

External processes (shelled out):
├── claude (subprocess via os/exec)
└── npm ecosystem (separate runtime)
    ├── node (JavaScript runtime)
    ├── resumed (PDF export)
    └── jsonresume-theme-* (themes)
```

## Alternatives Considered

| Category | Recommended | Alternative | Why Not |
|----------|-------------|-------------|---------|
| CLI framework | cobra | urfave/cli | cobra is more feature-rich, better docs, industry standard |
| Config parsing | yaml.v3 | spf13/viper | viper is overkill for simple file-based config |
| Markdown parsing | custom | goldmark/gomarkdown | Full parsers are overkill for constrained format |
| JSON validation | santhosh-tekuri | xeipuuv/gojsonschema | santhosh-tekuri has better draft-07 support, cleaner API |
| PDF export | resumed | resume-cli | resume-cli unmaintained; resumed is active fork |
| PDF export | resumed | custom Go PDF gen | resumed leverages 400+ existing themes; custom would be months of work |

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| resumed breaks/unmaintained | Low | High | PDF export is isolated; could swap to fork or alternative |
| claude CLI changes interface | Medium | High | Version checking in preflight; document supported versions |
| JSON Resume schema changes | Low | Medium | Schema is embedded; update on new releases |
| Node.js version incompatibility | Low | Low | Target LTS versions (18+); document requirements |
| Custom markdown parser bugs | Medium | Medium | Comprehensive test suite; validate against sample CVs |

## Version Pinning Strategy

**Go dependencies (go.mod):**
- Use `@latest` during development
- Pin to specific minor versions before v1.0 release
- Example: `github.com/spf13/cobra v1.8.0`

**npm dependencies (package.json managed by user):**
- User controls their own npm environment
- Document recommended versions in README
- resumed 4.0+ required; check version in preflight if possible

## Sources & Confidence Notes

| Technology | Source | Confidence |
|------------|--------|------------|
| Go 1.22+ | Training data (Go release schedule) | HIGH |
| cobra | Training data (widely used) | HIGH |
| yaml.v3 | Training data (Go YAML standard) | HIGH |
| jsonschema/v6 | Training data + project constraint | HIGH |
| resumed | Project context + training data | HIGH |
| Custom markdown parsing | Architectural decision based on requirements | MEDIUM |

**Overall confidence:** MEDIUM
- Core Go stack is HIGH confidence (constrained by project requirements)
- npm ecosystem (resumed, themes) is HIGH confidence (verified in project context)
- Custom markdown parsing is MEDIUM confidence (implementation risk, but architecturally sound)
- Specific versions need verification against official sources (Context7 or package repos)

## Open Questions for Validation

1. **resumed version:** Is 4.0+ current? Check npm registry.
2. **jsonschema/v6:** Does v6 exist, or is v5 latest? Verify on GitHub.
3. **Go 1.22 release status:** Is 1.22 released, or should we target 1.21? Check golang.org.
4. **cobra version:** Is v1.8 current? Check GitHub releases.

These should be verified with official sources (Context7, GitHub releases, package registries) before implementation begins.

---

**Next step:** Verify versions with authoritative sources, then proceed to roadmap creation.
