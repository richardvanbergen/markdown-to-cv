# m2cv — Markdown to CV

A Go CLI tool that uses [Claude](https://docs.anthropic.com/en/docs/claude-code) to generate tailored CVs from markdown and convert them to [JSON Resume](https://jsonresume.org) format for PDF export.

## Prerequisites

- [Go](https://go.dev/) 1.22+
- [Claude CLI](https://docs.anthropic.com/en/docs/claude-code) (`claude` in PATH)
- [resume-cli](https://github.com/jsonresume/resume-cli) (optional, for PDF export)

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

## Usage

### Generate a tailored CV

Takes a job description and your base CV, produces a tailored markdown CV:

```bash
m2cv generate job-description.txt base-cv.md -o tailored-cv.md
```

Edit the output in your favourite editor, then validate.

### Validate and convert to JSON Resume

Converts your markdown CV to JSON Resume format with schema validation:

```bash
m2cv validate tailored-cv.md -o resume.json
```

Export to PDF (requires `resume-cli`):

```bash
m2cv validate tailored-cv.md -o resume.json --pdf
```

### Flags

| Flag | Commands | Description |
|------|----------|-------------|
| `-o` | generate, validate | Output file path (default: `output.md` / `resume.json`) |
| `-m` | generate, validate | Claude model to use |
| `--pdf` | validate | Also export PDF via resume-cli |

## Markdown CV Format

The markdown format uses YAML frontmatter for contact details and headings for sections. This maps directly to the JSON Resume schema.

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
2-3 sentence professional summary.

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

1. Write or gather your base CV in the markdown format above
2. `m2cv generate job-description.txt base-cv.md -o tailored.md` — Claude tailors it
3. Edit `tailored.md` to your liking
4. `m2cv validate tailored.md -o resume.json --pdf` — validate, convert, export PDF
5. Repeat as needed

## License

MIT
