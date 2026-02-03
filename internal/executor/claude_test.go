package executor

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestClaudeExecutor_Execute tests basic execution with stdin piping
func TestClaudeExecutor_Execute(t *testing.T) {
	// Create a fake claude executable that echoes stdin
	tmpDir := t.TempDir()
	fakeClaude := filepath.Join(tmpDir, "claude")

	// Script that reads stdin and outputs it
	script := `#!/bin/sh
# Read stdin and echo it back
cat
`
	err := os.WriteFile(fakeClaude, []byte(script), 0755)
	if err != nil {
		t.Fatalf("failed to create fake claude: %v", err)
	}

	executor := NewClaudeExecutor(WithClaudePath(fakeClaude))
	ctx := context.Background()

	result, err := executor.Execute(ctx, "test prompt content")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	if strings.TrimSpace(result) != "test prompt content" {
		t.Errorf("expected 'test prompt content', got: %q", result)
	}
}

// TestClaudeExecutor_UsesStdinNotArgs verifies prompts go via stdin, not args
func TestClaudeExecutor_UsesStdinNotArgs(t *testing.T) {
	tmpDir := t.TempDir()
	fakeClaude := filepath.Join(tmpDir, "claude")

	// Script that outputs the number of arguments
	script := `#!/bin/sh
# Output number of positional arguments (excluding the script name)
echo "args: $#"
echo "stdin: $(cat)"
`
	err := os.WriteFile(fakeClaude, []byte(script), 0755)
	if err != nil {
		t.Fatalf("failed to create fake claude: %v", err)
	}

	executor := NewClaudeExecutor(WithClaudePath(fakeClaude))
	ctx := context.Background()

	// Use a large prompt that would fail with argument limits
	largePrompt := strings.Repeat("x", 10000)
	result, err := executor.Execute(ctx, largePrompt)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	// Args should be minimal (just flags), prompt should be in stdin
	if !strings.Contains(result, "stdin: "+largePrompt) {
		t.Errorf("prompt should be passed via stdin, got: %q", result)
	}
}

// TestClaudeExecutor_WithModel verifies -m flag is added
func TestClaudeExecutor_WithModel(t *testing.T) {
	tmpDir := t.TempDir()
	fakeClaude := filepath.Join(tmpDir, "claude")

	// Script that outputs all arguments
	script := `#!/bin/sh
echo "args: $@"
cat > /dev/null
`
	err := os.WriteFile(fakeClaude, []byte(script), 0755)
	if err != nil {
		t.Fatalf("failed to create fake claude: %v", err)
	}

	executor := NewClaudeExecutor(WithClaudePath(fakeClaude))
	ctx := context.Background()

	result, err := executor.Execute(ctx, "prompt", WithModel("sonnet"))
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	if !strings.Contains(result, "-m sonnet") && !strings.Contains(result, "--model sonnet") {
		t.Errorf("expected model flag in args, got: %q", result)
	}
}

// TestClaudeExecutor_WithOutputFormat verifies --output-format flag
func TestClaudeExecutor_WithOutputFormat(t *testing.T) {
	tmpDir := t.TempDir()
	fakeClaude := filepath.Join(tmpDir, "claude")

	script := `#!/bin/sh
echo "args: $@"
cat > /dev/null
`
	err := os.WriteFile(fakeClaude, []byte(script), 0755)
	if err != nil {
		t.Fatalf("failed to create fake claude: %v", err)
	}

	executor := NewClaudeExecutor(WithClaudePath(fakeClaude))
	ctx := context.Background()

	result, err := executor.Execute(ctx, "prompt", WithOutputFormat("json"))
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	if !strings.Contains(result, "--output-format json") {
		t.Errorf("expected output format flag, got: %q", result)
	}
}

// TestClaudeExecutor_StderrInError verifies stderr is included in errors
func TestClaudeExecutor_StderrInError(t *testing.T) {
	tmpDir := t.TempDir()
	fakeClaude := filepath.Join(tmpDir, "claude")

	// Script that outputs to stderr and exits with error
	script := `#!/bin/sh
echo "error details here" >&2
exit 1
`
	err := os.WriteFile(fakeClaude, []byte(script), 0755)
	if err != nil {
		t.Fatalf("failed to create fake claude: %v", err)
	}

	executor := NewClaudeExecutor(WithClaudePath(fakeClaude))
	ctx := context.Background()

	_, err = executor.Execute(ctx, "prompt")
	if err == nil {
		t.Fatal("expected error when process fails")
	}

	if !strings.Contains(err.Error(), "error details here") {
		t.Errorf("error should contain stderr, got: %v", err)
	}
}

// TestClaudeExecutor_RespectsContextCancellation verifies context timeout
func TestClaudeExecutor_RespectsContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()
	fakeClaude := filepath.Join(tmpDir, "claude")

	// Script that loops forever (more portable than sleep)
	script := `#!/bin/sh
while true; do
    : # no-op
done
`
	err := os.WriteFile(fakeClaude, []byte(script), 0755)
	if err != nil {
		t.Fatalf("failed to create fake claude: %v", err)
	}

	executor := NewClaudeExecutor(WithClaudePath(fakeClaude))

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err = executor.Execute(ctx, "prompt")
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected error when context is cancelled")
	}

	// Should return relatively quickly after context cancellation
	// Allow up to 5 seconds for process cleanup
	if elapsed > 5*time.Second {
		t.Errorf("context cancellation took too long: %v", elapsed)
	}
}

