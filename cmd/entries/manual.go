package entries

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/DylanDevelops/tmpo/internal/settings"
	"github.com/DylanDevelops/tmpo/internal/currency"
	"github.com/DylanDevelops/tmpo/internal/project"
	"github.com/DylanDevelops/tmpo/internal/storage"
	"github.com/DylanDevelops/tmpo/internal/ui"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

func ManualCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "manual",
		Short: "Create a manual time entry",
		Long:  `Create a completed time entry by specifying start and end times using an interactive menu.`,
		Run: func(cmd *cobra.Command, args []string) {
			ui.NewlineAbove()
			ui.PrintSuccess(ui.EmojiManual, "Create Manual Time Entry")
			fmt.Println()

			defaultProject := detectProjectNameWithSource()

			var projectLabel string
			if defaultProject != "" {
				projectLabel = fmt.Sprintf("Project name: (%s)", defaultProject)
			} else {
				projectLabel = "Project name"
			}

			projectPrompt := promptui.Prompt{
				Label: projectLabel,
				AllowEdit: true,
			}

			projectInput, err := projectPrompt.Run()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			projectName := strings.TrimSpace(projectInput)
			if projectName == "" {
				projectName = defaultProject
			}

			if projectName == "" {
				ui.PrintError(ui.EmojiError, "project name cannot be empty")
				os.Exit(1)
			}

			startDatePrompt := promptui.Prompt{
				Label:    "Start date (MM-DD-YYYY)",
				Validate: validateDate,
			}

			startDateInput, err := startDatePrompt.Run()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			startTimePrompt := promptui.Prompt{
				Label:    "Start time (e.g., 9:30 AM or 14:30)",
				Validate: validateTime,
			}

			startTimeStr, err := startTimePrompt.Run()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			endDateLabel := fmt.Sprintf("End date (MM-DD-YYYY): (%s)", startDateInput)

			endDatePrompt := promptui.Prompt{
				Label:     endDateLabel,
				AllowEdit: true,
			}

			endDateInput, err := endDatePrompt.Run()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			endDateInput = strings.TrimSpace(endDateInput)
			if endDateInput == "" {
				endDateInput = startDateInput
			}

			if err := validateDate(endDateInput); err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			endTimePrompt := promptui.Prompt{
				Label:    "End time (e.g., 5:00 PM or 17:00)",
				Validate: validateTime,
			}

			endTimeStr, err := endTimePrompt.Run()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			if err := validateEndDateTime(startDateInput, startTimeStr, endDateInput, endTimeStr); err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			descriptionPrompt := promptui.Prompt{
				Label: "Description (optional, press Enter to skip)",
			}

			description, err := descriptionPrompt.Run()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			startTime, err := parseDateTime(startDateInput, startTimeStr)
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("parsing start time: %v", err))
				os.Exit(1)
			}

			endTime, err := parseDateTime(endDateInput, endTimeStr)
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("parsing end time: %v", err))
				os.Exit(1)
			}

			var hourlyRate *float64
			if cfg, _, err := settings.FindAndLoad(); err == nil && cfg != nil && cfg.HourlyRate > 0 {
				hourlyRate = &cfg.HourlyRate
			}

			db, err := storage.Initialize()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}
			defer db.Close()

			entry, err := db.CreateManualEntry(projectName, description, startTime, endTime, hourlyRate)
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			duration := entry.Duration()
			fmt.Println()
			ui.PrintSuccess(ui.EmojiSuccess, fmt.Sprintf("Created manual entry for %s", ui.Bold(entry.ProjectName)))
			ui.PrintInfo(4, ui.Bold("Start"), settings.FormatDateTimeLong(startTime))
			ui.PrintInfo(4, ui.Bold("End"), settings.FormatDateTimeLong(endTime))
			ui.PrintInfo(4, ui.Bold("Duration"), ui.FormatDuration(duration))

			if entry.Description != "" {
				ui.PrintInfo(4, ui.Bold("Description"), entry.Description)
			}

			if entry.HourlyRate != nil {
				// Get currency from global config
				currencyCode := currency.DefaultCurrency
				if globalCfg, err := settings.LoadGlobalConfig(); err == nil {
					currencyCode = globalCfg.Currency
				}

				earnings := entry.RoundedHours() * *entry.HourlyRate
				fmt.Printf("    %s %s\n", ui.BoldInfo("Hourly Rate:"), currency.FormatCurrency(*entry.HourlyRate, currencyCode))
				fmt.Printf("    %s %s\n", ui.BoldInfo("Earnings:"), currency.FormatCurrency(earnings, currencyCode))
			}

			ui.NewlineBelow()
		},
	}

	return cmd
}

