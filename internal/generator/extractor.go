// Package generator provides the core pipeline components for resume generation:
// JSON extraction from Claude output, schema validation, and PDF export.
package generator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
)

// ExtractJSON extracts a JSON object from Claude output that may contain
// markdown fences, explanatory text, or other non-JSON content.
//
// The extraction process:
// 1. Strip markdown code fences (```json or ```)
// 2. Find JSON object boundaries (first '{' to last '}')
// 3. Validate the extracted content is valid JSON
//
// Returns the extracted JSON as json.RawMessage, preserving the original formatting.
func ExtractJSON(claudeOutput []byte) (json.RawMessage, error) {
	if len(claudeOutput) == 0 {
		return nil, fmt.Errorf("empty input: no content to extract JSON from")
	}

	// 1. Strip markdown code fences
	content := stripMarkdownFences(claudeOutput)

	// 2. Find JSON object boundaries
	start := bytes.IndexByte(content, '{')
	end := bytes.LastIndexByte(content, '}')

	if start == -1 || end == -1 || start > end {
		// Provide helpful error with snippet of input
		snippet := truncateForError(claudeOutput, 200)
		return nil, fmt.Errorf("no JSON object found in output (expected '{' and '}')\nInput snippet:\n%s", snippet)
	}

	extracted := content[start : end+1]

	// 3. Validate it's parseable JSON
	var raw json.RawMessage
	if err := json.Unmarshal(extracted, &raw); err != nil {
		// Include both the extraction attempt and the parse error
		snippet := truncateForError(extracted, 500)
		return nil, fmt.Errorf("extracted content is not valid JSON: %w\nExtracted content:\n%s", err, snippet)
	}

	return raw, nil
}

// stripMarkdownFences removes ```json...``` or ```...``` fences from the input.
// If fences are found, returns the content inside them.
// If no fences are found, returns the original data.
func stripMarkdownFences(data []byte) []byte {
	// Match ```json or ``` followed by optional newline, content, optional newline, closing ```
	// Using (?s) for DOTALL mode so . matches newlines
	re := regexp.MustCompile("(?s)```(?:json)?\\n?(.*?)\\n?```")
	matches := re.FindSubmatch(data)
	if matches != nil && len(matches) > 1 {
		return matches[1]
	}
	return data
}

// truncateForError truncates a byte slice for inclusion in error messages.
// Adds "..." suffix if truncated.
func truncateForError(data []byte, maxLen int) string {
	if len(data) <= maxLen {
		return string(data)
	}
	return string(data[:maxLen]) + "..."
}
