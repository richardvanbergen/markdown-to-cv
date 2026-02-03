// Package application provides utilities for managing application folders.
// This includes versioning for optimized CV output files.
package application

import (
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// Constants for optimized CV filename pattern.
const (
	// OptimizedCVPrefix is the prefix for optimized CV files.
	OptimizedCVPrefix = "optimized-cv-"
	// OptimizedCVSuffix is the suffix for optimized CV files.
	OptimizedCVSuffix = ".md"
)

// ListVersions returns a sorted slice of version numbers found in the application directory.
// It looks for files matching the pattern optimized-cv-N.md where N is a positive integer.
// Returns empty slice if no versions exist (not an error).
// Malformed filenames (e.g., optimized-cv-abc.md) are silently ignored.
func ListVersions(appDir string) ([]int, error) {
	pattern := filepath.Join(appDir, OptimizedCVPrefix+"*"+OptimizedCVSuffix)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("glob pattern error: %w", err)
	}

	var versions []int
	for _, match := range matches {
		base := filepath.Base(match)
		// Extract version number from filename
		// Format: optimized-cv-N.md
		if !strings.HasPrefix(base, OptimizedCVPrefix) || !strings.HasSuffix(base, OptimizedCVSuffix) {
			continue
		}

		// Extract the number part
		numStr := strings.TrimPrefix(base, OptimizedCVPrefix)
		numStr = strings.TrimSuffix(numStr, OptimizedCVSuffix)

		num, err := strconv.Atoi(numStr)
		if err != nil {
			// Malformed filename (e.g., optimized-cv-abc.md) - ignore
			continue
		}

		if num > 0 {
			versions = append(versions, num)
		}
	}

	sort.Ints(versions)
	return versions, nil
}

// LatestVersionPath returns the path to the highest versioned optimized CV file.
// Returns ("", nil) if no versions exist - this is not an error, just means no optimized CV yet.
func LatestVersionPath(appDir string) (string, error) {
	versions, err := ListVersions(appDir)
	if err != nil {
		return "", err
	}

	if len(versions) == 0 {
		return "", nil
	}

	latest := versions[len(versions)-1]
	return filepath.Join(appDir, fmt.Sprintf("%s%d%s", OptimizedCVPrefix, latest, OptimizedCVSuffix)), nil
}

// NextVersionPath returns the path for the next version of the optimized CV.
// If no versions exist, returns path for version 1.
// Otherwise returns path for (max existing version + 1).
func NextVersionPath(appDir string) (string, error) {
	versions, err := ListVersions(appDir)
	if err != nil {
		return "", err
	}

	nextVersion := 1
	if len(versions) > 0 {
		nextVersion = versions[len(versions)-1] + 1
	}

	return filepath.Join(appDir, fmt.Sprintf("%s%d%s", OptimizedCVPrefix, nextVersion, OptimizedCVSuffix)), nil
}
