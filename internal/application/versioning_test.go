package application

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestListVersions(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		want     []int
		wantErr  bool
		setupDir bool // if false, don't create the directory
	}{
		{
			name:     "empty directory",
			files:    []string{},
			want:     []int{},
			wantErr:  false,
			setupDir: true,
		},
		{
			name:     "single version",
			files:    []string{"optimized-cv-1.md"},
			want:     []int{1},
			wantErr:  false,
			setupDir: true,
		},
		{
			name:     "multiple versions unsorted",
			files:    []string{"optimized-cv-1.md", "optimized-cv-3.md", "optimized-cv-2.md"},
			want:     []int{1, 2, 3},
			wantErr:  false,
			setupDir: true,
		},
		{
			name:     "mixed files with valid and invalid names",
			files:    []string{"optimized-cv-1.md", "other.txt", "optimized-cv-abc.md", "readme.md"},
			want:     []int{1},
			wantErr:  false,
			setupDir: true,
		},
		{
			name: "ignores malformed version numbers",
			files: []string{
				"optimized-cv-1.md",
				"optimized-cv-abc.md",
				"optimized-cv-.md",
				"optimized-cv-1.5.md",
				"optimized-cv-2.md",
			},
			want:     []int{1, 2},
			wantErr:  false,
			setupDir: true,
		},
		{
			name:     "ignores zero and negative versions",
			files:    []string{"optimized-cv-0.md", "optimized-cv--1.md", "optimized-cv-1.md"},
			want:     []int{1},
			wantErr:  false,
			setupDir: true,
		},
		{
			name:     "non-existent directory",
			files:    nil,
			want:     []int{},
			wantErr:  false, // Glob returns empty, not error for non-existent dir
			setupDir: false,
		},
		{
			name:     "high version numbers",
			files:    []string{"optimized-cv-100.md", "optimized-cv-5.md", "optimized-cv-50.md"},
			want:     []int{5, 50, 100},
			wantErr:  false,
			setupDir: true,
		},
		{
			name:     "only invalid files",
			files:    []string{"optimized-cv-abc.md", "optimized-cv-.md", "other.txt"},
			want:     []int{},
			wantErr:  false,
			setupDir: true,
		},
	}

	for _, tt := range tests {
		tt := tt // Capture for parallel
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var appDir string
			if tt.setupDir {
				appDir = t.TempDir()
				// Create test files
				for _, f := range tt.files {
					path := filepath.Join(appDir, f)
					if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
						t.Fatalf("failed to create test file %s: %v", f, err)
					}
				}
			} else {
				// Use a path that definitely doesn't exist
				appDir = filepath.Join(t.TempDir(), "nonexistent")
			}

			got, err := ListVersions(appDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListVersions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Normalize nil vs empty slice for comparison
			if len(got) == 0 && len(tt.want) == 0 {
				return // Both empty, consider equal
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListVersions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLatestVersionPath(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		wantFile string // expected filename (not full path)
		wantErr  bool
	}{
		{
			name:     "empty directory returns empty string",
			files:    []string{},
			wantFile: "",
			wantErr:  false,
		},
		{
			name:     "single version returns that version",
			files:    []string{"optimized-cv-1.md"},
			wantFile: "optimized-cv-1.md",
			wantErr:  false,
		},
		{
			name:     "multiple versions returns highest",
			files:    []string{"optimized-cv-1.md", "optimized-cv-5.md", "optimized-cv-3.md"},
			wantFile: "optimized-cv-5.md",
			wantErr:  false,
		},
		{
			name:     "with gaps in versions",
			files:    []string{"optimized-cv-1.md", "optimized-cv-10.md"},
			wantFile: "optimized-cv-10.md",
			wantErr:  false,
		},
		{
			name:     "ignores invalid files returns highest valid",
			files:    []string{"optimized-cv-1.md", "optimized-cv-abc.md", "optimized-cv-3.md"},
			wantFile: "optimized-cv-3.md",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			appDir := t.TempDir()
			// Create test files
			for _, f := range tt.files {
				path := filepath.Join(appDir, f)
				if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
					t.Fatalf("failed to create test file %s: %v", f, err)
				}
			}

			got, err := LatestVersionPath(appDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("LatestVersionPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantFile == "" {
				if got != "" {
					t.Errorf("LatestVersionPath() = %v, want empty string", got)
				}
				return
			}

			wantPath := filepath.Join(appDir, tt.wantFile)
			if got != wantPath {
				t.Errorf("LatestVersionPath() = %v, want %v", got, wantPath)
			}
		})
	}
}

