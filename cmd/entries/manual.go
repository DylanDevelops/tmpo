package entries

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/DylanDevelops/tmpo/internal/currency"
	"github.com/DylanDevelops/tmpo/internal/project"
	"github.com/DylanDevelops/tmpo/internal/settings"
	"github.com/DylanDevelops/tmpo/internal/storage"
	"github.com/DylanDevelops/tmpo/internal/ui"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var (
	manualProjectFlag string
)

func getDateFormatInfo(configFormat string) (displayFormat, layout string) {
	switch configFormat {
	case "MM/DD/YYYY":
		return "MM-DD-YYYY", "01-02-2006"
	case "DD/MM/YYYY":
		return "DD-MM-YYYY", "02-01-2006"
	case "YYYY-MM-DD":
		return "YYYY-MM-DD", "2006-01-02"
	default:
		return "MM-DD-YYYY", "01-02-2006"
	}
}

func ManualCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "manual",
		Short: "Create a manual time entry",
		Long:  `Create a completed time entry by specifying start and end times using an interactive menu.`,
		Run: func(cmd *cobra.Command, args []string) {
			ui.NewlineAbove()
			ui.PrintSuccess(ui.EmojiManual, "Create Manual Time Entry")
			fmt.Println()

			globalCfg, err := settings.LoadGlobalConfig()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("loading config: %v", err))
				os.Exit(1)
			}

			dateFormatDisplay, dateFormatLayout := getDateFormatInfo(globalCfg.DateFormat)
			defaultProject := detectProjectNameWithSource(manualProjectFlag)

			db, err := storage.Initialize()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}
			defer db.Close()

			var hourlyRate *float64

			todayDate := time.Now().Format(dateFormatLayout)

			var (
				projectName    string
				startDateInput string
				startTimeStr   string
				endDateInput   string
				endTimeStr     string
				description    string
				milestoneName  *string
				startTime      time.Time
				endTime        time.Time
			)

			for {
				// project name
				projectHint := defaultProject
				if projectName != "" {
					projectHint = projectName
				}

				var projectLabel string
				if projectHint != "" {
					projectLabel = fmt.Sprintf("Project name: (%s)", projectHint)
				} else {
					projectLabel = "Project name"
				}

				projectPrompt := promptui.Prompt{
					Label:     projectLabel,
					AllowEdit: true,
				}

				projectInput, promptErr := projectPrompt.Run()
				if promptErr != nil {
					ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", promptErr))
					os.Exit(1)
				}

				projectInput = ui.SanitizeSingleLine(projectInput)
				if projectInput == "" {
					projectName = projectHint
				} else {
					projectName = projectInput
				}

				if projectName == "" {
					ui.PrintError(ui.EmojiError, "project name cannot be empty")
					os.Exit(1)
				}

				hourlyRate = nil
				if configRate, _, cfgErr := project.GetProjectConfig(projectName); cfgErr == nil && configRate != nil {
					hourlyRate = configRate
				}

				// start date
				startDateHint := startDateInput
				if startDateHint == "" {
					startDateHint = todayDate
				}

				startDatePrompt := promptui.Prompt{
					Label:     fmt.Sprintf("Start date (%s): (%s)", dateFormatDisplay, startDateHint),
					AllowEdit: true,
					Validate: func(input string) error {
						if strings.TrimSpace(input) == "" {
							return nil
						}
						return validateDate(input, dateFormatLayout, dateFormatDisplay)
					},
				}

				startDateVal, promptErr := startDatePrompt.Run()
				if promptErr != nil {
					ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", promptErr))
					os.Exit(1)
				}

				startDateVal = strings.TrimSpace(startDateVal)
				if startDateVal == "" {
					startDateInput = startDateHint
				} else {
					startDateInput = startDateVal
				}

				// start time
				var startTimePrompt promptui.Prompt
				if startTimeStr != "" {
					startTimePrompt = promptui.Prompt{
						Label:     fmt.Sprintf("Start time (e.g., 9:30 AM or 14:30): (%s)", startTimeStr),
						Validate:  validateTimeOptional,
						AllowEdit: true,
					}
				} else {
					startTimePrompt = promptui.Prompt{
						Label:    "Start time (e.g., 9:30 AM or 14:30)",
						Validate: validateTime,
					}
				}

				startTimeVal, promptErr := startTimePrompt.Run()
				if promptErr != nil {
					ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", promptErr))
					os.Exit(1)
				}

				startTimeVal = strings.TrimSpace(startTimeVal)
				if startTimeVal != "" {
					startTimeStr = startTimeVal
				}

				// end date
				endDateHint := endDateInput
				if endDateHint == "" {
					endDateHint = startDateInput
				}

				endDatePrompt := promptui.Prompt{
					Label:     fmt.Sprintf("End date (%s): (%s)", dateFormatDisplay, endDateHint),
					AllowEdit: true,
				}

				endDateVal, promptErr := endDatePrompt.Run()
				if promptErr != nil {
					ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", promptErr))
					os.Exit(1)
				}

				endDateVal = strings.TrimSpace(endDateVal)
				if endDateVal == "" {
					endDateInput = endDateHint
				} else {
					endDateInput = endDateVal
				}

				if err := validateDate(endDateInput, dateFormatLayout, dateFormatDisplay); err != nil {
					ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
					os.Exit(1)
				}

				// end time
				var endTimePrompt promptui.Prompt
				if endTimeStr != "" {
					endTimePrompt = promptui.Prompt{
						Label:     fmt.Sprintf("End time (e.g., 5:00 PM or 17:00): (%s)", endTimeStr),
						Validate:  validateTimeOptional,
						AllowEdit: true,
					}
				} else {
					endTimePrompt = promptui.Prompt{
						Label:    "End time (e.g., 5:00 PM or 17:00)",
						Validate: validateTime,
					}
				}

				endTimeVal, promptErr := endTimePrompt.Run()
				if promptErr != nil {
					ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", promptErr))
					os.Exit(1)
				}

				endTimeVal = strings.TrimSpace(endTimeVal)
				if endTimeVal != "" {
					endTimeStr = endTimeVal
				}

				if err := validateEndDateTime(startDateInput, startTimeStr, endDateInput, endTimeStr, dateFormatLayout); err != nil {
					ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
					os.Exit(1)
				}

				// description
				var descLabel string
				if description != "" {
					descLabel = fmt.Sprintf("Description: (%s)", description)
				} else {
					descLabel = "Description (optional, press Enter to skip)"
				}

				descriptionPrompt := promptui.Prompt{
					Label:     descLabel,
					AllowEdit: description != "",
				}

				descVal, promptErr := descriptionPrompt.Run()
				if promptErr != nil {
					ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", promptErr))
					os.Exit(1)
				}

				descVal = ui.SanitizeSingleLine(descVal)
				if descVal != "" {
					description = descVal
				}

				parsedStart, parseErr := parseDateTime(startDateInput, startTimeStr, dateFormatLayout)
				if parseErr != nil {
					ui.PrintError(ui.EmojiError, fmt.Sprintf("parsing start time: %v", parseErr))
					os.Exit(1)
				}
				startTime = parsedStart

				parsedEnd, parseErr := parseDateTime(endDateInput, endTimeStr, dateFormatLayout)
				if parseErr != nil {
					ui.PrintError(ui.EmojiError, fmt.Sprintf("parsing end time: %v", parseErr))
					os.Exit(1)
				}
				endTime = parsedEnd

				// milestone
				milestoneName = nil
				milestones, milestoneErr := db.GetMilestonesByProject(projectName)
				if milestoneErr == nil && len(milestones) > 0 {
					milestoneOptions := []string{"(None)"}
					for _, m := range milestones {
						status := "Active"
						if !m.IsActive() {
							status = "Finished"
						}
						milestoneOptions = append(milestoneOptions, fmt.Sprintf("%s (%s)", m.Name, status))
					}

					milestonePrompt := promptui.Select{
						Label: "Assign to milestone (optional)",
						Items: milestoneOptions,
					}

					milestoneIdx, _, promptErr := milestonePrompt.Run()
					if promptErr != nil {
						ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", promptErr))
						os.Exit(1)
					}

					if milestoneIdx > 0 {
						selectedMilestone := milestones[milestoneIdx-1]
						milestoneName = &selectedMilestone.Name
					}
				}

				// summary
				tempEntry := &storage.TimeEntry{StartTime: startTime, EndTime: &endTime, HourlyRate: hourlyRate}
				duration := tempEntry.Duration()

				fmt.Println()
				ui.PrintInfo(0, ui.Bold("Entry Summary"), "")
				fmt.Println()
				ui.PrintInfo(4, ui.Bold("Project"), projectName)
				ui.PrintInfo(4, ui.Bold("Start"), settings.FormatDateTimeLong(startTime))
				ui.PrintInfo(4, ui.Bold("End"), settings.FormatDateTimeLong(endTime))
				ui.PrintInfo(4, ui.Bold("Duration"), ui.FormatDuration(duration))

				if description != "" {
					ui.PrintInfo(4, ui.Bold("Description"), description)
				} else {
					ui.PrintInfo(4, ui.Bold("Description"), ui.Muted("(none)"))
				}

				if milestoneName != nil && *milestoneName != "" {
					ui.PrintInfo(4, ui.Bold("Milestone"), *milestoneName)
				} else {
					ui.PrintInfo(4, ui.Bold("Milestone"), ui.Muted("(none)"))
				}

				if hourlyRate != nil {
					earnings := tempEntry.RoundedHours() * *hourlyRate
					fmt.Printf("    %s %s\n", ui.BoldInfo("Hourly Rate:"), currency.FormatCurrency(*hourlyRate, globalCfg.Currency))
					fmt.Printf("    %s %s\n", ui.BoldInfo("Earnings:"), currency.FormatCurrency(earnings, globalCfg.Currency))
				}

				fmt.Println()

				// confirmation
				confirmPrompt := promptui.Select{
					Label: "What would you like to do?",
					Items: []string{"Confirm", "Edit", "Cancel"},
				}

				_, result, promptErr := confirmPrompt.Run()
				if promptErr != nil {
					ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", promptErr))
					os.Exit(1)
				}

				if result == "Confirm" {
					break
				} else if result == "Cancel" {
					fmt.Println()
					ui.PrintWarning(ui.EmojiWarning, "Entry creation cancelled")
					ui.NewlineBelow()
					os.Exit(0)
				}

				fmt.Println()
			}

			entry, err := db.CreateManualEntry(projectName, description, startTime, endTime, hourlyRate, milestoneName)
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			db.SaveLastAction(storage.UndoAction{Type: storage.ActionManual, EntryID: entry.ID, ProjectName: entry.ProjectName})

			duration := entry.Duration()
			fmt.Println()
			ui.PrintSuccess(ui.EmojiSuccess, fmt.Sprintf("Created manual entry for %s", ui.Bold(entry.ProjectName)))
			ui.PrintInfo(4, ui.Bold("Start"), settings.FormatDateTimeLong(startTime))
			ui.PrintInfo(4, ui.Bold("End"), settings.FormatDateTimeLong(endTime))
			ui.PrintInfo(4, ui.Bold("Duration"), ui.FormatDuration(duration))

			if entry.Description != "" {
				ui.PrintInfo(4, ui.Bold("Description"), entry.Description)
			}

			if entry.MilestoneName != nil && *entry.MilestoneName != "" {
				ui.PrintInfo(4, ui.Bold("Milestone"), *entry.MilestoneName)
			}

			if entry.HourlyRate != nil {
				earnings := entry.RoundedHours() * *entry.HourlyRate
				fmt.Printf("    %s %s\n", ui.BoldInfo("Hourly Rate:"), currency.FormatCurrency(*entry.HourlyRate, globalCfg.Currency))
				fmt.Printf("    %s %s\n", ui.BoldInfo("Earnings:"), currency.FormatCurrency(earnings, globalCfg.Currency))
			}

			ui.NewlineBelow()
		},
	}

	cmd.Flags().StringVarP(&manualProjectFlag, "project", "p", "", "Create entry for a specific global project")

	return cmd
}

