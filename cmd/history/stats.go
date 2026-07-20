package history

import (
	"fmt"
	"math"
	"os"
	"sort"
	"time"

	"github.com/DylanDevelops/tmpo/internal/currency"
	"github.com/DylanDevelops/tmpo/internal/export"
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
	statsJson  bool
)

type projectStat struct {
	Project    string   `json:"project"`
	Hours      float64  `json:"hours"`
	Percentage float64  `json:"percentage"`
	Earnings   *float64 `json:"earnings,omitempty"`
}

type statsOutput struct {
	Period          string        `json:"period"`
	TotalHours      float64       `json:"total_hours"`
	TotalEntries    int           `json:"total_entries"`
	ProjectsTracked *int          `json:"projects_tracked,omitempty"`
	TotalEarnings   *float64      `json:"total_earnings,omitempty"`
	Currency        string        `json:"currency,omitempty"`
	ByProject       []projectStat `json:"by_project"`
}

func StatsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show time tracking statistics",
		Long:  `Display statistics and summaries of your time tracking data.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !statsJson {
				ui.NewlineAbove()
			}

			db, err := storage.Initialize()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				return err
			}

			defer db.Close()

			var start, end time.Time
			var periodName string

			if statsDate != "" {
				parsedDate, err := parseDateFlag(statsDate)
				if err != nil {
					ui.PrintError(ui.EmojiError, err.Error())
					ui.NewlineBelow()
					return ui.ErrHandled
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
					return err
				}

				if statsJson {
					return export.EncodeJson(os.Stdout, buildAllTimeStatsOutput(entries, db))
				}

				ShowAllTimeStats(entries, db)
				return nil
			}

			entries, err := db.GetEntriesByDateRange(start, end)
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				return err
			}

			if statsJson {
				return export.EncodeJson(os.Stdout, buildStatsOutput(entries, periodName, nil))
			}

			ShowPeriodStats(entries, periodName)

			return nil
		},
	}

	cmd.Flags().BoolVarP(&statsToday, "today", "t", false, "Show today's stats")
	cmd.Flags().BoolVarP(&statsWeek, "week", "w", false, "Show this week's stats")
	cmd.Flags().BoolVarP(&statsMonth, "month", "m", false, "Show this month's stats")
	cmd.Flags().StringVarP(&statsDate, "date", "d", "", "Show stats for a specific date")
	cmd.Flags().BoolVar(&statsJson, "json", false, "Output stats as JSON")

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

func buildStatsOutput(entries []*storage.TimeEntry, period string, projectsTracked *int) statsOutput {
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

	var projects []string
	for project := range projectStats {
		projects = append(projects, project)
	}
	sort.Strings(projects)

	byProject := make([]projectStat, 0, len(projects))
	for _, project := range projects {
		duration := projectStats[project]
		percentage := 0.0
		if totalDuration > 0 {
			percentage = (duration.Seconds() / totalDuration.Seconds()) * 100
		}

		stat := projectStat{
			Project:    project,
			Hours:      duration.Hours(),
			Percentage: math.Round(percentage*10) / 10,
		}

		if earnings, ok := projectEarnings[project]; ok && earnings > 0 {
			stat.Earnings = &earnings
		}

		byProject = append(byProject, stat)
	}

	output := statsOutput{
		Period:          period,
		TotalHours:      totalDuration.Hours(),
		TotalEntries:    len(entries),
		ProjectsTracked: projectsTracked,
		ByProject:       byProject,
	}

	if hasAnyEarnings {
		output.TotalEarnings = &totalEarnings
		output.Currency = getCurrencyCode()
	}

	return output
}

func buildAllTimeStatsOutput(entries []*storage.TimeEntry, db *storage.Database) statsOutput {
	allProjects, _ := db.GetAllProjects()
	count := len(allProjects)

	return buildStatsOutput(entries, "All Time", &count)
}
