package filesystem

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateDir(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(tmpDir string) string // returns path to create
		wantErr bool
	}{
		{
			name: "creates single directory",
			setup: func(tmpDir string) string {
				return filepath.Join(tmpDir, "newdir")
			},
			wantErr: false,
		},
		{
			name: "creates nested directories",
			setup: func(tmpDir string) string {
				return filepath.Join(tmpDir, "a", "b", "c", "d")
			},
			wantErr: false,
		},
		{
			name: "handles existing directory",
			setup: func(tmpDir string) string {
				path := filepath.Join(tmpDir, "existing")
				os.MkdirAll(path, 0755)
				return path
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tmpDir := t.TempDir()
			path := tt.setup(tmpDir)

			ops := NewOperations()
			err := ops.CreateDir(path, 0755)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				info, err := os.Stat(path)
				if err != nil {
					t.Errorf("directory not created: %v", err)
					return
				}
				if !info.IsDir() {
					t.Error("created path is not a directory")
				}
			}
		})
	}
}

func TestCopyFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(tmpDir string) (src, dst string)
		content string
		wantErr bool
	}{
		{
			name: "copies file correctly",
			setup: func(tmpDir string) (string, string) {
				src := filepath.Join(tmpDir, "source.txt")
				dst := filepath.Join(tmpDir, "dest.txt")
				return src, dst
			},
			content: "hello world content",
			wantErr: false,
		},
		{
			name: "source not found returns error",
			setup: func(tmpDir string) (string, string) {
				src := filepath.Join(tmpDir, "nonexistent.txt")
				dst := filepath.Join(tmpDir, "dest.txt")
				return src, dst
			},
			content: "", // won't be written
			wantErr: true,
		},
		{
			name: "copies large file",
			setup: func(tmpDir string) (string, string) {
				src := filepath.Join(tmpDir, "large.txt")
				dst := filepath.Join(tmpDir, "large_copy.txt")
				return src, dst
			},
			content: string(make([]byte, 1024*1024)), // 1MB file
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tmpDir := t.TempDir()
			src, dst := tt.setup(tmpDir)

			// Write source file if content provided
			if tt.content != "" {
				if err := os.WriteFile(src, []byte(tt.content), 0644); err != nil {
					t.Fatalf("failed to write source file: %v", err)
				}
			}

			ops := NewOperations()
			err := ops.CopyFile(src, dst)

			if (err != nil) != tt.wantErr {
				t.Errorf("CopyFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify destination file exists and has correct content
				dstContent, err := os.ReadFile(dst)
				if err != nil {
					t.Errorf("failed to read destination file: %v", err)
					return
				}
				if string(dstContent) != tt.content {
					t.Errorf("destination content mismatch: got %d bytes, want %d bytes", len(dstContent), len(tt.content))
				}
			}
		})
	}
}

func TestCopyFile_DestinationDirMissing(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// Create source file
	src := filepath.Join(tmpDir, "source.txt")
	if err := os.WriteFile(src, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	// Destination is in a non-existent directory
	dst := filepath.Join(tmpDir, "nonexistent", "dest.txt")

	ops := NewOperations()
	err := ops.CopyFile(src, dst)

	// Should fail because destination directory doesn't exist
	if err == nil {
		t.Error("CopyFile() should return error when destination directory doesn't exist")
	}
}

func TestExists(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func(tmpDir string) string
		want  bool
	}{
		{
			name: "returns true for existing file",
			setup: func(tmpDir string) string {
				path := filepath.Join(tmpDir, "exists.txt")
				os.WriteFile(path, []byte("content"), 0644)
				return path
			},
			want: true,
		},
		{
			name: "returns true for existing directory",
			setup: func(tmpDir string) string {
				path := filepath.Join(tmpDir, "existsdir")
				os.MkdirAll(path, 0755)
				return path
			},
			want: true,
		},
		{
			name: "returns false for non-existent path",
			setup: func(tmpDir string) string {
				return filepath.Join(tmpDir, "nonexistent")
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tmpDir := t.TempDir()
			path := tt.setup(tmpDir)

			ops := NewOperations()
			got := ops.Exists(path)

			if got != tt.want {
				t.Errorf("Exists() = %v, want %v", got, tt.want)
			}
		})
	}
}