// TestClaudeExecutor_UsesBytesBuffer verifies bytes.Buffer pattern (not cmd.Output)
// This test ensures the streaming pattern is used by checking the executor
// can handle concurrent stdout/stderr without deadlock
func TestClaudeExecutor_UsesBytesBuffer(t *testing.T) {
	tmpDir := t.TempDir()
	fakeClaude := filepath.Join(tmpDir, "claude")

	// Script that outputs large amounts to both stdout and stderr simultaneously
	// This would deadlock with cmd.Output() if buffer fills
	script := `#!/bin/sh
# Generate large output to both streams
for i in $(seq 1 1000); do
    echo "stdout line $i"
    echo "stderr line $i" >&2
done
`
	err := os.WriteFile(fakeClaude, []byte(script), 0755)
	if err != nil {
		t.Fatalf("failed to create fake claude: %v", err)
	}

	executor := NewClaudeExecutor(WithClaudePath(fakeClaude))
	ctx := context.Background()

	result, err := executor.Execute(ctx, "prompt")
	if err != nil {
		t.Errorf("expected no error with large output, got: %v", err)
	}

	// Should have captured all stdout lines
	if !strings.Contains(result, "stdout line 1000") {
		t.Errorf("should have all stdout lines, got length: %d", len(result))
	}
}

// TestClaudeExecutor_DefaultFlags verifies -p and --output-format text are used by default
func TestClaudeExecutor_DefaultFlags(t *testing.T) {
	tmpDir := t.TempDir()
	fakeClaude := filepath.Join(tmpDir, "claude")

	script := `#!/bin/sh
echo "args: $@"
cat > /dev/null
`
	err := os.WriteFile(fakeClaude, []byte(script), 0755)
	if err != nil {
		t.Fatalf("failed to create fake claude: %v", err)
	}

	executor := NewClaudeExecutor(WithClaudePath(fakeClaude))
	ctx := context.Background()

	result, err := executor.Execute(ctx, "prompt")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	// Should have -p flag for print mode
	if !strings.Contains(result, "-p") {
		t.Errorf("expected -p flag, got: %q", result)
	}

	// Should have --output-format text by default
	if !strings.Contains(result, "--output-format text") {
		t.Errorf("expected --output-format text, got: %q", result)
	}
}

// TestClaudeExecutor_NotFound verifies error when claude binary not found
func TestClaudeExecutor_NotFound(t *testing.T) {
	executor := NewClaudeExecutor(WithClaudePath("/nonexistent/path/to/claude"))
	ctx := context.Background()

	_, err := executor.Execute(ctx, "prompt")
	if err == nil {
		t.Fatal("expected error when claude not found")
	}

	if !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), "no such file") {
		t.Errorf("error should indicate claude not found, got: %v", err)
	}
}

// verifyNotUsingCmdOutput is a compile-time check helper
// The actual verification is in code review - this test just ensures
// the implementation doesn't accidentally use cmd.Output()
func TestClaudeExecutor_ImplementationPattern(t *testing.T) {
	// This is more of a documentation test - the actual pattern
	// verification is done by code review and TestClaudeExecutor_UsesBytesBuffer
	// which would deadlock if cmd.Output() was used

	// Read the source file and verify pattern
	// Note: go test runs from the package directory
	source, err := os.ReadFile("claude.go")
	if err != nil {
		// Try with absolute path fallback
		source, err = os.ReadFile("/workspace/internal/executor/claude.go")
		if err != nil {
			t.Skip("claude.go not found: " + err.Error())
		}
	}

	// Check for forbidden patterns - look for actual method calls not comments
	sourceStr := string(source)

	// Use regex-like approach: look for = cmd.Output() or , cmd.Output()
	if strings.Contains(sourceStr, "= cmd.Output()") ||
		strings.Contains(sourceStr, ",cmd.Output()") ||
		strings.Contains(sourceStr, ", cmd.Output()") {
		t.Error("claude.go should not use cmd.Output() - use bytes.Buffer with cmd.Stdout/Stderr")
	}

	if strings.Contains(sourceStr, "= cmd.CombinedOutput()") ||
		strings.Contains(sourceStr, ",cmd.CombinedOutput()") ||
		strings.Contains(sourceStr, ", cmd.CombinedOutput()") {
		t.Error("claude.go should not use cmd.CombinedOutput() - use bytes.Buffer with cmd.Stdout/Stderr")
	}

	// Verify correct patterns are present
	if !strings.Contains(sourceStr, "bytes.Buffer") {
		t.Error("claude.go should use bytes.Buffer for output capture")
	}

	if !strings.Contains(sourceStr, "cmd.Start()") {
		t.Error("claude.go should use cmd.Start() + cmd.Wait() pattern")
	}

	if !strings.Contains(sourceStr, "cmd.Wait()") {
		t.Error("claude.go should use cmd.Start() + cmd.Wait() pattern")
	}
}

// Ensure exec package is used correctly (compile check)
var _ = exec.Command
