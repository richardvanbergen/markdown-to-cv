// Package assets provides access to embedded prompt templates and JSON schemas.
// These files are compiled into the binary using Go's embed directive.
package assets

import (
	"embed"
	"fmt"
	"path/filepath"
	"strings"
)

//go:embed prompts/*.txt
var promptFS embed.FS

//go:embed schema/*.json
var schemaFS embed.FS

// GetPrompt reads a prompt template by name (without extension).
// For example, GetPrompt("optimize") reads "prompts/optimize.txt".
func GetPrompt(name string) (string, error) {
	path := filepath.Join("prompts", name+".txt")
	data, err := promptFS.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("prompt %q not found: %w", name, err)
	}
	return string(data), nil
}

// GetSchema reads a schema file by name (including extension).
// For example, GetSchema("resume.schema.json") reads "schema/resume.schema.json".
func GetSchema(name string) ([]byte, error) {
	path := filepath.Join("schema", name)
	data, err := schemaFS.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("schema %q not found: %w", name, err)
	}
	return data, nil
}

// ListPrompts returns all available prompt names (without extension).
// Useful for debugging and validation.
func ListPrompts() ([]string, error) {
	entries, err := promptFS.ReadDir("prompts")
	if err != nil {
		return nil, fmt.Errorf("failed to read prompts directory: %w", err)
	}

	var names []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".txt") {
			name := strings.TrimSuffix(entry.Name(), ".txt")
			names = append(names, name)
		}
	}
	return names, nil
}
