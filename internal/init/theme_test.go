package init

import (
	"testing"
)

func TestIsValidTheme_ValidThemes(t *testing.T) {
	// All themes in AvailableThemes should pass validation
	for _, theme := range AvailableThemes {
		if !IsValidTheme(theme) {
			t.Errorf("IsValidTheme(%q) = false, want true", theme)
		}
	}
}

func TestIsValidTheme_InvalidTheme(t *testing.T) {
	invalidThemes := []string{
		"invalid",
		"nonexistent",
		"foobar",
		"",
		"EVEN", // case-sensitive
	}

	for _, theme := range invalidThemes {
		if IsValidTheme(theme) {
			t.Errorf("IsValidTheme(%q) = true, want false", theme)
		}
	}
}

func TestThemePackageName(t *testing.T) {
	tests := []struct {
		theme    string
		expected string
	}{
		{"even", "jsonresume-theme-even"},
		{"stackoverflow", "jsonresume-theme-stackoverflow"},
		{"elegant", "jsonresume-theme-elegant"},
		{"actual", "jsonresume-theme-actual"},
		{"class", "jsonresume-theme-class"},
		{"flat", "jsonresume-theme-flat"},
		{"kendall", "jsonresume-theme-kendall"},
		{"macchiato", "jsonresume-theme-macchiato"},
	}

	for _, tt := range tests {
		result := ThemePackageName(tt.theme)
		if result != tt.expected {
			t.Errorf("ThemePackageName(%q) = %q, want %q", tt.theme, result, tt.expected)
		}
	}
}

func TestAvailableThemes_ContainsExpectedThemes(t *testing.T) {
	expectedThemes := []string{
		"even",
		"stackoverflow",
		"elegant",
		"actual",
		"class",
		"flat",
		"kendall",
		"macchiato",
	}

	if len(AvailableThemes) != len(expectedThemes) {
		t.Errorf("AvailableThemes has %d themes, want %d", len(AvailableThemes), len(expectedThemes))
	}

	for _, theme := range expectedThemes {
		if !IsValidTheme(theme) {
			t.Errorf("Expected theme %q not in AvailableThemes", theme)
		}
	}
}

func TestThemeDescriptions_AllThemesHaveDescriptions(t *testing.T) {
	for _, theme := range AvailableThemes {
		desc, exists := ThemeDescriptions[theme]
		if !exists {
			t.Errorf("Theme %q missing from ThemeDescriptions", theme)
			continue
		}
		if desc == "" {
			t.Errorf("Theme %q has empty description", theme)
		}
	}
}

// TestSelectTheme is skipped because it requires an interactive terminal.
// The function uses charmbracelet/huh which needs stdin to be a tty.
func TestSelectTheme(t *testing.T) {
	t.Skip("SelectTheme requires interactive terminal - skipping in automated tests")
}
