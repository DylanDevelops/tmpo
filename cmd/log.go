package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/DylanDevelops/tmpo/internal/storage"
	"github.com/DylanDevelops/tmpo/internal/ui"
	"github.com/spf13/cobra"
)

var (
	logLimit int
	logProject string
	logToday bool
	logWeek bool
)

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "View time tracking history",
	Long:  `Display past time tracking entries with optional filtering.`,
	Run: func(cmd *cobra.Command, args []string) {
		ui.NewlineAbove()

		db, err := storage.Initialize()

		if err != nil {
			ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
			os.Exit(1)
		}

		defer db.Close()

		var entries []*storage.TimeEntry

		if logToday {
			start := time.Now().Truncate(24 * time.Hour)
			end := start.Add(24 * time.Hour)
			entries, err = db.GetEntriesByDateRange(start, end)
		} else if logWeek {
			now := time.Now()
			weekday := int(now.Weekday())
			if weekday == 0 {
				weekday = 7 // sunday
			}

			start := now.AddDate(0, 0, -weekday+1).Truncate(24 * time.Hour)
			end := start.AddDate(0, 0, 7)
			entries, err = db.GetEntriesByDateRange(start, end)
		} else if logProject != "" {
			entries, err = db.GetEntriesByProject(logProject)
		} else {
			entries, err = db.GetEntries(logLimit)
		}

		if err != nil {
			ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
			os.Exit(1)
		}

		if len(entries) == 0 {
			ui.PrintWarning(ui.EmojiWarning, "No time entries found.")
			ui.NewlineBelow()
			return
		}

		ui.PrintSuccess(ui.EmojiLog, fmt.Sprintf("Time Entries (%d total)", len(entries)))
		fmt.Println()

		var totalDuration time.Duration
		currentDate := ""

		for _, entry := range entries {
			entryDate := entry.StartTime.Format("Mon, Jan 2, 2006")
			if entryDate != currentDate {
				if currentDate != "" {
					fmt.Println()
				}

				fmt.Println(ui.Muted(fmt.Sprintf("─── %s ───", entryDate)))
				currentDate = entryDate
			}

			duration := entry.Duration()
			totalDuration += duration

			timeRange := entry.StartTime.Format("03:04 PM") + " - "
			if entry.EndTime != nil {
				timeRange += entry.EndTime.Format("03:04 PM") + "  "
			} else {
				timeRange += ui.Warning("(running)") + " "
			}

			fmt.Printf("  %s  %-20s  %s\n", timeRange, entry.ProjectName, ui.FormatDuration(duration))
			if entry.Description != "" {
				fmt.Printf("    %s %s\n", ui.Muted("└─"), entry.Description)
			}
		}

		fmt.Println()
		ui.PrintSeparator()
		fmt.Printf("%s %s\n", ui.Info("Total Time:"), ui.FormatDuration(totalDuration))

		ui.NewlineBelow()
	},
}

func init() {
	rootCmd.AddCommand(logCmd)

	logCmd.Flags().IntVarP(&logLimit, "limit", "l", 10, "Number of entries to show")
	logCmd.Flags().StringVarP(&logProject, "project", "p", "", "Filter by project name")
	logCmd.Flags().BoolVarP(&logToday, "today", "t", false, "Show today's entries")
	logCmd.Flags().BoolVarP(&logWeek, "week", "w", false, "Show this week's entries")
}
