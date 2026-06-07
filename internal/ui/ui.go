package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/DylanDevelops/tmpo/internal/shell"
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
	ResetBoldDim   = "\033[22m" // Reset bold and dim
	ResetItalic    = "\033[23m" // Reset italic
	ResetUnderline = "\033[24m" // Reset underline
)

// Emoji Constants
const (
	EmojiStart     = "✨"
	EmojiStop      = "🛑"
	EmojiStatus    = "⏱️"
	EmojiStats     = "📊"
	EmojiLog       = "📝"
	EmojiManual    = "✍️"
	EmojiInit      = "⚙️"
	EmojiExport    = "📤"
	EmojiMilestone = "🎯"
	EmojiBackup    = "💾"
	EmojiSuccess   = "✅"
	EmojiError     = "❌"
	EmojiWarning   = "⚠️"
	EmojiInfo      = "ℹ️"
	EmojiUndo      = "↩️"
)

func Success(message string) string {
	return ColorGreen + message + ColorReset
}

func Error(message string) string {
	return ColorRed + message + ColorReset
}

func Info(message string) string {
	return ColorBlue + message + ColorReset
}

func Warning(message string) string {
	return ColorYellow + message + ColorReset
}

func Muted(message string) string {
	return ColorGray + message + ColorReset
}

func Bold(message string) string {
	return FormatBold + message + ResetBoldDim
}

func Dim(message string) string {
	return FormatDim + message + ResetBoldDim
}

func Italic(message string) string {
	return FormatItalic + message + ResetItalic
}

func Underline(message string) string {
	return FormatUnderline + message + ResetUnderline
}

func BoldSuccess(message string) string {
	return FormatBold + ColorGreen + message + ColorReset
}

func BoldError(message string) string {
	return FormatBold + ColorRed + message + ColorReset
}

func BoldInfo(message string) string {
	return FormatBold + ColorBlue + message + ColorReset
}

func BoldWarning(message string) string {
	return FormatBold + ColorYellow + message + ColorReset
}

func PrintSuccess(emoji, message string) {
	fmt.Println(Success(fmt.Sprintf("%s  %s", emoji, message)))
}

func PrintError(emoji, message string) {
	fmt.Fprintf(os.Stderr, "%s\n", Error(fmt.Sprintf("%s  %s", emoji, message)))
}

func PrintWarning(emoji, message string) {
	fmt.Println(Warning(fmt.Sprintf("%s  %s", emoji, message)))
}

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

func PrintMuted(indent int, message string) {
	spaces := ""
	for i := 0; i < indent; i++ {
		spaces += " "
	}
	fmt.Printf("%s%s\n", spaces, Muted(message))
}

func PrintSeparator() {
	fmt.Println(Muted("─────────────────────────────────────────"))
}

func NewlineAbove() {
	fmt.Println()
}

func NewlineBelow() {
	if !shell.IsLegacyWindowsCMD() {
		fmt.Println()
	}
}

func FormatFileSize(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	}
	return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
}

func SanitizeSingleLine(s string) string {
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.TrimSpace(s)
}

func FormatDuration(d time.Duration) string {
	totalSeconds := int(d.Seconds())
	days := totalSeconds / 86400
	hours := (totalSeconds % 86400) / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}

	return fmt.Sprintf("%ds", seconds)
}
