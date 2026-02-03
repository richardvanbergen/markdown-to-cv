// Package init provides initialization functionality for m2cv projects.
// It handles config file creation, npm package installation, and
// interactive theme selection.
package init

import (
	"slices"

	"github.com/charmbracelet/huh"
)

// AvailableThemes lists the supported JSON Resume themes.
// These are pre-validated to work with resumed and produce good PDF output.
var AvailableThemes = []string{
	"even",
	"stackoverflow",
	"elegant",
	"actual",
	"class",
	"flat",
	"kendall",
	"macchiato",
}

// ThemeDescriptions provides human-readable descriptions for themes.
var ThemeDescriptions = map[string]string{
	"even":          "Clean, minimal design - great for most industries",
	"stackoverflow": "Developer-focused with brand icons and skills sections",
	"elegant":       "Professional and polished - classic resume style",
	"actual":        "Minimalist and modern - contemporary design",
	"class":         "Self-contained, works offline - portable HTML/PDF",
	"flat":          "Simple flat design - straightforward layout",
	"kendall":       "Modern professional - balanced and readable",
	"macchiato":     "Warm tones, modern feel - distinctive look",
}

// SelectTheme presents an interactive theme selection prompt.
// Returns the selected theme name or an error if selection is cancelled.
func SelectTheme() (string, error) {
	var selected string

	// Build options from available themes
	options := make([]huh.Option[string], len(AvailableThemes))
	for i, theme := range AvailableThemes {
		desc := ThemeDescriptions[theme]
		if desc == "" {
			desc = theme
		}
		options[i] = huh.NewOption(desc, theme)
	}

	err := huh.NewSelect[string]().
		Title("Select a JSON Resume theme").
		Description("Theme determines the visual style of your PDF resume").
		Options(options...).
		Value(&selected).
		Run()

	if err != nil {
		return "", err
	}

	return selected, nil
}

// IsValidTheme checks if the theme name is in the available list.
func IsValidTheme(theme string) bool {
	return slices.Contains(AvailableThemes, theme)
}

// ThemePackageName returns the full npm package name for a theme.
func ThemePackageName(theme string) string {
	return "jsonresume-theme-" + theme
}
