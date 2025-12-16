package cmd

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/DylanDevelops/tmpo/internal/ui"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  "Display the current version information including date and release URL.",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(GetVersionOutput())
	},
}

// GetVersionOutput returns the formatted version string used by both
// the version subcommand and the -v/--version flags
func GetVersionOutput() string {
	versionLine := fmt.Sprintf("tmpo version %s %s", ui.Success(Version), ui.Muted(GetFormattedDate(Date)))
	changelogLine := ui.Muted(GetChangelogUrl(Version))
	return fmt.Sprintf("\n%s\n%s\n\n", versionLine, changelogLine)
}

// GetFormattedDate parses inputDate as an RFC3339 timestamp and returns the date
// formatted as "MM-DD-YYYY" wrapped in parentheses (for example "(01-02-2006)").
// If inputDate is empty or cannot be parsed as RFC3339, it returns an empty string.
func GetFormattedDate(inputDate string) string {
	if inputDate == "" {
		return ""
	}

	date, err := time.Parse(time.RFC3339, inputDate)
	if err != nil {
		return ""
	}

	return fmt.Sprintf("(%s)", date.Format("01-02-2006"))
}

func GetChangelogUrl(version string) string {
	path := "https://github.com/DylanDevelops/tmpo"

	r := regexp.MustCompile(`^v?\d+\.\d+\.\d+(-[\w.]+)?$`)
	if !r.MatchString(version) {
		return fmt.Sprintf("%s/releases/latest", path)
	}

	return fmt.Sprintf("%s/releases/tag/v%s", path, strings.TrimPrefix(version, "v"))
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