// validateDate validates that the provided input is a non-empty date string in MM-DD-YYYY format.
// It attempts to parse the input using the layout "01-02-2006" and returns an error if parsing fails.
// It also rejects dates that are more than 24 hours in the future (i.e., strictly after time.Now().Add(24*time.Hour)).
// Returns nil if the input is valid.
func validateDate(input string) error {
	if input == "" {
		return fmt.Errorf("date cannot be empty")
	}

	date, err := time.Parse("01-02-2006", input)
	if err != nil {
		return fmt.Errorf("invalid date format, use MM-DD-YYYY")
	}

	if date.After(time.Now().Add(24 * time.Hour)) {
		return fmt.Errorf("date cannot be in the future")
	}

	return nil
}

// validateTime validates the provided time string.
// It accepts 12-hour formats with an AM/PM designator (e.g., "9:30 AM", "09:30 PM")
// and 24-hour format (e.g., "14:30"). Empty input yields an error. The function
// normalizes AM/PM markers before parsing and returns nil on success or an error
// describing the expected formats on failure.
func validateTime(input string) error {
	if input == "" {
		return fmt.Errorf("time cannot be empty")
	}

	normalizedInput := normalizeAMPM(input)

	if _, err := time.Parse("3:04 PM", normalizedInput); err == nil {
		return nil
	}

	if _, err := time.Parse("03:04 PM", normalizedInput); err == nil {
		return nil
	}

	if _, err := time.Parse("15:04", normalizedInput); err == nil {
		return nil
	}

	return fmt.Errorf("invalid time format, use 12-hour (e.g., 9:30 AM) or 24-hour (e.g., 14:30)")
}


// validateEndDateTime verifies that the end date/time represented by
// endDate and endTime is a valid datetime and occurs strictly after the
// start date/time represented by startDate and startTime.
// It returns nil when the end datetime is strictly after the start datetime.
// If parsing of the start or end datetime fails, it returns an error
// wrapping the parse error (prefixed with "invalid start datetime" or
// "invalid end datetime"). If the end is not after the start, it
// returns an error stating that the end time must be after the start time.
func validateEndDateTime(startDate, startTime, endDate, endTime string) error {
	start, err := parseDateTime(startDate, startTime)
	if err != nil {
		return fmt.Errorf("invalid start datetime: %w", err)
	}

	end, err := parseDateTime(endDate, endTime)
	if err != nil {
		return fmt.Errorf("invalid end datetime: %w", err)
	}

	if !end.After(start) {
		return fmt.Errorf("end time must be after start time")
	}

	return nil
}

// parseDateTime combines date and time strings into time.Time
// Expects date in MM-DD-YYYY format and time in either 12-hour (with AM/PM) or 24-hour format
func parseDateTime(date, timeStr string) (time.Time, error) {
	normalizedTime := normalizeAMPM(timeStr)
	dateTime := fmt.Sprintf("%s %s", date, normalizedTime)

	if dt, err := time.ParseInLocation("01-02-2006 3:04 PM", dateTime, time.Local); err == nil {
		return dt, nil
	}

	if dt, err := time.ParseInLocation("01-02-2006 03:04 PM", dateTime, time.Local); err == nil {
		return dt, nil
	}

	return time.ParseInLocation("01-02-2006 15:04", dateTime, time.Local)
}

// normalizeAMPM converts lowercase am/pm to uppercase AM/PM
func normalizeAMPM(input string) string {
	return strings.ToUpper(input)
}

// detectProjectNameWithSource returns the project name
func detectProjectNameWithSource() (string) {
	if cfg, _, err := settings.FindAndLoad(); err == nil && cfg != nil && cfg.ProjectName != "" {
		return cfg.ProjectName
	}

	projectName, err := project.DetectProject()
	if err != nil {
		return ""
	}

	return projectName
}
