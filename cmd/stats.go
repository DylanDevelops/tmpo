package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/DylanDevelops/tmpo/internal/config"
	"github.com/DylanDevelops/tmpo/internal/storage"
	"github.com/spf13/cobra"
)

var (
	statsToday bool
	statsWeek bool
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show time tracking statistics",
	Long:  `Display statistics and summaries of your time tracking data.`,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := storage.Initialize()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			
			os.Exit(1)
		}

		defer db.Close()

		var start, end time.Time
		var periodName string

		if statsToday {
			start = time.Now().Truncate(24 * time.Hour)
			end = start.Add(24 * time.Hour)
			periodName = "Today"
		} else if statsWeek {
			now := time.Now()
			weekday := int(now.Weekday())
			if weekday == 0 {
				weekday = 7
			}

			start = now.AddDate(0, 0, -weekday+1).Truncate(24 * time.Hour)
			end = start.AddDate(0, 0, 7)
			periodName = "This week"
		} else {
			entries, err := db.GetEntries(0)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)

				os.Exit(1)
			}

			ShowAllTimeStats(entries, db)

			return
		}

		entries, err := db.GetEntriesByDateRange(start, end)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)

			os.Exit(1)
		}

		ShowPeriodStats(entries, periodName)
	},
}

func ShowPeriodStats(entries []*storage.TimeEntry, periodName string) {
	if len(entries) == 0 {
		fmt.Printf("No entries for %s.\n", periodName)

		return
	}

	projectStats := make(map[string]time.Duration)
	var totalDuration time.Duration

	for _, entry := range entries {
		duration := entry.Duration()
		projectStats[entry.ProjectName] += duration
		totalDuration += duration
	}

	fmt.Printf("\n[tmpo] Stats for %s\n\n", periodName)
	fmt.Printf("    Total Time: %s (%.2f hours)\n", formatDuration(totalDuration), totalDuration.Hours())
	fmt.Printf("    Total Entries: %d\n\n", len(entries))

	fmt.Println("    By Project:")
	for project, duration := range projectStats {
		percentage := (duration.Seconds() / totalDuration.Seconds()) * 100
		fmt.Printf("        %-20s  %s  (%.1f%%)\n", project, formatDuration(duration), percentage)
	}

	cfg, _, _ := config.FindAndLoad()
	if cfg != nil && cfg.HourlyRate > 0 {
		earnings := totalDuration.Hours() * cfg.HourlyRate
		fmt.Printf("\n        Estimated Earnings: $%.2f (at $%.2f/hr)\n", earnings, cfg.HourlyRate)
	}
}

func ShowAllTimeStats(entries []*storage.TimeEntry, db *storage.Database) {
	if len(entries) == 0 {
		fmt.Println("No entries found.")
		
		return
	}

	projectStats := make(map[string]time.Duration)
	var totalDuration time.Duration

	for _, entry := range entries {
		duration := entry.Duration()
		projectStats[entry.ProjectName] += duration
		totalDuration += duration
	}

	projects, _ := db.GetAllProjects()

	fmt.Printf("\n[tmpo] All-Time Statistics\n")
	fmt.Printf("    Total Time: %s (%.2f hours)\n", formatDuration(totalDuration), totalDuration.Hours())
	fmt.Printf("    Total Entries: %d\n", len(entries))
	fmt.Printf("    Projects Tracked: %d\n\n", len(projects))

	fmt.Println("    By Project:")
	for project, duration := range projectStats {
		percentage := (duration.Seconds() / totalDuration.Seconds()) * 100
		fmt.Printf("        %-20s  %s  (%.1f%%)\n", project, formatDuration(duration), percentage)
	}
}

func init() {
	rootCmd.AddCommand(statsCmd)

	statsCmd.Flags().BoolVarP(&statsToday, "today", "t", false, "Show today's stats")
	statsCmd.Flags().BoolVarP(&statsWeek, "week", "w", false, "Show this week's stats")
}