func validateDate(input, layout, displayFormat string) error {
	if input == "" {
		return fmt.Errorf("date cannot be empty")
	}

	date, err := time.Parse(layout, input)
	if err != nil {
		return fmt.Errorf("invalid date format, use %s", displayFormat)
	}

	if date.After(time.Now().Add(24 * time.Hour)) {
		return fmt.Errorf("date cannot be in the future")
	}

	return nil
}

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

func validateEndDateTime(startDate, startTime, endDate, endTime, dateLayout string) error {
	start, err := parseDateTime(startDate, startTime, dateLayout)
	if err != nil {
		return fmt.Errorf("invalid start datetime: %w", err)
	}

	end, err := parseDateTime(endDate, endTime, dateLayout)
	if err != nil {
		return fmt.Errorf("invalid end datetime: %w", err)
	}

	if !end.After(start) {
		return fmt.Errorf("end time must be after start time")
	}

	return nil
}

func parseDateTime(date, timeStr, dateLayout string) (time.Time, error) {
	normalizedTime := normalizeAMPM(timeStr)
	dateTime := fmt.Sprintf("%s %s", date, normalizedTime)

	if dt, err := time.ParseInLocation(dateLayout+" 3:04 PM", dateTime, settings.GetDisplayTimezone()); err == nil {
		return dt, nil
	}

	if dt, err := time.ParseInLocation(dateLayout+" 03:04 PM", dateTime, settings.GetDisplayTimezone()); err == nil {
		return dt, nil
	}

	return time.ParseInLocation(dateLayout+" 15:04", dateTime, settings.GetDisplayTimezone())
}

func normalizeAMPM(input string) string {
	return strings.ToUpper(input)
}

func detectProjectNameWithSource(explicitProject string) string {
	projectName, err := project.DetectConfiguredProjectWithOverride(explicitProject)
	if err != nil {
		return ""
	}

	return projectName
}
