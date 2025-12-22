package history

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

func LogCmd() *cobra.Command {
	cmd := &cobra.Command{
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

					fmt.Println(ui.Bold(ui.Muted(fmt.Sprintf("─── %s ───", entryDate))))
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

				fmt.Printf("  %s  %s  %s\n", timeRange, ui.Bold(fmt.Sprintf("%-20s", entry.ProjectName)), ui.FormatDuration(duration))
				if entry.Description != "" {
					fmt.Printf("    %s %s\n", ui.Muted("└─"), entry.Description)
				}
			}

			fmt.Println()
			ui.PrintSeparator()
			fmt.Printf("%s %s\n", ui.BoldInfo("Total Time:"), ui.Bold(ui.FormatDuration(totalDuration)))

			ui.NewlineBelow()
		},
	}

	cmd.Flags().IntVarP(&logLimit, "limit", "l", 10, "Number of entries to show")
	cmd.Flags().StringVarP(&logProject, "project", "p", "", "Filter by project name")
	cmd.Flags().BoolVarP(&logToday, "today", "t", false, "Show today's entries")
	cmd.Flags().BoolVarP(&logWeek, "week", "w", false, "Show this week's entries")

	return cmd
}
