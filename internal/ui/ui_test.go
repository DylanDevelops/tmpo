package ui

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestColorFunctions(t *testing.T) {
	t.Run("Success adds green color", func(t *testing.T) {
		result := Success("test")
		assert.Contains(t, result, ColorGreen)
		assert.Contains(t, result, "test")
		assert.Contains(t, result, ColorReset)
	})

	t.Run("Error adds red color", func(t *testing.T) {
		result := Error("test")
		assert.Contains(t, result, ColorRed)
		assert.Contains(t, result, "test")
		assert.Contains(t, result, ColorReset)
	})

	t.Run("Info adds blue color", func(t *testing.T) {
		result := Info("test")
		assert.Contains(t, result, ColorBlue)
		assert.Contains(t, result, "test")
		assert.Contains(t, result, ColorReset)
	})

	t.Run("Warning adds yellow color", func(t *testing.T) {
		result := Warning("test")
		assert.Contains(t, result, ColorYellow)
		assert.Contains(t, result, "test")
		assert.Contains(t, result, ColorReset)
	})

	t.Run("Muted adds gray color", func(t *testing.T) {
		result := Muted("test")
		assert.Contains(t, result, ColorGray)
		assert.Contains(t, result, "test")
		assert.Contains(t, result, ColorReset)
	})
}

func TestFormattingFunctions(t *testing.T) {
	t.Run("Bold adds bold formatting", func(t *testing.T) {
		result := Bold("test")
		assert.Contains(t, result, FormatBold)
		assert.Contains(t, result, "test")
		assert.Contains(t, result, ColorReset)
	})

	t.Run("Dim adds dim formatting", func(t *testing.T) {
		result := Dim("test")
		assert.Contains(t, result, FormatDim)
		assert.Contains(t, result, "test")
		assert.Contains(t, result, ColorReset)
	})

	t.Run("Italic adds italic formatting", func(t *testing.T) {
		result := Italic("test")
		assert.Contains(t, result, FormatItalic)
		assert.Contains(t, result, "test")
		assert.Contains(t, result, ColorReset)
	})

	t.Run("Underline adds underline formatting", func(t *testing.T) {
		result := Underline("test")
		assert.Contains(t, result, FormatUnderline)
		assert.Contains(t, result, "test")
		assert.Contains(t, result, ColorReset)
	})
}

func TestCombinedFormattingFunctions(t *testing.T) {
	t.Run("BoldSuccess adds bold and green", func(t *testing.T) {
		result := BoldSuccess("test")
		assert.Contains(t, result, FormatBold)
		assert.Contains(t, result, ColorGreen)
		assert.Contains(t, result, "test")
		assert.Contains(t, result, ColorReset)
	})

	t.Run("BoldError adds bold and red", func(t *testing.T) {
		result := BoldError("test")
		assert.Contains(t, result, FormatBold)
		assert.Contains(t, result, ColorRed)
		assert.Contains(t, result, "test")
		assert.Contains(t, result, ColorReset)
	})

	t.Run("BoldInfo adds bold and blue", func(t *testing.T) {
		result := BoldInfo("test")
		assert.Contains(t, result, FormatBold)
		assert.Contains(t, result, ColorBlue)
		assert.Contains(t, result, "test")
		assert.Contains(t, result, ColorReset)
	})

	t.Run("BoldWarning adds bold and yellow", func(t *testing.T) {
		result := BoldWarning("test")
		assert.Contains(t, result, FormatBold)
		assert.Contains(t, result, ColorYellow)
		assert.Contains(t, result, "test")
		assert.Contains(t, result, ColorReset)
	})
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "seconds only",
			duration: 45 * time.Second,
			expected: "45s",
		},
		{
			name:     "minutes and seconds",
			duration: 5*time.Minute + 30*time.Second,
			expected: "5m 30s",
		},
		{
			name:     "hours, minutes, and seconds",
			duration: 2*time.Hour + 15*time.Minute + 45*time.Second,
			expected: "2h 15m 45s",
		},
		{
			name:     "exact hours",
			duration: 3 * time.Hour,
			expected: "3h 0m 0s",
		},
		{
			name:     "exact minutes",
			duration: 10 * time.Minute,
			expected: "10m 0s",
		},
		{
			name:     "zero duration",
			duration: 0,
			expected: "0s",
		},
		{
			name:     "one second",
			duration: 1 * time.Second,
			expected: "1s",
		},
		{
			name:     "large duration",
			duration: 25*time.Hour + 45*time.Minute + 30*time.Second,
			expected: "25h 45m 30s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDuration(tt.duration)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConstants(t *testing.T) {
	t.Run("Color constants are defined", func(t *testing.T) {
		assert.NotEmpty(t, ColorReset)
		assert.NotEmpty(t, ColorGreen)
		assert.NotEmpty(t, ColorRed)
		assert.NotEmpty(t, ColorBlue)
		assert.NotEmpty(t, ColorYellow)
		assert.NotEmpty(t, ColorCyan)
		assert.NotEmpty(t, ColorGray)
	})

	t.Run("Format constants are defined", func(t *testing.T) {
		assert.NotEmpty(t, FormatBold)
		assert.NotEmpty(t, FormatDim)
		assert.NotEmpty(t, FormatItalic)
		assert.NotEmpty(t, FormatUnderline)
	})

	t.Run("Emoji constants are defined", func(t *testing.T) {
		assert.NotEmpty(t, EmojiStart)
		assert.NotEmpty(t, EmojiStop)
		assert.NotEmpty(t, EmojiStatus)
		assert.NotEmpty(t, EmojiStats)
		assert.NotEmpty(t, EmojiLog)
		assert.NotEmpty(t, EmojiManual)
		assert.NotEmpty(t, EmojiInit)
		assert.NotEmpty(t, EmojiExport)
		assert.NotEmpty(t, EmojiSuccess)
		assert.NotEmpty(t, EmojiError)
		assert.NotEmpty(t, EmojiWarning)
		assert.NotEmpty(t, EmojiInfo)
	})
}
