package executor

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNPMExecutor_CheckInstalled verifies node_modules existence check
func TestNPMExecutor_CheckInstalled(t *testing.T) {
	// Create temp directory with node_modules structure
	tmpDir := t.TempDir()
	nodeModules := filepath.Join(tmpDir, "node_modules", "some-package")
	err := os.MkdirAll(nodeModules, 0755)
	if err != nil {
		t.Fatalf("failed to create node_modules: %v", err)
	}

	// Create fake npm executable
	fakeNpm := filepath.Join(tmpDir, "bin", "npm")
	err = os.MkdirAll(filepath.Dir(fakeNpm), 0755)
	if err != nil {
		t.Fatalf("failed to create bin dir: %v", err)
	}
	err = os.WriteFile(fakeNpm, []byte("#!/bin/sh\necho 'fake npm'"), 0755)
	if err != nil {
		t.Fatalf("failed to create fake npm: %v", err)
	}

	executor, err := NewNPMExecutor(WithNPMPath(fakeNpm))
	if err != nil {
		t.Fatalf("failed to create executor: %v", err)
	}

	ctx := context.Background()

	// Check installed package
	installed, err := executor.CheckInstalled(ctx, tmpDir, "some-package")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if !installed {
		t.Error("expected package to be installed")
	}

	// Check non-installed package
	installed, err = executor.CheckInstalled(ctx, tmpDir, "not-installed")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if installed {
		t.Error("expected package to NOT be installed")
	}
}

// TestNPMExecutor_Init runs npm init -y in directory
func TestNPMExecutor_Init(t *testing.T) {
	tmpDir := t.TempDir()

	// Create fake npm that creates package.json
	fakeNpm := filepath.Join(tmpDir, "bin", "npm")
	err := os.MkdirAll(filepath.Dir(fakeNpm), 0755)
	if err != nil {
		t.Fatalf("failed to create bin dir: %v", err)
	}

	// Script that simulates npm init -y
	script := `#!/bin/sh
if [ "$1" = "init" ] && [ "$2" = "-y" ]; then
    echo '{"name": "test"}' > "$PWD/package.json"
    exit 0
fi
exit 1
`
	err = os.WriteFile(fakeNpm, []byte(script), 0755)
	if err != nil {
		t.Fatalf("failed to create fake npm: %v", err)
	}

	executor, err := NewNPMExecutor(WithNPMPath(fakeNpm))
	if err != nil {
		t.Fatalf("failed to create executor: %v", err)
	}

	// Create project directory
	projectDir := filepath.Join(tmpDir, "project")
	err = os.MkdirAll(projectDir, 0755)
	if err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}

	ctx := context.Background()

	err = executor.Init(ctx, projectDir)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	// Verify package.json was created
	pkgPath := filepath.Join(projectDir, "package.json")
	if _, err := os.Stat(pkgPath); os.IsNotExist(err) {
		t.Error("expected package.json to be created")
	}
}

// TestNPMExecutor_Install runs npm install with packages
func TestNPMExecutor_Install(t *testing.T) {
	tmpDir := t.TempDir()

	// Create fake npm that records arguments
	fakeNpm := filepath.Join(tmpDir, "bin", "npm")
	argsFile := filepath.Join(tmpDir, "npm_args.txt")
	err := os.MkdirAll(filepath.Dir(fakeNpm), 0755)
	if err != nil {
		t.Fatalf("failed to create bin dir: %v", err)
	}

	// Script that records arguments to file
	script := `#!/bin/sh
echo "$@" >> ` + argsFile + `
exit 0
`
	err = os.WriteFile(fakeNpm, []byte(script), 0755)
	if err != nil {
		t.Fatalf("failed to create fake npm: %v", err)
	}

	executor, err := NewNPMExecutor(WithNPMPath(fakeNpm))
	if err != nil {
		t.Fatalf("failed to create executor: %v", err)
	}

	// Create project directory
	projectDir := filepath.Join(tmpDir, "project")
	err = os.MkdirAll(projectDir, 0755)
	if err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}

	ctx := context.Background()

	err = executor.Install(ctx, projectDir, "resumed", "jsonresume-theme-flat")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	// Verify npm was called with correct arguments
	args, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("failed to read args file: %v", err)
	}

	argsStr := string(args)
	if !strings.Contains(argsStr, "install") {
		t.Errorf("expected 'install' in args, got: %s", argsStr)
	}
	if !strings.Contains(argsStr, "resumed") {
		t.Errorf("expected 'resumed' in args, got: %s", argsStr)
	}
	if !strings.Contains(argsStr, "jsonresume-theme-flat") {
		t.Errorf("expected 'jsonresume-theme-flat' in args, got: %s", argsStr)
	}
}

