package history

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/DylanDevelops/tmpo/internal/config"
	"github.com/DylanDevelops/tmpo/internal/currency"
	"github.com/DylanDevelops/tmpo/internal/storage"
	"github.com/DylanDevelops/tmpo/internal/ui"
	"github.com/spf13/cobra"
)

var (
	statsToday bool
	statsWeek bool
)

func StatsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show time tracking statistics",
		Long:  `Display statistics and summaries of your time tracking data.`,
		Run: func(cmd *cobra.Command, args []string) {
			ui.NewlineAbove()

			db, err := storage.Initialize()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			defer db.Close()

			var start, end time.Time
			var periodName string

			if statsToday {
				now := time.Now()
				start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
				end = start.Add(24 * time.Hour)
				periodName = "Today"
			} else if statsWeek {
				now := time.Now()
				weekday := int(now.Weekday())
				if weekday == 0 {
					weekday = 7
				}

				start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -weekday+1)
				end = start.AddDate(0, 0, 7)
				periodName = "This Week"
			} else {
				entries, err := db.GetEntries(0)
				if err != nil {
					ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
					os.Exit(1)
				}

				ShowAllTimeStats(entries, db)
				return
			}

			entries, err := db.GetEntriesByDateRange(start, end)
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			ShowPeriodStats(entries, periodName)
		},
	}

	cmd.Flags().BoolVarP(&statsToday, "today", "t", false, "Show today's stats")
	cmd.Flags().BoolVarP(&statsWeek, "week", "w", false, "Show this week's stats")

	return cmd
}

// ShowPeriodStats prints aggregated statistics for a named period to standard
// output. Given a slice of *storage.TimeEntry and a human-readable periodName,
// the function:
//
//  - returns early with a message if entries is empty,
//  - computes and prints the total accumulated time and its hour equivalent,
//  - prints the total number of entries,
//  - aggregates time by project and prints a per-project line with duration and
//    percentage of the total,
//  - attempts to load configuration and, if a positive hourly rate is present,
//    prints an estimated earnings line.
//
// Aggregation is done via a map[string]time.Duration; iteration order is
// therefore non-deterministic. Percentages are computed as projectSeconds /
// totalSeconds * 100, so if the total duration is zero the percentage values
// may be undefined (NaN/Inf). All output is produced using fmt.
func ShowPeriodStats(entries []*storage.TimeEntry, periodName string) {
	if len(entries) == 0 {
		ui.PrintWarning(ui.EmojiWarning, fmt.Sprintf("No entries for %s.", periodName))
		ui.NewlineBelow()
		return
	}

	projectStats := make(map[string]time.Duration)
	projectEarnings := make(map[string]float64)
	var totalDuration time.Duration
	var totalEarnings float64
	hasAnyEarnings := false

	for _, entry := range entries {
		duration := entry.Duration()
		projectStats[entry.ProjectName] += duration
		totalDuration += duration

		if entry.HourlyRate != nil {
			earnings := entry.RoundedHours() * *entry.HourlyRate
			projectEarnings[entry.ProjectName] += earnings
			totalEarnings += earnings
			hasAnyEarnings = true
		}
	}

	currencyCode := getCurrencyCode()

	ui.PrintSuccess(ui.EmojiStats, fmt.Sprintf("Stats for %s", ui.Bold(periodName)))
	fmt.Println()
	ui.PrintInfo(4, ui.Bold("Total Time"), fmt.Sprintf("%s (%.2f hours)", ui.FormatDuration(totalDuration), totalDuration.Hours()))
	ui.PrintInfo(4, ui.Bold("Total Entries"), fmt.Sprintf("%d", len(entries)))

	if hasAnyEarnings {
		ui.PrintInfo(4, ui.Bold("Earnings"), currency.FormatCurrency(totalEarnings, currencyCode))
	}

	fmt.Println()
	ui.PrintInfo(4, ui.Bold("By Project"), "")

	// Sort projects alphabetically for consistent display order
	var projects []string
	for project := range projectStats {
		projects = append(projects, project)
	}
	sort.Strings(projects)

	for _, project := range projects {
		duration := projectStats[project]
		percentage := (duration.Seconds() / totalDuration.Seconds()) * 100
		fmt.Printf("        %s  %s  (%.1f%%)\n", ui.Bold(fmt.Sprintf("%-20s", project)), ui.FormatDuration(duration), percentage)

		if earnings, ok := projectEarnings[project]; ok && earnings > 0 {
			fmt.Printf("        %s %s\n", ui.Muted("└─ Earnings:"), currency.FormatCurrency(earnings, currencyCode))
		}
	}

	ui.NewlineBelow()
}

