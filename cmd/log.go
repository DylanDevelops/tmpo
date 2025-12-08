package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/DylanDevelops/tmpo/internal/storage"
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
		db, err := storage.Initialize()

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
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
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(entries) == 0 {
			fmt.Println("No time entries found.")
			
			return
		}

		fmt.Printf("\n[tmpo] Time Entries (%d total)\n\n", len(entries))

		var totalDuration time.Duration
		currentDate := ""

		for _, entry := range entries {
			entryDate := entry.StartTime.Format("Mon, Jan 2, 2006")
			if entryDate != currentDate {
				if currentDate != "" {
					fmt.Println()
				}

				fmt.Printf("─── %s ───\n", entryDate)
				currentDate = entryDate
			}

			duration := entry.Duration()
			totalDuration += duration

			timeRange := entry.StartTime.Format("3:04 PM")
			if entry.EndTime != nil {
				timeRange += " - " + entry.EndTime.Format("3:04 PM")
			} else {
				timeRange += " - (running)"
			}

			fmt.Printf("  %s  %-20s  %s\n", timeRange, entry.ProjectName, formatDuration(duration))
			if entry.Description != "" {
				fmt.Printf("    └─ %s\n", entry.Description)
			}
		}

		fmt.Printf("\n─────────────────────────────────────────\n")
		fmt.Printf("Total Time: %s\n", formatDuration(totalDuration))
	},
}

func init() {
	rootCmd.AddCommand(logCmd)

	logCmd.Flags().IntVarP(&logLimit, "limit", "l", 10, "Number of entries to show")
	logCmd.Flags().StringVarP(&logProject, "project", "p", "", "Filter by project name")
	logCmd.Flags().BoolVarP(&logToday, "today", "t", false, "Show today's entries")
	logCmd.Flags().BoolVarP(&logWeek, "week", "w", false, "Show this week's entries")
}
