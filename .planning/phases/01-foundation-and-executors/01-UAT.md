---
status: diagnosed
phase: 01-foundation-and-executors
source: [01-01-SUMMARY.md, 01-02-SUMMARY.md, 01-03-SUMMARY.md]
started: 2026-02-03T05:45:00Z
updated: 2026-02-03T05:50:00Z
---

## Current Test

[testing complete]

## Tests

### 1. Binary builds successfully
expected: Run `go build -o m2cv .` — should complete without errors and produce executable binary
result: pass

### 2. Version command works
expected: Run `./m2cv version` — should print version info (dev/unknown/unknown) without preflight errors
result: issue
reported: "there is no version information"
severity: minor

### 3. Help shows flags
expected: Run `./m2cv --help` — should show usage with --config and --base-cv persistent flags listed
result: pass

### 4. Config tests pass
expected: Run `go test ./internal/config/ -v` — all 9 tests should pass
result: pass

### 5. Executor tests pass
expected: Run `go test ./internal/executor/ -v` — all 23 tests should pass
result: issue
reported: "2 tests fail: TestFindNodeExecutable_NotFound and TestNPMExecutor_NotFound - expected error when executable not found"
severity: major

### 6. Preflight tests pass
expected: Run `go test ./internal/preflight/ -v` — all 5 tests should pass
result: pass

## Summary

total: 6
passed: 4
issues: 2
pending: 0
skipped: 0

## Gaps

- truth: "Version command prints version info"
  status: failed
  reason: "User reported: there is no version information"
  severity: minor
  test: 2
  root_cause: "Build process does not pass ldflags to go build, causing version variables to retain hardcoded placeholder values"
  artifacts:
    - path: "cmd/root.go"
      issue: "Package variables version/commit/date have hardcoded defaults"
    - path: "cmd/version.go"
      issue: "Displays placeholders correctly but no build tooling sets real values"
  missing:
    - "Create Makefile or build script that passes -ldflags with version info"
  debug_session: ".planning/debug/version-placeholder-values.md"

- truth: "All executor tests pass"
  status: failed
  reason: "User reported: 2 tests fail: TestFindNodeExecutable_NotFound and TestNPMExecutor_NotFound - expected error when executable not found"
  severity: major
  test: 5
  root_cause: "Tests set PATH='' but FindNodeExecutable still finds executables via hardcoded system paths (/usr/local/bin, /opt/homebrew/bin)"
  artifacts:
    - path: "internal/executor/find.go"
      issue: "Lines 50-53 check system paths unconditionally"
    - path: "internal/executor/find_test.go"
      issue: "TestFindNodeExecutable_NotFound doesn't mock system paths"
    - path: "internal/executor/npm_test.go"
      issue: "TestNPMExecutor_NotFound has same isolation problem"
  missing:
    - "Refactor FindNodeExecutable to accept optional fallback paths for testing"
    - "Or use test isolation that prevents os.Stat from finding system executables"
  debug_session: ""
