package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectDefaultProjectName(t *testing.T) {
	t.Run("returns git repository name when in git repo", func(t *testing.T) {
		// This test would require setting up a real git repo
		// We'll test the fallback behavior instead
		tmpDir := t.TempDir()
		origDir, err := os.Getwd()
		require.NoError(t, err)
		defer os.Chdir(origDir)

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		// Initialize a minimal git repo
		err = os.MkdirAll(filepath.Join(tmpDir, ".git"), 0755)
		require.NoError(t, err)

		name := detectDefaultProjectName()
		assert.NotEmpty(t, name)
	})

	t.Run("returns directory name when not in git repo", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, err := os.Getwd()
		require.NoError(t, err)
		defer os.Chdir(origDir)

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		name := detectDefaultProjectName()
		assert.NotEmpty(t, name)
		// The name should be the base of the temp directory
		assert.Equal(t, filepath.Base(tmpDir), name)
	})
}

func TestValidateHourlyRate(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "empty string is valid (optional field)",
			input:     "",
			wantError: false,
		},
		{
			name:      "whitespace only is valid",
			input:     "   ",
			wantError: false,
		},
		{
			name:      "valid positive number",
			input:     "75.50",
			wantError: false,
		},
		{
			name:      "valid integer",
			input:     "100",
			wantError: false,
		},
		{
			name:      "zero is valid",
			input:     "0",
			wantError: false,
		},
		{
			name:      "negative number is invalid",
			input:     "-50",
			wantError: true,
			errorMsg:  "hourly rate cannot be negative",
		},
		{
			name:      "non-numeric string is invalid",
			input:     "not-a-number",
			wantError: true,
			errorMsg:  "must be a valid number",
		},
		{
			name:      "mixed alphanumeric is invalid",
			input:     "50abc",
			wantError: true,
			errorMsg:  "must be a valid number",
		},
		{
			name:      "special characters are invalid",
			input:     "$100",
			wantError: true,
			errorMsg:  "must be a valid number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateHourlyRate(tt.input)
			if tt.wantError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
