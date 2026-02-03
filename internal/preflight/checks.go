// Package preflight provides pre-execution checks for external dependencies.
package preflight

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/richq/m2cv/internal/executor"
)

// CheckClaude verifies that the Claude CLI is available in PATH.
// Returns an error with installation instructions if not found.
func CheckClaude() error {
	if _, err := exec.LookPath("claude"); err != nil {
		return fmt.Errorf(`claude CLI not found in PATH

Install from: https://claude.ai/download
Requires: Claude Pro subscription

After installing, verify with: claude --version`)
	}
	return nil
}

// CheckNPM verifies that npm is available using the Node.js executable finder.
// Returns an error with installation instructions if not found.
func CheckNPM() error {
	_, err := executor.FindNodeExecutable("npm")
	return err
}

// CheckResumed verifies that the 'resumed' tool is available.
// It checks two locations:
//  1. Local project install: <projectDir>/node_modules/resumed
//  2. Global install: 'resumed' in PATH
//
// Returns an error with installation instructions if not found.
func CheckResumed(projectDir string) error {
	// Check local node_modules first
	localPath := filepath.Join(projectDir, "node_modules", "resumed")
	if info, err := os.Stat(localPath); err == nil && info.IsDir() {
		return nil
	}

	// Check global install via PATH
	if _, err := exec.LookPath("resumed"); err == nil {
		return nil
	}

	return fmt.Errorf(`resumed not found in node_modules or PATH

Install with one of:
  Local:  npm install resumed (in your project directory)
  Global: npm install -g resumed

Or run: m2cv init (to set up a new application)`)
}