// TestNPMExecutor_InstallError handles npm errors
func TestNPMExecutor_InstallError(t *testing.T) {
	tmpDir := t.TempDir()

	// Create fake npm that fails
	fakeNpm := filepath.Join(tmpDir, "bin", "npm")
	err := os.MkdirAll(filepath.Dir(fakeNpm), 0755)
	if err != nil {
		t.Fatalf("failed to create bin dir: %v", err)
	}

	script := `#!/bin/sh
echo "npm ERR! some error" >&2
exit 1
`
	err = os.WriteFile(fakeNpm, []byte(script), 0755)
	if err != nil {
		t.Fatalf("failed to create fake npm: %v", err)
	}

	executor, err := NewNPMExecutor(WithNPMPath(fakeNpm))
	if err != nil {
		t.Fatalf("failed to create executor: %v", err)
	}

	projectDir := filepath.Join(tmpDir, "project")
	err = os.MkdirAll(projectDir, 0755)
	if err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}

	ctx := context.Background()

	err = executor.Install(ctx, projectDir, "some-package")
	if err == nil {
		t.Fatal("expected error when npm fails")
	}

	// Error should include stderr
	if !strings.Contains(err.Error(), "npm ERR!") {
		t.Errorf("error should include stderr, got: %v", err)
	}
}

// TestNPMExecutor_UsesDir verifies commands run in specified directory
func TestNPMExecutor_UsesDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create fake npm that outputs current directory
	fakeNpm := filepath.Join(tmpDir, "bin", "npm")
	outputFile := filepath.Join(tmpDir, "cwd.txt")
	err := os.MkdirAll(filepath.Dir(fakeNpm), 0755)
	if err != nil {
		t.Fatalf("failed to create bin dir: %v", err)
	}

	script := `#!/bin/sh
pwd > ` + outputFile + `
exit 0
`
	err = os.WriteFile(fakeNpm, []byte(script), 0755)
	if err != nil {
		t.Fatalf("failed to create fake npm: %v", err)
	}

	executor, err := NewNPMExecutor(WithNPMPath(fakeNpm))
	if err != nil {
		t.Fatalf("failed to create executor: %v", err)
	}

	// Create specific project directory
	projectDir := filepath.Join(tmpDir, "specific", "project", "path")
	err = os.MkdirAll(projectDir, 0755)
	if err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}

	ctx := context.Background()

	err = executor.Install(ctx, projectDir, "test-package")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	// Verify command ran in correct directory
	cwd, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read cwd file: %v", err)
	}

	cwdStr := strings.TrimSpace(string(cwd))
	if cwdStr != projectDir {
		t.Errorf("expected cwd %q, got %q", projectDir, cwdStr)
	}
}

// TestNPMExecutor_NotFound verifies error when npm not found
func TestNPMExecutor_NotFound(t *testing.T) {
	// Set up empty PATH and HOME
	tmpHome := t.TempDir()

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	origPath := os.Getenv("PATH")
	os.Setenv("PATH", "")
	defer os.Setenv("PATH", origPath)

	_, err := NewNPMExecutor()
	if err == nil {
		t.Fatal("expected error when npm not found")
	}

	if !strings.Contains(err.Error(), "npm") {
		t.Errorf("error should mention npm, got: %v", err)
	}
}

// TestNPMExecutor_FindsNPMInPath verifies npm is found via PATH
func TestNPMExecutor_FindsNPMInPath(t *testing.T) {
	tmpDir := t.TempDir()

	// Create fake npm in temp directory
	fakeNpm := filepath.Join(tmpDir, "npm")
	err := os.WriteFile(fakeNpm, []byte("#!/bin/sh\necho 'fake'"), 0755)
	if err != nil {
		t.Fatalf("failed to create fake npm: %v", err)
	}

	// Add to PATH
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", tmpDir+":"+origPath)
	defer os.Setenv("PATH", origPath)

	executor, err := NewNPMExecutor()
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if executor == nil {
		t.Error("expected executor to be created")
	}
}
