package extractor

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/richq/m2cv/internal/executor"
)

func TestSanitizeFilename(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "company name with spaces",
			input: "Google Inc",
			want:  "google-inc",
		},
		{
			name:  "role with spaces",
			input: "Software Engineer",
			want:  "software-engineer",
		},
		{
			name:  "company with forward slash",
			input: "Google/Software Engineer",
			want:  "google-software-engineer",
		},
		{
			name:  "multiple spaces collapse to single hyphen",
			input: "  spaced  out  ",
			want:  "spaced-out",
		},
		{
			name:  "special characters removed",
			input: "Special $#@! chars",
			want:  "special-chars",
		},
		{
			name:  "unicode letters preserved",
			input: "Uber Technologies",
			want:  "uber-technologies",
		},
		{
			name:  "underscores preserved",
			input: "my_company_name",
			want:  "my_company_name",
		},
		{
			name:  "backslash to hyphen",
			input: "Company\\Role",
			want:  "company-role",
		},
		{
			name:  "numbers preserved",
			input: "Web3 Company 2024",
			want:  "web3-company-2024",
		},
		{
			name:  "leading and trailing hyphens trimmed",
			input: "---trimmed---",
			want:  "trimmed",
		},
		{
			name:  "empty string stays empty",
			input: "   ",
			want:  "",
		},
		{
			name:  "only special chars becomes empty",
			input: "$#@!",
			want:  "",
		},
		{
			name:  "very long name truncated to 50",
			input: "this-is-a-very-long-company-name-that-exceeds-the-maximum-allowed-length-for-folder-names",
			want:  "this-is-a-very-long-company-name-that-exceeds-the",
		},
		{
			name:  "long name truncates at hyphen boundary",
			input: strings.Repeat("a", 45) + "-toolong",
			want:  strings.Repeat("a", 45),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := SanitizeFilename(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeFilename(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSanitizeFilename_MaxLength(t *testing.T) {
	t.Parallel()

	// Generate a 100 character input
	longInput := strings.Repeat("a", 100)
	got := SanitizeFilename(longInput)

	if len(got) > 50 {
		t.Errorf("SanitizeFilename() returned length %d, want <= 50", len(got))
	}
}

// mockExecutor implements executor.ClaudeExecutor for testing.
type mockExecutor struct {
	response string
	err      error
}

func (m *mockExecutor) Execute(ctx context.Context, prompt string, opts ...executor.ExecuteOption) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.response, nil
}

func TestExtractFolderName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		executor *mockExecutor
		jobDesc  string
		want     string
		wantErr  bool
	}{
		{
			name: "valid extraction returns sanitized name",
			executor: &mockExecutor{
				response: "Google-Software-Engineer\n",
			},
			jobDesc: "Software Engineer at Google...",
			want:    "google-software-engineer",
			wantErr: false,
		},
		{
			name: "response with extra whitespace",
			executor: &mockExecutor{
				response: "  stripe-backend-engineer  \n",
			},
			jobDesc: "Backend Engineer at Stripe",
			want:    "stripe-backend-engineer",
			wantErr: false,
		},
		{
			name: "empty Claude response returns error",
			executor: &mockExecutor{
				response: "   \n",
			},
			jobDesc: "Some job",
			want:    "",
			wantErr: true,
		},
		{
			name: "Claude execution error propagates",
			executor: &mockExecutor{
				err: errors.New("claude not available"),
			},
			jobDesc: "Some job",
			want:    "",
			wantErr: true,
		},
		{
			name: "response with special chars sanitized",
			executor: &mockExecutor{
				response: "Meta (Facebook) / ML Engineer",
			},
			jobDesc: "ML Engineer at Meta",
			want:    "meta-facebook-ml-engineer",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			got, err := ExtractFolderName(ctx, tt.executor, tt.jobDesc)

			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractFolderName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("ExtractFolderName() = %q, want %q", got, tt.want)
			}
		})
	}
}
