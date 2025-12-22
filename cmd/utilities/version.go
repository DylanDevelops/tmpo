package utilities

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/DylanDevelops/tmpo/internal/ui"
	"github.com/DylanDevelops/tmpo/internal/update"
	"github.com/spf13/cobra"
)

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

func VersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  "Display the current version information including date and release URL.",
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			DisplayVersionWithUpdateCheck()
		},
	}

	return cmd
}

// DisplayVersionWithUpdateCheck displays the version information and checks for updates.
// This is the single source of truth for displaying version info with update notifications.
func DisplayVersionWithUpdateCheck() {
	fmt.Print(GetVersionOutput())
	checkForUpdates()
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

// checkForUpdates checks if a newer version is available and displays a message if so.
// It silently fails if there's no internet connection or if the check fails.
func checkForUpdates() {
	// Only check if we have a valid version (not "dev" or empty)
	if Version == "" || Version == "dev" {
		return
	}

	updateInfo, err := update.CheckForUpdate(Version)
	if err != nil {
		// Silently fail and don't bother the user with network errors
		return
	}

	if updateInfo.HasUpdate {
		fmt.Printf("%s %s\n", ui.Info("New Update Available:"), ui.Bold(strings.TrimPrefix(updateInfo.LatestVersion, "v")))
		fmt.Printf("%s\n\n", ui.Muted(updateInfo.UpdateURL))
	}
}
