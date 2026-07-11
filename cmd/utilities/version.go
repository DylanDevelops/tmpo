package utilities

import (
	"fmt"
	"os"
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

var releaseVersionRegex = regexp.MustCompile(`^v?\d+\.\d+\.\d+(-[\w.]+)?$`)

func VersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "version",
		Short:  "Show version information",
		Long:   "Display the current version information including date and release URL.",
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			DisplayVersionWithUpdateCheck()
		},
	}

	return cmd
}

func DisplayVersionWithUpdateCheck() {
	fmt.Println(GetVersionOutput())
	checkForUpdates()
	ui.NewlineBelow()
}

func GetVersionOutput() string {
	versionLine := fmt.Sprintf("tmpo version %s %s", ui.Success(Version), ui.Muted(GetFormattedDate(Date)))
	changelogLine := ui.Muted(GetChangelogUrl(Version))
	return fmt.Sprintf("\n%s\n%s", versionLine, changelogLine)
}

func GetFormattedDate(inputDate string) string {
	date, err := time.Parse(time.RFC3339, inputDate)
	if err != nil {
		return ""
	}

	return fmt.Sprintf("(%s)", date.Format("01-02-2006"))
}

func GetChangelogUrl(version string) string {
	path := "https://github.com/DylanDevelops/tmpo"

	if !isReleaseVersion(version) {
		return fmt.Sprintf("%s/releases/latest", path)
	}

	return fmt.Sprintf("%s/releases/tag/v%s", path, strings.TrimPrefix(version, "v"))
}

func checkForUpdates() {
	if isUpdateCheckDisabled() {
		return
	}

	if !isReleaseVersion(Version) {
		return
	}

	updateInfo, err := update.CheckForUpdate(Version)
	if err != nil {
		return
	}

	if updateInfo.HasUpdate {
		fmt.Printf("\n%s %s\n", ui.Info("New Update Available:"), ui.Bold(strings.TrimPrefix(updateInfo.LatestVersion, "v")))
		fmt.Printf("%s", ui.Muted(updateInfo.UpdateURL))
		ui.NewlineBelow()
	}
}

func isReleaseVersion(version string) bool {
	return releaseVersionRegex.MatchString(version)
}

func isUpdateCheckDisabled() bool {
	switch strings.ToLower(os.Getenv("TMPO_NO_UPDATE_CHECK")) {
	case "1", "true", "yes":
		return true
	default:
		return false
	}
}
