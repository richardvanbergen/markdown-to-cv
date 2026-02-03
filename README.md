# m2cv — Markdown to CV

A Go CLI tool that manages job applications with AI-powered resume tailoring. It uses [Claude CLI](https://docs.anthropic.com/en/docs/claude-code) to tailor your base CV for specific roles, converts to [JSON Resume](https://jsonresume.org) format, and exports professional themed PDFs via [resumed](https://github.com/rbardini/resumed).

## Prerequisites

- [Go](https://go.dev/) 1.22+
- [Claude CLI](https://docs.anthropic.com/en/docs/claude-code) (`claude` in PATH)
- [Node.js](https://nodejs.org/) with npm (for `resumed` and themes)

## Install

```bash
go install github.com/richardvanbergen/markdown-to-cv@latest
```

Or build from source:

```bash
git clone https://github.com/richardvanbergen/markdown-to-cv.git
cd markdown-to-cv
go build -o m2cv .
```

## Quick Start

```bash
# 1. Initialize a project with config and theme selection
m2cv init

# 2. Create an application folder from a job description
m2cv apply job-posting.txt

# 3. Tailor your CV for that specific role
m2cv optimize acme-software-engineer

# 4. Generate a professional PDF
m2cv generate acme-software-engineer
```

## Commands

### `m2cv init`

Initialize a new m2cv project in the current directory. Creates `m2cv.yml` config, installs `resumed` via npm, and lets you select a JSON Resume theme.

```bash
# Interactive theme selection
m2cv init

# Non-interactive with specific theme
m2cv init --theme even --base-cv ~/cv/base.md

# Overwrite existing config
m2cv init --force
```

**Flags:**
- `--theme`, `-t` — Specify theme (skips interactive selection)
- `--base-cv` — Path to your base CV markdown file
- `--force`, `-f` — Overwrite existing configuration

### `m2cv apply`

Create a job application folder with an AI-extracted name from the job description. The folder is created under `applications/` and the job description is copied into it.

```bash
# Auto-name via Claude
m2cv apply job-posting.txt

# Manual folder name
m2cv apply --name acme-corp-engineer job.txt

# Custom applications directory
m2cv apply --dir my-apps job.txt
```

**Flags:**
- `--name`, `-n` — Override folder name (skip Claude extraction)
- `--dir`, `-d` — Applications directory (default: `applications`)

### `m2cv optimize`

Tailor your base CV to a specific job description using Claude AI. Produces versioned output (`optimized-cv-1.md`, `optimized-cv-2.md`, etc.) in the application folder.

```bash
# Standard optimization
m2cv optimize acme-software-engineer

# ATS-friendly optimization
m2cv optimize --ats google-sre

# Override Claude model
m2cv optimize -m claude-sonnet-4-20250514 my-dream-job
```

**Flags:**
- `--model`, `-m` — Override Claude model
- `--ats` — Optimize for ATS (Applicant Tracking Systems)

### `m2cv generate`

Convert the latest optimized CV to JSON Resume format and export a themed PDF via `resumed`. Validates against JSON Resume schema before export.

```bash
# Use default theme from config
m2cv generate acme-software-engineer

# Override theme
m2cv generate --theme stackoverflow my-app

# Override Claude model for JSON conversion
m2cv generate -m claude-sonnet-4-20250514 my-dream-job
```

**Flags:**
- `--theme` — Override JSON Resume theme
- `--model`, `-m` — Override Claude model

**Output files** (written to application folder):
- `resume.json` — JSON Resume format (useful for debugging)
- `resume.pdf` — Final PDF output

### Global Flags

Available for all commands:

- `--config` — Config file path (default: searches for `m2cv.yml` in current/parent directories)
- `--base-cv` — Override base CV path from config

## Configuration

The `m2cv.yml` file stores project configuration:

```yaml
base_cv_path: base-cv.md
default_theme: even
themes:
  - even
  - stackoverflow
default_model: claude-sonnet-4-20250514
```

**Config discovery** (in order):
1. `--config` flag
2. `M2CV_CONFIG` environment variable
3. Walk up directory tree looking for `m2cv.yml`

## Markdown CV Format

Your base CV uses YAML frontmatter for contact details and headings for sections. This maps directly to the JSON Resume schema.

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
2-3 sentence professional summary highlighting key achievements and expertise.

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

### Section conventions

| Section | Heading | Entry format |
|---------|---------|-------------|
| Summary | `# Summary` | Paragraph text |
| Experience | `# Experience` | `## Title \| Company` with `*dates*`, `> department`, bullet highlights |
| Education | `# Education` | `## Degree \| Institution` with `*dates*`, bullet courses |
| Skills | `# Skills` | `## Category` with bullet items |
| Projects | `# Projects` | `## Name` with `*dates*`, `> description`, bullets |
| Languages | `# Languages` | `- Language: Level` |
| Certificates | `# Certificates` | `- Name \| Issuer \| Date` |

## Workflow

```
my-job-search/
├── m2cv.yml
├── base-cv.md
├── node_modules/
└── applications/
    ├── acme-software-engineer/
    │   ├── job-posting.txt
    │   ├── optimized-cv-1.md
    │   ├── optimized-cv-2.md
    │   ├── resume.json
    │   └── resume.pdf
    └── google-sre/
        ├── job-description.txt
        └── optimized-cv-1.md
```

**Typical workflow:**

1. Write your base CV in markdown format (once)
2. Initialize project: `m2cv init`
3. For each job application:
   - Create application: `m2cv apply job-posting.txt`
   - Tailor CV: `m2cv optimize <app-name>`
   - Review and edit the optimized CV in your editor
   - Re-optimize if needed (creates new version)
   - Generate PDF: `m2cv generate <app-name>`

## Available Themes

m2cv uses the [JSON Resume theme ecosystem](https://jsonresume.org/themes). Popular themes include:

- `even` — Clean two-column layout
- `stackoverflow` — Stack Overflow careers style
- `flat` — Minimalist flat design
- `jsonresume-theme-caffeine` — Modern with accent colors
- `jsonresume-theme-actual` — Professional single-column

During `m2cv init`, you'll be prompted to select from available themes. Additional themes can be installed manually via npm.

## License

MIT
