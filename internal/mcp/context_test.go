package mcp

import (
	"testing"
)

func TestInteractiveContextEncodeDecode(t *testing.T) {
	original := &InteractiveContext{
		ApplicationDir: "applications/test-job",
		BaseCV:         "# John Doe\n\nSoftware Engineer",
		JobDescription: "We are looking for a software engineer...",
		ATSMode:        true,
		Model:          "claude-sonnet-4-20250514",
	}

	// Encode
	encoded, err := original.Encode()
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if encoded == "" {
		t.Fatal("Encoded string is empty")
	}

	// Decode
	decoded, err := DecodeContext(encoded)
	if err != nil {
		t.Fatalf("DecodeContext failed: %v", err)
	}

	// Verify fields
	if decoded.ApplicationDir != original.ApplicationDir {
		t.Errorf("ApplicationDir mismatch: got %q, want %q", decoded.ApplicationDir, original.ApplicationDir)
	}
	if decoded.BaseCV != original.BaseCV {
		t.Errorf("BaseCV mismatch: got %q, want %q", decoded.BaseCV, original.BaseCV)
	}
	if decoded.JobDescription != original.JobDescription {
		t.Errorf("JobDescription mismatch: got %q, want %q", decoded.JobDescription, original.JobDescription)
	}
	if decoded.ATSMode != original.ATSMode {
		t.Errorf("ATSMode mismatch: got %v, want %v", decoded.ATSMode, original.ATSMode)
	}
	if decoded.Model != original.Model {
		t.Errorf("Model mismatch: got %q, want %q", decoded.Model, original.Model)
	}
}

func TestDecodeContextInvalidBase64(t *testing.T) {
	_, err := DecodeContext("not-valid-base64!!!")
	if err == nil {
		t.Error("Expected error for invalid base64, got nil")
	}
}

func TestDecodeContextInvalidJSON(t *testing.T) {
	// Valid base64 but invalid JSON
	_, err := DecodeContext("bm90LWpzb24=") // "not-json"
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}
