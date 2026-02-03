package preflight

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckResumed_FindsInNodeModules(t *testing.T) {
	// Create a temp directory with node_modules/resumed
	tmpDir := t.TempDir()
	resumedPath := filepath.Join(tmpDir, "node_modules", "resumed")
	if err := os.MkdirAll(resumedPath, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	err := CheckResumed(tmpDir)
	if err != nil {
		t.Errorf("CheckResumed() = %v, want nil (should find in node_modules)", err)
	}
}

func TestCheckResumed_ReturnsErrorWhenNotFound(t *testing.T) {
	// Use a temp directory without node_modules
	tmpDir := t.TempDir()

	// Temporarily clear PATH to ensure resumed can't be found globally
	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", "")
	defer os.Setenv("PATH", oldPath)

	err := CheckResumed(tmpDir)
	if err == nil {
		t.Error("CheckResumed() = nil, want error when resumed not found")
	}
}

func TestCheckResumed_ErrorContainsInstallInstructions(t *testing.T) {
	tmpDir := t.TempDir()

	// Clear PATH to ensure resumed can't be found globally
	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", "")
	defer os.Setenv("PATH", oldPath)

	err := CheckResumed(tmpDir)
	if err == nil {
		t.Fatal("CheckResumed() = nil, want error with install instructions")
	}

	errMsg := err.Error()

	// Check for actionable install instructions
	expectedSubstrings := []string{
		"resumed not found",
		"npm install",
	}

	for _, expected := range expectedSubstrings {
		if !strings.Contains(errMsg, expected) {
			t.Errorf("error message missing %q\ngot: %s", expected, errMsg)
		}
	}
}

func TestCheckClaude_ErrorContainsInstallInstructions(t *testing.T) {
	// Clear PATH to ensure claude can't be found
	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", "")
	defer os.Setenv("PATH", oldPath)

	err := CheckClaude()
	if err == nil {
		t.Fatal("CheckClaude() = nil, want error when claude not in PATH")
	}

	errMsg := err.Error()

	expectedSubstrings := []string{
		"claude CLI not found",
		"https://claude.ai/download",
		"claude --version",
	}

	for _, expected := range expectedSubstrings {
		if !strings.Contains(errMsg, expected) {
			t.Errorf("error message missing %q\ngot: %s", expected, errMsg)
		}
	}
}

func TestCheckNPM_ErrorContainsInstallInstructions(t *testing.T) {
	// Clear PATH to ensure npm can't be found via PATH
	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", "")
	defer os.Setenv("PATH", oldPath)

	err := CheckNPM()
	if err == nil {
		// npm was found via fallback paths (e.g., /usr/local/bin, ~/.nvm, etc.)
		// This is expected behavior - FindNodeExecutable has hardcoded fallbacks.
		// Skip the test since we can't reliably test the error case on this system.
		t.Skip("npm found via fallback paths, cannot test error case")
	}

	errMsg := err.Error()

	// Should mention npm and installation options (from executor.FindNodeExecutable)
	expectedSubstrings := []string{
		"npm not found",
		"Node.js",
	}

	for _, expected := range expectedSubstrings {
		if !strings.Contains(errMsg, expected) {
			t.Errorf("error message missing %q\ngot: %s", expected, errMsg)
		}
	}
}
