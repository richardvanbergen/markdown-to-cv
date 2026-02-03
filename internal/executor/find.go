// Package executor provides subprocess execution utilities for the m2cv CLI.
package executor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// FindOptions configures the behavior of FindNodeExecutableWithOptions.
type FindOptions struct {
	// SkipSystemPaths disables checking /usr/local/bin and /opt/homebrew/bin.
	// Useful for testing to ensure isolation from host system binaries.
	SkipSystemPaths bool
}

// FindNodeExecutable finds a Node.js ecosystem executable (npm, npx, node)
// by first checking exec.LookPath, then falling back to common version manager
// installation locations.
//
// Fallback locations checked in order:
//   - ~/.nvm/current/bin
//   - ~/.volta/bin
//   - ~/.asdf/shims
//   - ~/.fnm/current/bin
//   - /usr/local/bin
//   - /opt/homebrew/bin
//
// Returns the full path to the executable or an error with install instructions.
func FindNodeExecutable(name string) (string, error) {
	return FindNodeExecutableWithOptions(name, nil)
}

// FindNodeExecutableWithOptions finds a Node.js ecosystem executable with
// configurable behavior. See FindNodeExecutable for the default behavior.
//
// If opts is nil, uses default behavior (checks all locations including system paths).
// If opts.SkipSystemPaths is true, skips /usr/local/bin and /opt/homebrew/bin.
func FindNodeExecutableWithOptions(name string, opts *FindOptions) (string, error) {
	// Try exec.LookPath first (uses PATH)
	if path, err := exec.LookPath(name); err == nil {
		return path, nil
	}

	// Get home directory for fallback paths
	home, err := os.UserHomeDir()
	if err != nil {
		// Can't determine home, try system paths only
		home = ""
	}

	// Build list of fallback locations to check
	var candidates []string

	if home != "" {
		candidates = append(candidates,
			filepath.Join(home, ".nvm", "current", "bin", name),
			filepath.Join(home, ".volta", "bin", name),
			filepath.Join(home, ".asdf", "shims", name),
			filepath.Join(home, ".fnm", "current", "bin", name),
		)
	}

	// System-wide locations (skip if requested for test isolation)
	if opts == nil || !opts.SkipSystemPaths {
		candidates = append(candidates,
			filepath.Join("/usr/local/bin", name),
			filepath.Join("/opt/homebrew/bin", name),
		)
	}

	// Check each candidate path
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil {
			// Check if it's executable (not a directory)
			if !info.IsDir() && info.Mode()&0111 != 0 {
				return candidate, nil
			}
		}
	}

	// Not found - return descriptive error
	return "", fmt.Errorf(
		"%s not found in PATH or common Node.js version manager locations.\n"+
			"Please install Node.js using one of:\n"+
			"  - nvm: https://github.com/nvm-sh/nvm\n"+
			"  - volta: https://volta.sh/\n"+
			"  - asdf: https://asdf-vm.com/\n"+
			"  - fnm: https://github.com/Schniz/fnm\n"+
			"  - Direct download: https://nodejs.org/",
		name,
	)
}