// ShowAllTimeStats prints aggregated all-time statistics to standard output.
// Given a slice of *storage.TimeEntry and a pointer to the database, the
// function:
//
//  - returns early with a message if entries is empty,
//  - computes and prints the total accumulated time and its hour equivalent,
//  - prints the total number of entries and number of tracked projects,
//  - aggregates time by project and prints a per-project line with duration and
//    percentage of the total.
//
// The function fetches the list of projects from the provided database to
// determine the number of projects tracked. Aggregation is done via a
// map[string]time.Duration; iteration order is therefore non-deterministic.
// If the total duration is zero, percentage values may be undefined. All
// output is produced using fmt.
func ShowAllTimeStats(entries []*storage.TimeEntry, db *storage.Database) {
	if len(entries) == 0 {
		ui.PrintWarning(ui.EmojiWarning, "No entries found.")
		ui.NewlineBelow()
		return
	}

	projectStats := make(map[string]time.Duration)
	projectEarnings := make(map[string]float64)
	var totalDuration time.Duration
	var totalEarnings float64
	hasAnyEarnings := false

	for _, entry := range entries {
		duration := entry.Duration()
		projectStats[entry.ProjectName] += duration
		totalDuration += duration

		if entry.HourlyRate != nil {
			earnings := entry.RoundedHours() * *entry.HourlyRate
			projectEarnings[entry.ProjectName] += earnings
			totalEarnings += earnings
			hasAnyEarnings = true
		}
	}

	allProjects, _ := db.GetAllProjects()
	currencyCode := getCurrencyCode()

	ui.PrintSuccess(ui.EmojiStats, ui.Bold("All-Time Statistics"))
	ui.PrintInfo(4, ui.Bold("Total Time"), fmt.Sprintf("%s (%.2f hours)", ui.FormatDuration(totalDuration), totalDuration.Hours()))
	ui.PrintInfo(4, ui.Bold("Total Entries"), fmt.Sprintf("%d", len(entries)))
	ui.PrintInfo(4, ui.Bold("Projects Tracked"), fmt.Sprintf("%d", len(allProjects)))

	if hasAnyEarnings {
		ui.PrintInfo(4, ui.Bold("Earnings"), currency.FormatCurrency(totalEarnings, currencyCode))
	}

	fmt.Println()
	ui.PrintInfo(4, ui.Bold("By Project"), "")

	// Sort projects alphabetically for consistent display order
	var projects []string
	for project := range projectStats {
		projects = append(projects, project)
	}
	sort.Strings(projects)

	for _, project := range projects {
		duration := projectStats[project]
		percentage := (duration.Seconds() / totalDuration.Seconds()) * 100
		fmt.Printf("        %s  %s  (%.1f%%)\n", ui.Bold(fmt.Sprintf("%-20s", project)), ui.FormatDuration(duration), percentage)

		if earnings, ok := projectEarnings[project]; ok && earnings > 0 {
			fmt.Printf("        %s %s\n", ui.Muted("└─ Earnings:"), currency.FormatCurrency(earnings, currencyCode))
		}
	}

	ui.NewlineBelow()
}

// getCurrencyCode attempts to load the currency code from the current project's
// .tmporc configuration file. If no config is found or currency is not set,
// it returns "USD" as the default.
func getCurrencyCode() string {
	cfg, _, err := config.FindAndLoad()
	if err != nil {
		return "USD"
	}
	return cfg.GetCurrencyOrDefault()
}
