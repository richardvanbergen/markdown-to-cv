// Package extractor provides AI-powered extraction utilities for m2cv.
package extractor

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"unicode"

	"github.com/richq/m2cv/internal/assets"
	"github.com/richq/m2cv/internal/executor"
)

// maxFilenameLength is the maximum length for sanitized folder names.
const maxFilenameLength = 50

// SanitizeFilename converts a string to a filesystem-safe folder name.
// It:
//   - Converts to lowercase
//   - Replaces spaces, slashes, and backslashes with hyphens
//   - Removes characters other than letters, numbers, hyphens, and underscores
//   - Collapses multiple consecutive hyphens into one
//   - Trims leading and trailing hyphens
//   - Truncates to maxFilenameLength at word boundary if possible
func SanitizeFilename(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)

	// Replace spaces, slashes, backslashes with hyphens
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ReplaceAll(s, "\\", "-")

	// Keep only letters, numbers, hyphens, underscores
	var result strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' {
			result.WriteRune(r)
		}
	}
	s = result.String()

	// Collapse multiple hyphens into one
	multiHyphen := regexp.MustCompile(`-+`)
	s = multiHyphen.ReplaceAllString(s, "-")

	// Trim leading and trailing hyphens
	s = strings.Trim(s, "-")

	// Truncate to max length at word boundary if possible
	if len(s) > maxFilenameLength {
		s = truncateAtBoundary(s, maxFilenameLength)
	}

	return s
}

// truncateAtBoundary truncates the string to maxLen, preferring to break at a hyphen.
func truncateAtBoundary(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	// Try to find a hyphen to break at within the allowed length
	truncated := s[:maxLen]
	lastHyphen := strings.LastIndex(truncated, "-")

	// If we find a hyphen reasonably close to the end, break there
	// "reasonably close" = in the last third of the string
	if lastHyphen > maxLen*2/3 {
		return truncated[:lastHyphen]
	}

	// Otherwise, just truncate at maxLen and trim any trailing hyphen
	return strings.TrimRight(truncated, "-")
}

// ExtractFolderName uses Claude to extract a company-role folder name from a job description.
// It loads the extract-name prompt template, calls the Claude executor, and sanitizes the result.
func ExtractFolderName(ctx context.Context, exec executor.ClaudeExecutor, jobDesc string) (string, error) {
	// Load prompt template
	promptTemplate, err := assets.GetPrompt("extract-name")
	if err != nil {
		return "", err
	}

	// Replace placeholder with job description
	prompt := strings.ReplaceAll(promptTemplate, "{{.JobDescription}}", jobDesc)

	// Execute via Claude with default settings (text output)
	result, err := exec.Execute(ctx, prompt)
	if err != nil {
		return "", err
	}

	// Trim whitespace from response
	result = strings.TrimSpace(result)

	// Sanitize the result
	sanitized := SanitizeFilename(result)

	// Return error if result is empty after sanitization
	if sanitized == "" {
		return "", errors.New("extracted folder name is empty after sanitization")
	}

	return sanitized, nil
}
