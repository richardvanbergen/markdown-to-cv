# Plan: m2cv — Markdown to CV CLI Tool

## Overview

A Go CLI tool with two commands that uses `claude -p` to generate tailored CVs from markdown and convert them to JSON Resume format for PDF export via `resume-cli`.

**Workflow:** Job description + base CV → Claude generates markdown → user edits → Claude converts to JSON Resume → validate → PDF

## File Structure

```
markdown-to-cv/
├── go.mod
├── main.go
├── cmd/
│   ├── root.go              # Root cobra command + preflight checks
│   ├── generate.go           # generate subcommand
│   └── validate.go           # validate subcommand
├── internal/
│   ├── assets/
│   │   ├── assets.go         # Embedded prompts + schema
│   │   ├── prompts/
│   │   │   ├── generate.txt  # Prompt template for generate
│   │   │   └── validate.txt  # Prompt template for validate
│   │   └── schema/
│   │       └── resume.schema.json
│   ├── claude/
│   │   └── claude.go         # Wrapper for exec'ing claude -p (stdin pipe)
│   └── preflight/
│       └── preflight.go      # Check claude + resume-cli in PATH
└── examples/
    └── sample-cv.md          # Example CV in the markdown format
```

## Markdown CV Format Spec

YAML frontmatter for `basics`, then strict heading conventions:

```markdown
---
name: Jane Doe
label: Senior Software Engineer
email: jane@example.com
phone: "+1-555-123-4567"
url: https://janedoe.dev
location:
  city: Amsterdam
  countryCode: NL
profiles:
  - network: GitHub
    username: janedoe
    url: https://github.com/janedoe
---

# Summary
2-3 sentence professional summary...

# Experience
## Senior Developer | Acme Corp
*2021-01 - present*
> Cloud infrastructure team
- Led migration of monolith to microservices
- Reduced deployment time by 70%

# Education
## MSc Computer Science | University of Amsterdam
*2016-09 - 2018-06*
- Distributed Systems
- Machine Learning

# Skills
## Backend Development
- Go
- Python
- Node.js

# Projects
## Open Source CLI Tool
*2022-01 - present*
> Personal project
- Built a CLI tool with 500+ GitHub stars

# Languages
- English: Native
- Dutch: Professional

# Certificates
- AWS SA Associate | Amazon Web Services | 2023-03
```

## Commands

### m2cv generate \<job-description\> \<base-cv\> [-o output.md] [-m model]

1. Read both input files
2. Render `generate.txt` template with file contents injected
3. Pipe prompt to `claude -p --output-format text` via stdin
4. Write Claude's output to `-o` path (default: `output.md`)

### m2cv validate \<markdown-file\> [-o resume.json] [--pdf] [-m model]

1. Read markdown file
2. Render `validate.txt` template with markdown + embedded JSON Resume schema
3. Pipe to `claude -p --output-format text` via stdin
4. Strip any markdown code fences from output
5. Validate JSON syntax (`json.Unmarshal`)
6. Validate against JSON Resume schema using `santhosh-tekuri/jsonschema/v6`
7. Pretty-print and write to `-o` path (default: `resume.json`)
8. If `--pdf` flag: check `resume-cli` exists, run `resume-cli export resume.pdf --resume resume.json`

## Preflight Checks

- `PersistentPreRunE` on root command checks `claude` is in PATH (hard error)
- `validate` command checks `resume-cli` if `--pdf` is used (hard error only then, otherwise warn)

## Claude Interaction

Prompts piped via stdin to avoid shell argument limits:

```go
cmd := exec.Command("claude", "-p", "--output-format", "text")
cmd.Stdin = strings.NewReader(prompt)
cmd.Stderr = os.Stderr
output, err := cmd.Output()
```

If `--model` flag is set, add `"-m", model` args.

## Dependencies

- `github.com/spf13/cobra` — CLI framework
- `github.com/santhosh-tekuri/jsonschema/v6` — JSON Schema validation

## Implementation Order

1. `go mod init` + install deps
2. Scaffold directory structure
3. Download JSON Resume schema into `internal/assets/schema/`
4. Write prompt templates
5. Implement `internal/assets/assets.go` (embed directives)
6. Implement `internal/preflight/preflight.go`
7. Implement `internal/claude/claude.go`
8. Implement `cmd/root.go`
9. Implement `cmd/generate.go`
10. Implement `cmd/validate.go`
11. Implement `main.go`
12. Write example CV file
13. Build and test end-to-end

## Verification

1. `go build -o m2cv .` — confirms compilation
2. `./m2cv --help` — shows subcommands
3. `./m2cv generate examples/job.txt examples/sample-cv.md -o test-output.md` — generates markdown CV
4. `./m2cv validate test-output.md -o test-resume.json` — converts + validates JSON
5. Inspect `test-resume.json` for correct JSON Resume structure
6. `./m2cv validate test-output.md --pdf` — generates PDF (requires resume-cli installed)
