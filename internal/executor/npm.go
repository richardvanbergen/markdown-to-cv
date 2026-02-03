package executor

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// NPMExecutor executes npm commands for package management.
// It uses FindNodeExecutable to resolve npm paths, supporting
// various Node.js version managers.
type NPMExecutor interface {
	// Install installs npm packages in the specified directory.
	Install(ctx context.Context, dir string, packages ...string) error

	// CheckInstalled checks if a package is installed in node_modules.
	CheckInstalled(ctx context.Context, dir string, pkg string) (bool, error)

	// Init initializes a new package.json in the directory.
	Init(ctx context.Context, dir string) error
}

// npmExecutor is the default implementation of NPMExecutor.
type npmExecutor struct {
	npmPath     string
	findOptions *FindOptions
}

// NPMOption modifies the NPMExecutor construction.
type NPMOption func(*npmExecutor)

// WithNPMPath sets a custom path to the npm binary.
func WithNPMPath(path string) NPMOption {
	return func(e *npmExecutor) {
		e.npmPath = path
	}
}

// WithFindOptions sets options for FindNodeExecutable when locating npm.
// Useful for testing to ensure isolation from host system binaries.
func WithFindOptions(opts *FindOptions) NPMOption {
	return func(e *npmExecutor) {
		e.findOptions = opts
	}
}

// NewNPMExecutor creates a new NPMExecutor.
// If no custom path is provided, it uses FindNodeExecutable to locate npm.
func NewNPMExecutor(opts ...NPMOption) (NPMExecutor, error) {
	e := &npmExecutor{}

	// Apply options first
	for _, opt := range opts {
		opt(e)
	}

	// If no custom path, find npm using FindNodeExecutable
	if e.npmPath == "" {
		path, err := FindNodeExecutableWithOptions("npm", e.findOptions)
		if err != nil {
			return nil, fmt.Errorf("could not find npm: %w", err)
		}
		e.npmPath = path
	}

	return e, nil
}

// Install installs npm packages in the specified directory.
// Runs: npm install <packages...>
func (e *npmExecutor) Install(ctx context.Context, dir string, packages ...string) error {
	args := append([]string{"install"}, packages...)
	return e.runNPM(ctx, dir, args...)
}

// CheckInstalled checks if a package exists in node_modules.
// This is a filesystem check, not an npm command.
func (e *npmExecutor) CheckInstalled(ctx context.Context, dir string, pkg string) (bool, error) {
	pkgPath := filepath.Join(dir, "node_modules", pkg)
	info, err := os.Stat(pkgPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("error checking package %s: %w", pkg, err)
	}
	return info.IsDir(), nil
}

// Init initializes a new package.json.
// Runs: npm init -y
func (e *npmExecutor) Init(ctx context.Context, dir string) error {
	return e.runNPM(ctx, dir, "init", "-y")
}

// runNPM executes an npm command in the specified directory.
// Uses bytes.Buffer for output capture (consistent with ClaudeExecutor pattern).
func (e *npmExecutor) runNPM(ctx context.Context, dir string, args ...string) error {
	cmd := exec.CommandContext(ctx, e.npmPath, args...)
	cmd.Dir = dir

	// Use bytes.Buffer for stdout/stderr (Pattern 1: streaming subprocess)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Start and wait (Pattern 1: cmd.Start() + cmd.Wait())
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start npm: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		stderrContent := strings.TrimSpace(stderr.String())
		if stderrContent != "" {
			return fmt.Errorf("npm %s failed: %w\nstderr: %s", args[0], err, stderrContent)
		}
		return fmt.Errorf("npm %s failed: %w", args[0], err)
	}

	return nil
}
