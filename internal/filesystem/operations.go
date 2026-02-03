// Package filesystem provides file system operations for m2cv.
// It uses a repository interface pattern for testability.
package filesystem

import (
	"io"
	"os"
)

// Operations defines the interface for filesystem operations.
type Operations interface {
	// CreateDir creates a directory and any necessary parent directories.
	CreateDir(path string, perm os.FileMode) error
	// CopyFile copies a file from src to dst with streaming.
	CopyFile(src, dst string) error
	// Exists checks if a path exists.
	Exists(path string) bool
}

// osOperations implements Operations using the real filesystem.
type osOperations struct{}

// NewOperations creates a new filesystem Operations implementation.
func NewOperations() Operations {
	return &osOperations{}
}

// CreateDir creates a directory and any necessary parent directories.
func (o *osOperations) CreateDir(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// CopyFile copies a file from src to dst using streaming.
// The destination file is synced before closing to ensure durability.
func (o *osOperations) CopyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	return dstFile.Sync()
}

// Exists checks if a path exists (file or directory).
func (o *osOperations) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
