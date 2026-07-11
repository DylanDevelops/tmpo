package utilities

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsUpdateCheckDisabled(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		setEnv   bool
		expected bool
	}{
		{
			name:     "unset variable is enabled",
			setEnv:   false,
			expected: false,
		},
		{
			name:     "empty value is enabled",
			envValue: "",
			setEnv:   true,
			expected: false,
		},
		{
			name:     "1 disables",
			envValue: "1",
			setEnv:   true,
			expected: true,
		},
		{
			name:     "true disables",
			envValue: "true",
			setEnv:   true,
			expected: true,
		},
		{
			name:     "yes disables",
			envValue: "yes",
			setEnv:   true,
			expected: true,
		},
		{
			name:     "uppercase TRUE disables",
			envValue: "TRUE",
			setEnv:   true,
			expected: true,
		},
		{
			name:     "0 does not disable",
			envValue: "0",
			setEnv:   true,
			expected: false,
		},
		{
			name:     "arbitrary value does not disable",
			envValue: "maybe",
			setEnv:   true,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				t.Setenv("TMPO_NO_UPDATE_CHECK", tt.envValue)
			} else {
				// Ensure a deterministic "unset" state regardless of the
				// ambient environment. t.Setenv registers cleanup that
				// restores the original value after the subtest.
				t.Setenv("TMPO_NO_UPDATE_CHECK", "placeholder")
				os.Unsetenv("TMPO_NO_UPDATE_CHECK")
			}
			result := isUpdateCheckDisabled()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetFormattedDate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid RFC3339 date",
			input:    "2024-01-15T10:30:00Z",
			expected: "(01-15-2024)",
		},
		{
			name:     "valid RFC3339 date with timezone",
			input:    "2024-12-25T15:45:30-05:00",
			expected: "(12-25-2024)",
		},
		{
			name:     "empty string returns empty",
			input:    "",
			expected: "",
		},
		{
			name:     "invalid date format returns empty",
			input:    "2024-01-15",
			expected: "",
		},
		{
			name:     "invalid date string returns empty",
			input:    "not a date",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetFormattedDate(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetChangelogUrl(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{
			name:     "valid version without v prefix",
			version:  "1.0.0",
			expected: "https://github.com/DylanDevelops/tmpo/releases/tag/v1.0.0",
		},
		{
			name:     "valid version with v prefix",
			version:  "v2.3.4",
			expected: "https://github.com/DylanDevelops/tmpo/releases/tag/v2.3.4",
		},
		{
			name:     "version with prerelease tag",
			version:  "1.0.0-beta.1",
			expected: "https://github.com/DylanDevelops/tmpo/releases/tag/v1.0.0-beta.1",
		},
		{
			name:     "version with v prefix and prerelease",
			version:  "v1.0.0-rc.2",
			expected: "https://github.com/DylanDevelops/tmpo/releases/tag/v1.0.0-rc.2",
		},
		{
			name:     "dev version returns latest",
			version:  "dev",
			expected: "https://github.com/DylanDevelops/tmpo/releases/latest",
		},
		{
			name:     "unstable version returns latest",
			version:  "unstable",
			expected: "https://github.com/DylanDevelops/tmpo/releases/latest",
		},
		{
			name:     "empty version returns latest",
			version:  "",
			expected: "https://github.com/DylanDevelops/tmpo/releases/latest",
		},
		{
			name:     "invalid version format returns latest",
			version:  "1.0",
			expected: "https://github.com/DylanDevelops/tmpo/releases/latest",
		},
		{
			name:     "invalid version string returns latest",
			version:  "not-a-version",
			expected: "https://github.com/DylanDevelops/tmpo/releases/latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetChangelogUrl(tt.version)
			assert.Equal(t, tt.expected, result)
		})
	}
}
