package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/DylanDevelops/tmpo/internal/export"
	"github.com/DylanDevelops/tmpo/internal/storage"
	"github.com/DylanDevelops/tmpo/internal/ui"
	"github.com/spf13/cobra"
)

var (
	exportFormat string
	exportOutput string
	exportProject string
	exportToday bool
	exportWeek bool
)

var exportCmd = &cobra.Command{
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

		if exportToday {
			start := time.Now().Truncate(24 * time.Hour)
			end := start.Add(24 * time.Hour)
			entries, err = db.GetEntriesByDateRange(start, end)
		} else if exportWeek {
			now := time.Now()
			weekday := int(now.Weekday())
			if weekday == 0 {
				weekday = 7 // sunday
			}

			start := now.AddDate(0, 0, -weekday+1).Truncate(24 * time.Hour)
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

		switch exportFormat {
		case "csv":
			err = export.ToCSV(entries, filename)
		case "json":
			err = export.ToJson(entries, filename)
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

func init() {
	rootCmd.AddCommand(exportCmd)

	exportCmd.Flags().StringVarP(&exportFormat, "format", "f", "csv", "Export format (csv or json)")
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "Output filename")
	exportCmd.Flags().StringVarP(&exportProject, "project", "p", "", "Filter by project")
	exportCmd.Flags().BoolVarP(&exportToday, "today", "t", false, "Export today's entries")
	exportCmd.Flags().BoolVarP(&exportWeek, "week", "w", false, "Export this week's entries")
}
