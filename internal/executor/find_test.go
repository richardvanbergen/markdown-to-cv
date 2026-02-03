package executor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestFindNodeExecutable_UsesLookPathFirst verifies that exec.LookPath is tried first
func TestFindNodeExecutable_UsesLookPathFirst(t *testing.T) {
	// Create a temp directory with a fake npm executable
	tmpDir := t.TempDir()
	fakeNpm := filepath.Join(tmpDir, "npm")

	// Create executable file
	err := os.WriteFile(fakeNpm, []byte("#!/bin/sh\necho 'fake npm'"), 0755)
	if err != nil {
		t.Fatalf("failed to create fake npm: %v", err)
	}

	// Prepend to PATH
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", tmpDir+":"+origPath)
	defer os.Setenv("PATH", origPath)

	// FindNodeExecutable should find our fake npm via LookPath
	path, err := FindNodeExecutable("npm")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if path != fakeNpm {
		t.Errorf("expected path %s, got %s", fakeNpm, path)
	}
}

// TestFindNodeExecutable_FallbackLocations verifies fallback search when not in PATH
func TestFindNodeExecutable_FallbackLocations(t *testing.T) {
	// Create fake home directory structure with nvm
	tmpHome := t.TempDir()
	nvmBin := filepath.Join(tmpHome, ".nvm", "current", "bin")
	err := os.MkdirAll(nvmBin, 0755)
	if err != nil {
		t.Fatalf("failed to create nvm dir: %v", err)
	}

	fakeNpx := filepath.Join(nvmBin, "npx")
	err = os.WriteFile(fakeNpx, []byte("#!/bin/sh\necho 'fake npx'"), 0755)
	if err != nil {
		t.Fatalf("failed to create fake npx: %v", err)
	}

	// Set HOME to our temp directory
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	// Clear PATH so LookPath fails
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", "")
	defer os.Setenv("PATH", origPath)

	// FindNodeExecutable should find npx in fallback location
	path, err := FindNodeExecutable("npx")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if path != fakeNpx {
		t.Errorf("expected path %s, got %s", fakeNpx, path)
	}
}

// TestFindNodeExecutable_VoltaFallback verifies volta location is checked
func TestFindNodeExecutable_VoltaFallback(t *testing.T) {
	tmpHome := t.TempDir()
	voltaBin := filepath.Join(tmpHome, ".volta", "bin")
	err := os.MkdirAll(voltaBin, 0755)
	if err != nil {
		t.Fatalf("failed to create volta dir: %v", err)
	}

	fakeNode := filepath.Join(voltaBin, "node")
	err = os.WriteFile(fakeNode, []byte("#!/bin/sh\necho 'fake node'"), 0755)
	if err != nil {
		t.Fatalf("failed to create fake node: %v", err)
	}

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	origPath := os.Getenv("PATH")
	os.Setenv("PATH", "")
	defer os.Setenv("PATH", origPath)

	path, err := FindNodeExecutable("node")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if path != fakeNode {
		t.Errorf("expected path %s, got %s", fakeNode, path)
	}
}

// TestFindNodeExecutable_NotFound verifies descriptive error when not found
func TestFindNodeExecutable_NotFound(t *testing.T) {
	// Create empty home with no node installations
	tmpHome := t.TempDir()

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	origPath := os.Getenv("PATH")
	os.Setenv("PATH", "")
	defer os.Setenv("PATH", origPath)

	// Use FindNodeExecutableWithOptions with SkipSystemPaths to ensure test isolation
	// This prevents the test from finding system-installed npm in /usr/local/bin etc.
	_, err := FindNodeExecutableWithOptions("npm", &FindOptions{SkipSystemPaths: true})
	if err == nil {
		t.Fatal("expected error when executable not found")
	}

	// Error should be descriptive with install instructions
	errMsg := err.Error()
	if !strings.Contains(errMsg, "npm") {
		t.Errorf("error should mention executable name, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "not found") || !strings.Contains(errMsg, "install") {
		t.Errorf("error should include install instructions, got: %s", errMsg)
	}
}

// TestFindNodeExecutable_AsdfFallback verifies asdf shims location is checked
func TestFindNodeExecutable_AsdfFallback(t *testing.T) {
	tmpHome := t.TempDir()
	asdfBin := filepath.Join(tmpHome, ".asdf", "shims")
	err := os.MkdirAll(asdfBin, 0755)
	if err != nil {
		t.Fatalf("failed to create asdf dir: %v", err)
	}

	fakeNpm := filepath.Join(asdfBin, "npm")
	err = os.WriteFile(fakeNpm, []byte("#!/bin/sh\necho 'fake npm'"), 0755)
	if err != nil {
		t.Fatalf("failed to create fake npm: %v", err)
	}

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	origPath := os.Getenv("PATH")
	os.Setenv("PATH", "")
	defer os.Setenv("PATH", origPath)

	path, err := FindNodeExecutable("npm")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if path != fakeNpm {
		t.Errorf("expected path %s, got %s", fakeNpm, path)
	}
}

// TestFindNodeExecutable_FnmFallback verifies fnm location is checked
func TestFindNodeExecutable_FnmFallback(t *testing.T) {
	tmpHome := t.TempDir()
	fnmBin := filepath.Join(tmpHome, ".fnm", "current", "bin")
	err := os.MkdirAll(fnmBin, 0755)
	if err != nil {
		t.Fatalf("failed to create fnm dir: %v", err)
	}

	fakeNpm := filepath.Join(fnmBin, "npm")
	err = os.WriteFile(fakeNpm, []byte("#!/bin/sh\necho 'fake npm'"), 0755)
	if err != nil {
		t.Fatalf("failed to create fake npm: %v", err)
	}

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	origPath := os.Getenv("PATH")
	os.Setenv("PATH", "")
	defer os.Setenv("PATH", origPath)

	path, err := FindNodeExecutable("npm")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if path != fakeNpm {
		t.Errorf("expected path %s, got %s", fakeNpm, path)
	}
}
