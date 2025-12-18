package ui

import (
	"fmt"
	"os"
	"time"
)

// ANSI Color Constants
const (
	ColorReset  = "\033[0m"
	ColorGreen  = "\033[32m" // Success
	ColorRed    = "\033[31m" // Errors
	ColorBlue   = "\033[34m" // Info
	ColorYellow = "\033[33m" // Warnings
	ColorCyan   = "\033[36m" // Highlights
	ColorGray   = "\033[90m" // Muted text
)

// ANSI Text Formatting Constants
const (
	FormatBold      = "\033[1m"
	FormatDim       = "\033[2m"
	FormatItalic    = "\033[3m"
	FormatUnderline = "\033[4m"

	// Specific reset codes (don't reset colors)
	ResetBoldDim    = "\033[22m" // Reset bold and dim
	ResetItalic     = "\033[23m" // Reset italic
	ResetUnderline  = "\033[24m" // Reset underline
)

// Emoji Constants
const (
	EmojiStart   = "✨"
	EmojiStop    = "🛑"
	EmojiStatus  = "⏱️"
	EmojiStats   = "📊"
	EmojiLog     = "📝"
	EmojiManual  = "✍️"
	EmojiInit    = "⚙️"
	EmojiExport  = "📤"
	EmojiSuccess = "✅"
	EmojiError   = "❌"
	EmojiWarning = "⚠️"
	EmojiInfo    = "ℹ️"
)

// Success colored output functions that returns colored string
func Success(message string) string {
	return ColorGreen + message + ColorReset
}

// Error colored output functions that returns colored string
func Error(message string) string {
	return ColorRed + message + ColorReset
}

// Info colored output functions that returns colored string
func Info(message string) string {
	return ColorBlue + message + ColorReset
}

// Warning colored output functions that returns colored string
func Warning(message string) string {
	return ColorYellow + message + ColorReset
}

// Muted colored output functions that returns colored string
func Muted(message string) string {
	return ColorGray + message + ColorReset
}

// Bold text formatting functions that return formatted string
func Bold(message string) string {
	return FormatBold + message + ResetBoldDim
}

// Dim text formatting functions that return formatted string
func Dim(message string) string {
	return FormatDim + message + ResetBoldDim
}

// Italic text formatting functions that return formatted string
func Italic(message string) string {
	return FormatItalic + message + ResetItalic
}

// Underline text formatting functions that return formatted string
func Underline(message string) string {
	return FormatUnderline + message + ResetUnderline
}

// Bold success combined formatting functions for common use cases
func BoldSuccess(message string) string {
	return FormatBold + ColorGreen + message + ColorReset
}

// Bold error combined formatting functions for common use cases
func BoldError(message string) string {
	return FormatBold + ColorRed + message + ColorReset
}

// Bold info combined formatting functions for common use cases
func BoldInfo(message string) string {
	return FormatBold + ColorBlue + message + ColorReset
}

// Bold warning combined formatting functions for common use cases
func BoldWarning(message string) string {
	return FormatBold + ColorYellow + message + ColorReset
}

// PrintSuccess prints a success message with emoji and color to stdout
func PrintSuccess(emoji, message string) {
	fmt.Println(Success(fmt.Sprintf("%s  %s", emoji, message)))
}

// PrintError prints an error message with emoji and color to stderr
func PrintError(emoji, message string) {
	fmt.Fprintf(os.Stderr, "%s\n", Error(fmt.Sprintf("%s  %s", emoji, message)))
}

// PrintWarning prints a warning message with emoji and color to stdout
func PrintWarning(emoji, message string) {
	fmt.Println(Warning(fmt.Sprintf("%s  %s", emoji, message)))
}

// PrintInfo prints an info line with proper indentation and color
// indent specifies the number of spaces (typically 4 or 8)
// If value is empty, only label is printed
func PrintInfo(indent int, label, value string) {
	spaces := ""
	for i := 0; i < indent; i++ {
		spaces += " "
	}

	if value != "" {
		fmt.Printf("%s%s: %s\n", spaces, Info(label), value)
	} else {
		fmt.Printf("%s%s\n", spaces, Info(label))
	}
}

// PrintMuted prints muted (gray) text with optional indentation
func PrintMuted(indent int, message string) {
	spaces := ""
	for i := 0; i < indent; i++ {
		spaces += " "
	}
	fmt.Printf("%s%s\n", spaces, Muted(message))
}

// PrintSeparator prints a subtle horizontal separator line
func PrintSeparator() {
	fmt.Println(Muted("─────────────────────────────────────────"))
}

// NewlineAbove prints a single newline before output
// This creates visual separation from the user's command input
func NewlineAbove() {
	fmt.Println()
}

// NewlineBelow prints a single newline after output
func NewlineBelow() {
	fmt.Println()
}

// FormatDuration formats d into a concise, human-readable string using hours, minutes and seconds.
// It returns "<h>h <m>m <s>s" when the duration is at least one hour, "<m>m <s>s" when the duration
// is at least one minute but less than an hour, and "<s>s" for durations under one minute.
// Hours, minutes and seconds are derived from d using integer truncation (no fractional parts).
// This function is intended for non-negative durations; behavior for negative durations is unspecified.
func FormatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}

	return fmt.Sprintf("%ds", seconds)
}
