package history

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/DylanDevelops/tmpo/internal/export"
	"github.com/DylanDevelops/tmpo/internal/project"
	"github.com/DylanDevelops/tmpo/internal/settings"
	"github.com/DylanDevelops/tmpo/internal/storage"
	"github.com/DylanDevelops/tmpo/internal/ui"
	"github.com/spf13/cobra"
)

var (
	exportFormat    string
	exportOutput    string
	exportProject   string
	exportMilestone string
	exportToday     bool
	exportWeek      bool
	exportUtc       bool
)

func ExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export time entries",
		Long:  `Export time tracking data to different formats.`,
		Run: func(cmd *cobra.Command, args []string) {
			ui.NewlineAbove()

			db, err := storage.Initialize()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			defer db.Close()

			var entries []*storage.TimeEntry

			if exportMilestone != "" {
				// if --project flag is used, ensure global project config is used
				projectName := exportProject
				if projectName == "" {
					detectedProject, err := project.DetectConfiguredProject()
					if err != nil {
						ui.PrintError(ui.EmojiError, fmt.Sprintf("detecting project: %v", err))
						os.Exit(1)
					}
					projectName = detectedProject
				}
				entries, err = db.GetEntriesByMilestone(projectName, exportMilestone)
			} else if exportToday {
				tz := settings.GetDisplayTimezone()
				if exportUtc {
					tz = time.UTC
				}

				now := time.Now().In(tz)
				start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, tz)
				end := start.Add(24 * time.Hour)

				entries, err = db.GetEntriesByDateRange(start, end)
			} else if exportWeek {
				tz := settings.GetDisplayTimezone()
				if exportUtc {
					tz = time.UTC
				}

				now := time.Now().In(tz)
				weekday := int(now.Weekday())
				if weekday == 0 {
					weekday = 7 // sunday
				}

				firstOfWeek := now.AddDate(0, 0, -weekday+1)
				start := time.Date(
					firstOfWeek.Year(),
					firstOfWeek.Month(),
					firstOfWeek.Day(),
					0, 0, 0, 0,
					tz)
				end := start.AddDate(0, 0, 7)

				entries, err = db.GetEntriesByDateRange(start, end)
			} else if exportProject != "" {
				entries, err = db.GetEntriesByProject(exportProject)
			} else {
				entries, err = db.GetEntries(0) // all
			}

			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			if len(entries) == 0 {
				ui.PrintWarning(ui.EmojiWarning, "No entries to export.")
				ui.NewlineBelow()
				os.Exit(0)
			}

			var exportPath string

			// try to load .tmporc config first
			if config, _, err := settings.FindAndLoad(); err == nil && config.ExportPath != "" {
				exportPath = config.ExportPath
			} else {
				// fall back to use global config
				if globalConfig, err := settings.LoadGlobalConfig(); err == nil && globalConfig.ExportPath != "" {
					exportPath = globalConfig.ExportPath
				}
			}

			if exportPath != "" {
				if len(exportPath) >= 2 && exportPath[:2] == "~/" {
					home, err := os.UserHomeDir()
					if err == nil {
						exportPath = filepath.Join(home, exportPath[2:])
					}
				} else if exportPath == "~" {
					home, err := os.UserHomeDir()
					if err == nil {
						exportPath = home
					}
				}

				// make sure that that the path is valid
				if err := os.MkdirAll(exportPath, 0755); err != nil {
					ui.PrintError(ui.EmojiError, fmt.Sprintf("Failed to create export directory: %v", err))
					os.Exit(1)
				}
			}

			filename := exportOutput
			if filename == "" {
				timestamp := time.Now().Format("2006-01-02")
				ext := "csv"

				if exportFormat == "json" {
					ext = "json"
				}

				filename = fmt.Sprintf("tmpo-export-%s.%s", timestamp, ext)
			}

			if exportFormat == "csv" && filepath.Ext(filename) != ".csv" {
				filename += ".csv"
			} else if exportFormat == "json" && filepath.Ext(filename) != ".json" {
				filename += ".json"
			}

			if exportPath != "" {
				// add export path to beginning of path
				filename = filepath.Join(exportPath, filepath.Base(filename))
			}

			switch exportFormat {
			case "csv":
				err = export.ToCSV(entries, filename, exportUtc)
			case "json":
				err = export.ToJson(entries, filename, exportUtc)
			default:
				ui.PrintError(ui.EmojiError, fmt.Sprintf("Unknown format '%s'. Use 'csv' or 'json'", exportFormat))
				os.Exit(1)
			}

			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			ui.PrintSuccess(ui.EmojiExport, fmt.Sprintf("Exported %s to %s", ui.Bold(fmt.Sprintf("%d entries", len(entries))), ui.Bold(filename)))

			ui.NewlineBelow()
		},
	}

	cmd.Flags().StringVarP(&exportFormat, "format", "f", "csv", "Export format (csv or json)")
	cmd.Flags().StringVarP(&exportOutput, "output", "o", "", "Output filename")
	cmd.Flags().StringVarP(&exportProject, "project", "p", "", "Filter by project")
	cmd.Flags().StringVarP(&exportMilestone, "milestone", "m", "", "Filter by milestone")
	cmd.Flags().BoolVarP(&exportToday, "today", "t", false, "Export today's entries")
	cmd.Flags().BoolVarP(&exportWeek, "week", "w", false, "Export this week's entries")
	cmd.Flags().BoolVar(&exportUtc, "utc", false, "Export timestamps in UTC rather than local timezone")

	return cmd
}
