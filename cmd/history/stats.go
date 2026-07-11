package history

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/DylanDevelops/tmpo/internal/currency"
	"github.com/DylanDevelops/tmpo/internal/settings"
	"github.com/DylanDevelops/tmpo/internal/storage"
	"github.com/DylanDevelops/tmpo/internal/ui"
	"github.com/spf13/cobra"
)

var (
	statsToday bool
	statsWeek  bool
	statsMonth bool
	statsDate  string
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

			if statsDate != "" {
				parsedDate, err := parseDateFlag(statsDate)
				if err != nil {
					ui.PrintError(ui.EmojiError, err.Error())
					ui.NewlineBelow()
					os.Exit(1)
				}
				start = time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 0, 0, 0, 0, parsedDate.Location()).UTC()
				end = start.Add(24 * time.Hour)
				periodName = parsedDate.Format("Jan 2, 2006")
			} else if statsToday {
				now := time.Now()
				start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).UTC()
				end = start.Add(24 * time.Hour)
				periodName = "Today"
			} else if statsWeek {
				now := time.Now()
				weekday := int(now.Weekday())
				if weekday == 0 {
					weekday = 7
				}

				start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).UTC().AddDate(0, 0, -weekday+1)
				end = start.AddDate(0, 0, 7)
				periodName = "This Week"
			} else if statsMonth {
				now := time.Now()
				start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).UTC()
				end = start.AddDate(0, 1, 0)
				periodName = "This Month"
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
	cmd.Flags().BoolVarP(&statsMonth, "month", "m", false, "Show this month's stats")
	cmd.Flags().StringVarP(&statsDate, "date", "d", "", "Show stats for a specific date")

	return cmd
}

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

	var projects []string
	for project := range projectStats {
		projects = append(projects, project)
	}
	sort.Strings(projects)

	for _, project := range projects {
		duration := projectStats[project]
		percentage := 0.0

		if totalDuration > 0 {
			percentage = (duration.Seconds() / totalDuration.Seconds()) * 100
		}

		fmt.Printf("        %s  %s  (%.1f%%)\n", ui.Bold(fmt.Sprintf("%-20s", project)), ui.FormatDuration(duration), percentage)

		if earnings, ok := projectEarnings[project]; ok && earnings > 0 {
			fmt.Printf("        %s %s\n", ui.Muted("└─ Earnings:"), currency.FormatCurrency(earnings, currencyCode))
		}
	}

	ui.NewlineBelow()
}

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

func getCurrencyCode() string {
	globalCfg, err := settings.LoadGlobalConfig()
	if err != nil {
		return currency.DefaultCurrency
	}
	return globalCfg.Currency
}
