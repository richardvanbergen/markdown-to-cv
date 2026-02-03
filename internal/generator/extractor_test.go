package generator

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestExtractJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantJSON    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "clean JSON object",
			input:    `{"name": "John Doe", "email": "john@example.com"}`,
			wantJSON: `{"name": "John Doe", "email": "john@example.com"}`,
			wantErr:  false,
		},
		{
			name: "JSON wrapped in markdown fences",
			input: "```json\n" + `{"basics": {"name": "Jane"}}` + "\n```",
			wantJSON: `{"basics": {"name": "Jane"}}`,
			wantErr:  false,
		},
		{
			name: "JSON wrapped in plain markdown fences (no json tag)",
			input: "```\n" + `{"work": []}` + "\n```",
			wantJSON: `{"work": []}`,
			wantErr:  false,
		},
		{
			name: "JSON with explanatory text before",
			input: `Here is the JSON Resume document:

{"basics": {"name": "Test User"}}`,
			wantJSON: `{"basics": {"name": "Test User"}}`,
			wantErr:  false,
		},
		{
			name: "JSON with explanatory text after",
			input: `{"education": [{"institution": "MIT"}]}

I hope this JSON Resume format meets your needs!`,
			wantJSON: `{"education": [{"institution": "MIT"}]}`,
			wantErr:  false,
		},
		{
			name: "JSON with explanatory text before and after",
			input: `Here is the converted JSON Resume:

{"basics": {"name": "Full Example", "label": "Developer"}, "work": []}

Let me know if you need any changes.`,
			wantJSON: `{"basics": {"name": "Full Example", "label": "Developer"}, "work": []}`,
			wantErr:  false,
		},
		{
			name: "mixed text and fences",
			input: `I've converted your CV to JSON Resume format.

` + "```json\n" + `{
  "basics": {
    "name": "Mixed Example"
  }
}` + "\n```" + `

This follows the JSON Resume schema.`,
			wantJSON: `{
  "basics": {
    "name": "Mixed Example"
  }
}`,
			wantErr: false,
		},
		{
			name:        "empty input",
			input:       "",
			wantErr:     true,
			errContains: "empty input",
		},
		{
			name:        "no JSON content - plain text",
			input:       "This is just some regular text without any JSON.",
			wantErr:     true,
			errContains: "no JSON object found",
		},
		{
			name:        "no JSON content - only array",
			input:       `["item1", "item2"]`,
			wantErr:     true,
			errContains: "no JSON object found",
		},
		{
			name:        "invalid JSON syntax - unclosed brace (no closing)",
			input:       `{"name": "Unclosed`,
			wantErr:     true,
			errContains: "no JSON object found", // No closing brace means no valid boundaries
		},
		{
			name:        "invalid JSON syntax - unclosed brace (with closing but still invalid)",
			input:       `{"name": "Unclosed}`,
			wantErr:     true,
			errContains: "not valid JSON", // Has boundaries but content is invalid
		},
		{
			name:        "invalid JSON syntax - trailing comma",
			input:       `{"name": "Test",}`,
			wantErr:     true,
			errContains: "not valid JSON",
		},
		{
			name:        "invalid JSON syntax - missing quotes",
			input:       `{name: "Test"}`,
			wantErr:     true,
			errContains: "not valid JSON",
		},
		{
			name: "multiple JSON objects - first valid one wins",
			input: `{"first": true}

Some text

{"second": true}`,
			wantJSON: `{"first": true}

Some text

{"second": true}`,
			// Note: The current implementation finds first '{' to last '}',
			// so it will capture everything between them. This is intentional
			// as valid JSON Resume should be a single object.
			wantErr: true, // This will fail JSON parsing due to text in between
			errContains: "not valid JSON",
		},
		{
			name:     "nested JSON objects",
			input:    `{"outer": {"inner": {"deep": "value"}}}`,
			wantJSON: `{"outer": {"inner": {"deep": "value"}}}`,
			wantErr:  false,
		},
		{
			name: "JSON with escaped characters",
			input: `{"text": "Hello \"World\"", "path": "C:\\Users\\test"}`,
			wantJSON: `{"text": "Hello \"World\"", "path": "C:\\Users\\test"}`,
			wantErr:  false,
		},
		{
			name:     "minimal valid JSON",
			input:    `{}`,
			wantJSON: `{}`,
			wantErr:  false,
		},
		{
			name: "realistic JSON Resume with fences",
			input: "Here's your JSON Resume:\n\n```json\n" + `{
  "basics": {
    "name": "John Doe",
    "label": "Software Engineer",
    "email": "john@example.com",
    "phone": "+1-555-1234",
    "summary": "Experienced software engineer"
  },
  "work": [
    {
      "name": "Tech Corp",
      "position": "Senior Developer",
      "startDate": "2020-01-01"
    }
  ],
  "education": [
    {
      "institution": "State University",
      "area": "Computer Science",
      "studyType": "Bachelor"
    }
  ]
}` + "\n```\n\nThis follows the JSON Resume schema.",
			wantJSON: `{
  "basics": {
    "name": "John Doe",
    "label": "Software Engineer",
    "email": "john@example.com",
    "phone": "+1-555-1234",
    "summary": "Experienced software engineer"
  },
  "work": [
    {
      "name": "Tech Corp",
      "position": "Senior Developer",
      "startDate": "2020-01-01"
    }
  ],
  "education": [
    {
      "institution": "State University",
      "area": "Computer Science",
      "studyType": "Bachelor"
    }
  ]
}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractJSON([]byte(tt.input))

			if tt.wantErr {
				if err == nil {
					t.Errorf("ExtractJSON() error = nil, wantErr = true")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ExtractJSON() error = %q, want error containing %q", err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("ExtractJSON() error = %v, wantErr = false", err)
				return
			}

			// Compare as JSON to ignore whitespace differences
			var gotParsed, wantParsed interface{}
			if err := json.Unmarshal(got, &gotParsed); err != nil {
				t.Errorf("ExtractJSON() returned invalid JSON: %v", err)
				return
			}
			if err := json.Unmarshal([]byte(tt.wantJSON), &wantParsed); err != nil {
				t.Fatalf("test wantJSON is invalid: %v", err)
			}

			gotBytes, _ := json.Marshal(gotParsed)
			wantBytes, _ := json.Marshal(wantParsed)
			if string(gotBytes) != string(wantBytes) {
				t.Errorf("ExtractJSON() = %s, want %s", string(got), tt.wantJSON)
			}
		})
	}
}

func TestStripMarkdownFences(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no fences",
			input: `{"plain": "json"}`,
			want:  `{"plain": "json"}`,
		},
		{
			name:  "json fences",
			input: "```json\n{\"inside\": \"fences\"}\n```",
			want:  `{"inside": "fences"}`,
		},
		{
			name:  "plain fences",
			input: "```\n{\"plain\": \"fences\"}\n```",
			want:  `{"plain": "fences"}`,
		},
		{
			name:  "fences without newlines",
			input: "```json{\"compact\": true}```",
			want:  `{"compact": true}`,
		},
		{
			name:  "multiline content",
			input: "```json\n{\n  \"multi\": \"line\"\n}\n```",
			want:  "{\n  \"multi\": \"line\"\n}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripMarkdownFences([]byte(tt.input))
			if string(got) != tt.want {
				t.Errorf("stripMarkdownFences() = %q, want %q", string(got), tt.want)
			}
		})
	}
}

func TestTruncateForError(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{
			name:   "short input",
			input:  "short",
			maxLen: 10,
			want:   "short",
		},
		{
			name:   "exact length",
			input:  "exact",
			maxLen: 5,
			want:   "exact",
		},
		{
			name:   "truncated",
			input:  "this is a long string",
			maxLen: 10,
			want:   "this is a ...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateForError([]byte(tt.input), tt.maxLen)
			if got != tt.want {
				t.Errorf("truncateForError() = %q, want %q", got, tt.want)
			}
		})
	}
}
