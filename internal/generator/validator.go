package generator

import (
	"encoding/json"
	"fmt"

	"github.com/richq/m2cv/internal/assets"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

// Validator validates JSON Resume documents against the JSON Resume schema.
type Validator struct {
	schema *jsonschema.Schema
}

// NewValidator creates a new Validator with the embedded JSON Resume schema.
// The schema is loaded once and compiled for efficient repeated validation.
func NewValidator() (*Validator, error) {
	// Load embedded schema
	schemaData, err := assets.GetSchema("resume.schema.json")
	if err != nil {
		return nil, fmt.Errorf("failed to load schema: %w", err)
	}

	// Parse schema into interface{} for the compiler
	var schemaObj interface{}
	if err := json.Unmarshal(schemaData, &schemaObj); err != nil {
		return nil, fmt.Errorf("failed to parse schema JSON: %w", err)
	}

	// Create compiler and add the schema as a resource
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("resume.schema.json", schemaObj); err != nil {
		return nil, fmt.Errorf("failed to add schema resource: %w", err)
	}

	// Compile the schema
	schema, err := compiler.Compile("resume.schema.json")
	if err != nil {
		return nil, fmt.Errorf("failed to compile schema: %w", err)
	}

	return &Validator{schema: schema}, nil
}

// Validate checks if the JSON Resume document is valid according to the schema.
// Returns nil if valid, or an error describing validation failures.
func (v *Validator) Validate(resumeJSON []byte) error {
	// First verify the input is valid JSON
	var doc interface{}
	if err := json.Unmarshal(resumeJSON, &doc); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	// Validate against schema
	if err := v.schema.Validate(doc); err != nil {
		// jsonschema returns detailed validation errors
		return fmt.Errorf("schema validation failed: %w", err)
	}

	return nil
}
