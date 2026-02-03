---
status: resolved
trigger: "version command shows placeholder values (dev, unknown, unknown)"
created: 2026-02-03T00:00:00Z
updated: 2026-02-03T00:00:00Z
symptoms_prefilled: true
goal: find_root_cause_only
---

## Current Focus

hypothesis: ldflags not passed during build, so version variables retain default placeholder values
test: examine how version variables are initialized and whether ldflags are configured
expecting: will find either unset ldflags configuration or default placeholder values in code
next_action: read cmd/version.go and cmd/root.go to understand version initialization

## Symptoms

expected: version command prints real version info (e.g., semantic version, commit hash, build timestamp)
actual: version command prints "dev", "unknown", "unknown"
errors: none - command runs successfully but shows placeholders
reproduction: run ./m2cv version
started: unclear if this is always the case or recent regression

## Eliminated

- hypothesis: there is a bug in the version command display logic
  evidence: code is correct - fmt.Printf correctly references package vars. The bug is not in display.
  timestamp: 2026-02-03T00:02:00Z

## Evidence

- timestamp: 2026-02-03T00:01:00Z
  checked: cmd/root.go package-level variables
  found: three package vars initialized with placeholders (version="dev", commit="unknown", date="unknown")
  implication: these are the defaults when ldflags are not provided during build

- timestamp: 2026-02-03T00:02:00Z
  checked: cmd/version.go implementation
  found: version command prints from package vars (fmt.Printf uses version, commit, date)
  implication: command correctly references the variables, no bug in display logic

- timestamp: 2026-02-03T00:03:00Z
  checked: build infrastructure (Makefile, .github/workflows, .yml files)
  found: no build configuration files found in repository root or subdirectories
  implication: no ldflags are configured anywhere; build happens without version injection

- timestamp: 2026-02-03T00:04:00Z
  checked: planning docs (.planning/phases/01-foundation-and-executors/01-03-PLAN.md)
  found: plan documents "define package-level version variables (set via ldflags at build time)" but no actual build command or CI/CD configuration implements it
  implication: feature was planned but build tooling was never implemented

## Resolution

root_cause: Build process does not pass ldflags to `go build` command, so version variables retain their hardcoded placeholder values (version="dev", commit="unknown", date="unknown")
fix: N/A (root cause identified, no fix needed yet - awaiting decision on build tooling)
verification: N/A
files_changed: []
