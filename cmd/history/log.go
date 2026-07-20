package history

import (
	"fmt"
	"slices"
	"time"

	"github.com/DylanDevelops/tmpo/internal/project"
	"github.com/DylanDevelops/tmpo/internal/settings"
	"github.com/DylanDevelops/tmpo/internal/storage"
	"github.com/DylanDevelops/tmpo/internal/ui"
	"github.com/spf13/cobra"
)

var (
	logLimit     int
	logProject   string
	logMilestone string
	logToday     bool
	logWeek      bool
	logDate      string
)

func LogCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "log",
		Short: "View time tracking history",
		Long:  `Display past time tracking entries with optional filtering.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ui.NewlineAbove()

			db, err := storage.Initialize()

			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				return err
			}

			defer db.Close()

			var entries []*storage.TimeEntry

			if logMilestone != "" {
				// if --project flag is used, ensure global project config is used
				projectName := logProject
				if projectName == "" {
					detectedProject, err := project.DetectConfiguredProject()
					if err != nil {
						ui.PrintError(ui.EmojiError, fmt.Sprintf("detecting project: %v", err))
						return err
					}
					projectName = detectedProject
				}
				entries, err = db.GetEntriesByMilestone(projectName, logMilestone)
			} else if logDate != "" {
				var parsedDate time.Time
				parsedDate, err = parseDateFlag(logDate)
				if err != nil {
					ui.PrintError(ui.EmojiError, err.Error())
					ui.NewlineBelow()
					return ui.ErrHandled
				}
				start := time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 0, 0, 0, 0, time.Local)
				end := start.Add(24 * time.Hour)
				entries, err = db.GetEntriesByDateRange(start, end)
			} else if logToday {
				year, month, day := time.Now().Year(), time.Now().Month(), time.Now().Day()
				start := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
				end := time.Now()
				entries, err = db.GetEntriesByDateRange(start, end)
			} else if logWeek {
				now := time.Now()
				weekday := int(now.Weekday())
				if weekday == 0 {
					weekday = 7 // sunday
				}

				startDay := now.AddDate(0, 0, -weekday+1)
				start := time.Date(startDay.Year(), startDay.Month(), startDay.Day(), 0, 0, 0, 0, time.Local)
				end := time.Now()
				entries, err = db.GetEntriesByDateRange(start, end)
			} else if logProject != "" {
				entries, err = db.GetEntriesByProject(logProject)
			} else {
				entries, err = db.GetEntries(logLimit)
			}

			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				return err
			}

			if len(entries) == 0 {
				ui.PrintWarning(ui.EmojiWarning, "No time entries found.")
				ui.NewlineBelow()
				return nil
			}

			ui.PrintSuccess(ui.EmojiLog, fmt.Sprintf("Time Entries (%d total)", len(entries)))
			fmt.Println()

			var totalDuration time.Duration
			currentDate := ""

			for _, entry := range slices.Backward(entries) {
				entryDate := settings.FormatDateLong(entry.StartTime)
				if entryDate != currentDate {
					if currentDate != "" {
						fmt.Println()
					}

					fmt.Println(ui.Bold(ui.Muted(fmt.Sprintf("─── %s ───", entryDate))))
					currentDate = entryDate
				}

				duration := entry.Duration()
				totalDuration += duration

				timeRange := settings.FormatTimePadded(entry.StartTime) + " - "
				if entry.EndTime != nil {
					timeRange += settings.FormatTimePadded(*entry.EndTime) + "  "
				} else {
					timeRange += ui.Warning("(running)") + " "
				}

				fmt.Printf("  %s  %s  %s\n", timeRange, ui.Bold(fmt.Sprintf("%-20s", entry.ProjectName)), ui.FormatDuration(duration))
				if entry.MilestoneName != nil {
					symbol := "└─"
					if entry.Description != "" {
						symbol = "├─"
					}
					fmt.Printf("    %s %s %s\n", ui.Muted(symbol), ui.Muted("Milestone:"), *entry.MilestoneName)
				}
				if entry.Description != "" {
					fmt.Printf("    %s %s\n", ui.Muted("└─"), entry.Description)
				}
			}

			fmt.Println()
			ui.PrintSeparator()
			fmt.Printf("%s %s\n", ui.BoldInfo("Total Time:"), ui.Bold(ui.FormatDuration(totalDuration)))

			ui.NewlineBelow()

			return nil
		},
	}

	cmd.Flags().IntVarP(&logLimit, "limit", "l", 10, "Number of entries to show")
	cmd.Flags().StringVarP(&logProject, "project", "p", "", "Filter by project name")
	cmd.Flags().StringVarP(&logMilestone, "milestone", "m", "", "Filter by milestone")
	cmd.Flags().BoolVarP(&logToday, "today", "t", false, "Show today's entries")
	cmd.Flags().BoolVarP(&logWeek, "week", "w", false, "Show this week's entries")
	cmd.Flags().StringVarP(&logDate, "date", "d", "", "Show entries for a specific date")

	return cmd
}

func parseDateFlag(dateStr string) (time.Time, error) {
	globalCfg, err := settings.LoadGlobalConfig()
	if err != nil {
		return time.Time{}, fmt.Errorf("loading config: %w", err)
	}

	layout := "2006-01-02"
	displayFormat := "YYYY-MM-DD"

	switch globalCfg.DateFormat {
	case "MM/DD/YYYY":
		layout = "01-02-2006"
		displayFormat = "MM-DD-YYYY"
	case "DD/MM/YYYY":
		layout = "02-01-2006"
		displayFormat = "DD-MM-YYYY"
	}

	parsedDate, err := time.ParseInLocation(layout, dateStr, time.Local)
	if err == nil {
		return parsedDate, nil
	}

	if layout != "2006-01-02" {
		parsedDate, err = time.ParseInLocation("2006-01-02", dateStr, time.Local)
		if err == nil {
			return parsedDate, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid date format. Please use %s or YYYY-MM-DD", displayFormat)
}