func TestNextVersionPath(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		wantFile string // expected filename (not full path)
		wantErr  bool
	}{
		{
			name:     "empty directory returns version 1",
			files:    []string{},
			wantFile: "optimized-cv-1.md",
			wantErr:  false,
		},
		{
			name:     "with version 1 returns version 2",
			files:    []string{"optimized-cv-1.md"},
			wantFile: "optimized-cv-2.md",
			wantErr:  false,
		},
		{
			name:     "with gaps returns max plus one",
			files:    []string{"optimized-cv-1.md", "optimized-cv-2.md", "optimized-cv-5.md"},
			wantFile: "optimized-cv-6.md",
			wantErr:  false,
		},
		{
			name:     "ignores invalid files",
			files:    []string{"optimized-cv-1.md", "optimized-cv-abc.md"},
			wantFile: "optimized-cv-2.md",
			wantErr:  false,
		},
		{
			name:     "high version number",
			files:    []string{"optimized-cv-100.md"},
			wantFile: "optimized-cv-101.md",
			wantErr:  false,
		},
		{
			name:     "with only invalid files returns version 1",
			files:    []string{"optimized-cv-abc.md", "other.txt"},
			wantFile: "optimized-cv-1.md",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			appDir := t.TempDir()
			// Create test files
			for _, f := range tt.files {
				path := filepath.Join(appDir, f)
				if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
					t.Fatalf("failed to create test file %s: %v", f, err)
				}
			}

			got, err := NextVersionPath(appDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("NextVersionPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			wantPath := filepath.Join(appDir, tt.wantFile)
			if got != wantPath {
				t.Errorf("NextVersionPath() = %v, want %v", got, wantPath)
			}
		})
	}
}

// TestVersioningIntegration tests the functions work together correctly
func TestVersioningIntegration(t *testing.T) {
	appDir := t.TempDir()

	// Initially empty
	versions, err := ListVersions(appDir)
	if err != nil {
		t.Fatalf("ListVersions failed: %v", err)
	}
	if len(versions) != 0 {
		t.Errorf("expected empty versions, got %v", versions)
	}

	latest, err := LatestVersionPath(appDir)
	if err != nil {
		t.Fatalf("LatestVersionPath failed: %v", err)
	}
	if latest != "" {
		t.Errorf("expected empty latest path, got %v", latest)
	}

	next, err := NextVersionPath(appDir)
	if err != nil {
		t.Fatalf("NextVersionPath failed: %v", err)
	}
	expectedNext := filepath.Join(appDir, "optimized-cv-1.md")
	if next != expectedNext {
		t.Errorf("NextVersionPath = %v, want %v", next, expectedNext)
	}

	// Simulate creating version 1
	if err := os.WriteFile(next, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// Check versions updated
	versions, err = ListVersions(appDir)
	if err != nil {
		t.Fatalf("ListVersions failed: %v", err)
	}
	if !reflect.DeepEqual(versions, []int{1}) {
		t.Errorf("expected [1], got %v", versions)
	}

	latest, err = LatestVersionPath(appDir)
	if err != nil {
		t.Fatalf("LatestVersionPath failed: %v", err)
	}
	if latest != next {
		t.Errorf("LatestVersionPath = %v, want %v", latest, next)
	}

	next, err = NextVersionPath(appDir)
	if err != nil {
		t.Fatalf("NextVersionPath failed: %v", err)
	}
	expectedNext = filepath.Join(appDir, "optimized-cv-2.md")
	if next != expectedNext {
		t.Errorf("NextVersionPath = %v, want %v", next, expectedNext)
	}

	// Create version 2
	if err := os.WriteFile(next, []byte("content v2"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// Check versions updated
	versions, err = ListVersions(appDir)
	if err != nil {
		t.Fatalf("ListVersions failed: %v", err)
	}
	if !reflect.DeepEqual(versions, []int{1, 2}) {
		t.Errorf("expected [1, 2], got %v", versions)
	}
}

// TestConstants verifies the constants are correct
func TestConstants(t *testing.T) {
	if OptimizedCVPrefix != "optimized-cv-" {
		t.Errorf("OptimizedCVPrefix = %q, want %q", OptimizedCVPrefix, "optimized-cv-")
	}
	if OptimizedCVSuffix != ".md" {
		t.Errorf("OptimizedCVSuffix = %q, want %q", OptimizedCVSuffix, ".md")
	}
}
