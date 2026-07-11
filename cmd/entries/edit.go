package entries

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/DylanDevelops/tmpo/internal/project"
	"github.com/DylanDevelops/tmpo/internal/settings"
	"github.com/DylanDevelops/tmpo/internal/storage"
	"github.com/DylanDevelops/tmpo/internal/ui"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var (
	showAllProjects bool
	editProjectFlag string
)

func EditCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit an existing time entry",
		Long:  `Edit an existing time entry using an interactive menu.`,
		Run: func(cmd *cobra.Command, args []string) {
			ui.NewlineAbove()
			ui.PrintSuccess("✏️", "Edit Time Entry")
			fmt.Println()

			// Load global config to get date format preference
			globalCfg, err := settings.LoadGlobalConfig()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("loading config: %v", err))
				os.Exit(1)
			}

			// Get date format for prompts and validation
			dateFormatDisplay, dateFormatLayout := getDateFormatInfo(globalCfg.DateFormat)

			db, err := storage.Initialize()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}
			defer db.Close()

			var entries []*storage.TimeEntry
			var projectName string

			if showAllProjects {
				// Show project selection first
				projects, err := db.GetProjectsWithCompletedEntries()
				if err != nil {
					ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
					os.Exit(1)
				}

				if len(projects) == 0 {
					ui.PrintError(ui.EmojiError, "No completed time entries found")
					ui.NewlineBelow()
					os.Exit(1)
				}

				projectPrompt := promptui.Select{
					Label: "Select project",
					Items: projects,
				}

				_, selectedProject, err := projectPrompt.Run()
				if err != nil {
					ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
					os.Exit(1)
				}

				projectName = selectedProject
			} else {
				// use current project or explicit project using --project flag
				detectedProject, err := project.DetectConfiguredProjectWithOverride(editProjectFlag)
				if err != nil {
					ui.PrintError(ui.EmojiError, fmt.Sprintf("detecting project: %v", err))
					os.Exit(1)
				}
				projectName = detectedProject
			}

			// Get completed entries for the selected/detected project
			entries, err = db.GetCompletedEntriesByProject(projectName)
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			if len(entries) == 0 {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("No completed time entries found for project '%s'", projectName))
				if !showAllProjects {
					ui.PrintMuted(0, "Use 'tmpo edit --show-all-projects' to see entries from all projects")
				}
				ui.NewlineBelow()
				os.Exit(1)
			}

			// Format entries for selection
			templates := &promptui.SelectTemplates{
				Label:    "{{ . }}",
				Active:   "▸ {{ .Label }}",
				Inactive: "  {{ .Label }}",
				Selected: "{{ .Label }}",
			}

			type entryItem struct {
				Label string
				Entry *storage.TimeEntry
			}

			var items []entryItem
			for _, entry := range entries {
				label := formatEntryLabel(entry)
				items = append(items, entryItem{Label: label, Entry: entry})
			}

			entryPrompt := promptui.Select{
				Label:     "Select entry to edit",
				Items:     items,
				Templates: templates,
			}

			idx, _, err := entryPrompt.Run()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			selectedEntry := items[idx].Entry

			editedEntry := &storage.TimeEntry{
				ID:            selectedEntry.ID,
				ProjectName:   selectedEntry.ProjectName,
				StartTime:     selectedEntry.StartTime,
				EndTime:       selectedEntry.EndTime,
				Description:   selectedEntry.Description,
				HourlyRate:    selectedEntry.HourlyRate,
				MilestoneName: selectedEntry.MilestoneName,
			}

			// Edit start date
			currentStartDate := settings.FormatDateDashed(selectedEntry.StartTime)
			startDatePrompt := promptui.Prompt{
				Label:     fmt.Sprintf("Start date (%s): (%s)", dateFormatDisplay, currentStartDate),
				Validate:  func(input string) error { return validateDateOptional(input, dateFormatLayout, dateFormatDisplay) },
				AllowEdit: true,
			}

			startDateInput, err := startDatePrompt.Run()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			startDateInput = strings.TrimSpace(startDateInput)
			if startDateInput == "" {
				startDateInput = currentStartDate
			}

			// Edit start time
			currentStartTime := settings.FormatTime(selectedEntry.StartTime)
			startTimePrompt := promptui.Prompt{
				Label:     fmt.Sprintf("Start time (e.g., 9:30 AM or 14:30): (%s)", currentStartTime),
				Validate:  validateTimeOptional,
				AllowEdit: true,
			}

			startTimeInput, err := startTimePrompt.Run()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			startTimeInput = strings.TrimSpace(startTimeInput)
			if startTimeInput == "" {
				startTimeInput = currentStartTime
			}

			// Edit end date
			currentEndDate := settings.FormatDateDashed(*selectedEntry.EndTime)
			endDatePrompt := promptui.Prompt{
				Label:     fmt.Sprintf("End date (%s): (%s)", dateFormatDisplay, currentEndDate),
				Validate:  func(input string) error { return validateDateOptional(input, dateFormatLayout, dateFormatDisplay) },
				AllowEdit: true,
			}

			endDateInput, err := endDatePrompt.Run()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			endDateInput = strings.TrimSpace(endDateInput)
			if endDateInput == "" {
				endDateInput = currentEndDate
			}

			// Edit end time
			currentEndTime := settings.FormatTime(*selectedEntry.EndTime)
			endTimePrompt := promptui.Prompt{
				Label:     fmt.Sprintf("End time (e.g., 5:00 PM or 17:00): (%s)", currentEndTime),
				Validate:  validateTimeOptional,
				AllowEdit: true,
			}

			endTimeInput, err := endTimePrompt.Run()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			endTimeInput = strings.TrimSpace(endTimeInput)
			if endTimeInput == "" {
				endTimeInput = currentEndTime
			}

			// Validate that end is after start
			if err := validateEndDateTime(startDateInput, startTimeInput, endDateInput, endTimeInput, dateFormatLayout); err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			// Edit description
			currentDescription := selectedEntry.Description
			descriptionLabel := "Description"
			if currentDescription != "" {
				descriptionLabel = fmt.Sprintf("Description: (%s)", currentDescription)
			}
			descriptionPrompt := promptui.Prompt{
				Label:     descriptionLabel,
				AllowEdit: true,
			}

			descriptionInput, err := descriptionPrompt.Run()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			descriptionInput = ui.SanitizeSingleLine(descriptionInput)
			if descriptionInput == "" {
				descriptionInput = currentDescription
			}

			// edit assignment of milestone
			milestones, err := db.GetMilestonesByProject(projectName)
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			var newMilestoneName *string
			if len(milestones) > 0 {
				milestoneOptions := []string{"(None)"}
				selectedIndex := 0

				for i, m := range milestones {
					status := "Active"
					if !m.IsActive() {
						status = "Finished"
					}
					milestoneOptions = append(milestoneOptions, fmt.Sprintf("%s (%s)", m.Name, status))

					// first select current milestone if match
					if selectedEntry.MilestoneName != nil && m.Name == *selectedEntry.MilestoneName {
						selectedIndex = i + 1
					}
				}

				milestonePrompt := promptui.Select{
					Label:     "Assign to milestone (optional)",
					Items:     milestoneOptions,
					CursorPos: selectedIndex,
				}

				milestoneIdx, _, err := milestonePrompt.Run()
				if err != nil {
					ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
					os.Exit(1)
				}

				// if not its not empty set the milestone
				if milestoneIdx > 0 {
					selectedMilestone := milestones[milestoneIdx-1]
					newMilestoneName = &selectedMilestone.Name
				}
			}

			// Parse the new times
			newStartTime, err := parseDateTime(startDateInput, startTimeInput, dateFormatLayout)
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("parsing start time: %v", err))
				os.Exit(1)
			}

			newEndTime, err := parseDateTime(endDateInput, endTimeInput, dateFormatLayout)
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("parsing end time: %v", err))
				os.Exit(1)
			}

			editedEntry.StartTime = newStartTime
			editedEntry.EndTime = &newEndTime
			editedEntry.Description = descriptionInput
			editedEntry.MilestoneName = newMilestoneName

			// warn if outside of time range
			if newMilestoneName != nil {
				milestone, err := db.GetMilestoneByName(projectName, *newMilestoneName)
				if err == nil && milestone != nil {
					entryOutsideRange := false
					warningMsg := ""

					if newStartTime.Before(milestone.StartTime) {
						entryOutsideRange = true
						warningMsg = fmt.Sprintf("Entry starts (%s) before milestone began (%s)",
							settings.FormatDateTimeDashed(newStartTime),
							settings.FormatDateTimeDashed(milestone.StartTime))
					} else if milestone.EndTime != nil && newStartTime.After(*milestone.EndTime) {
						entryOutsideRange = true
						warningMsg = fmt.Sprintf("Entry starts (%s) after milestone ended (%s)",
							settings.FormatDateTimeDashed(newStartTime),
							settings.FormatDateTimeDashed(*milestone.EndTime))
					}

					if entryOutsideRange {
						fmt.Println()
						ui.PrintWarning(ui.EmojiWarning, "Entry not within milestone timeframe")
						ui.PrintMuted(0, warningMsg)
						ui.PrintMuted(0, "This is allowed - milestones are organizational tags and work with any date range.")
						fmt.Println()

						confirmPrompt := promptui.Select{
							Label: "Assign this entry to the milestone?",
							Items: []string{"Yes", "No"},
						}

						_, result, err := confirmPrompt.Run()
						if err != nil {
							ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
							os.Exit(1)
						}

						if result == "No" {
							ui.PrintWarning(ui.EmojiWarning, "Milestone assignment cancelled")
							ui.NewlineBelow()
							os.Exit(0)
						}
					}
				}
			}

			// Show confirmation with diff
			fmt.Println()
			ui.PrintInfo(0, ui.Bold("Changes to entry"), fmt.Sprintf("#%d", selectedEntry.ID))
			fmt.Println()

			hasChanges := false

			selectedStartTrunc := selectedEntry.StartTime.Truncate(time.Minute)
			editedStartTrunc := editedEntry.StartTime.Truncate(time.Minute)

			if !selectedStartTrunc.Equal(editedStartTrunc) {
				hasChanges = true
				oldStr := settings.FormatDateTimeDashed(selectedEntry.StartTime)
				newStr := settings.FormatDateTimeDashed(editedEntry.StartTime)
				fmt.Printf("    %s %s → %s\n", ui.Bold("Start time:"), ui.Muted(oldStr), newStr)
			}

			selectedEndTrunc := selectedEntry.EndTime.Truncate(time.Minute)
			editedEndTrunc := editedEntry.EndTime.Truncate(time.Minute)

			if !selectedEndTrunc.Equal(editedEndTrunc) {
				hasChanges = true
				oldStr := settings.FormatDateTimeDashed(*selectedEntry.EndTime)
				newStr := settings.FormatDateTimeDashed(*editedEntry.EndTime)
				fmt.Printf("    %s %s → %s\n", ui.Bold("End time:"), ui.Muted(oldStr), newStr)
			}

			if selectedEntry.Description != editedEntry.Description {
				hasChanges = true
				fmt.Printf("    %s %s → %s\n", ui.Bold("Description:"), ui.Muted(fmt.Sprintf("%q", selectedEntry.Description)), fmt.Sprintf("%q", editedEntry.Description))
			}

			// check if the milestone has changed
			oldMilestone := "(None)"
			if selectedEntry.MilestoneName != nil {
				oldMilestone = *selectedEntry.MilestoneName
			}
			newMilestone := "(None)"
			if editedEntry.MilestoneName != nil {
				newMilestone = *editedEntry.MilestoneName
			}
			if oldMilestone != newMilestone {
				hasChanges = true
				fmt.Printf("    %s %s → %s\n", ui.Bold("Milestone:"), ui.Muted(oldMilestone), newMilestone)
			}

			if !hasChanges {
				ui.PrintWarning(ui.EmojiWarning, "No changes detected")
				ui.NewlineBelow()
				os.Exit(0)
			}

			fmt.Println()

			confirmPrompt := promptui.Select{
				Label: "Save changes?",
				Items: []string{"Yes", "No"},
			}

			_, result, err := confirmPrompt.Run()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			if result == "No" {
				ui.PrintWarning(ui.EmojiWarning, "Changes discarded")
				ui.NewlineBelow()
				os.Exit(0)
			}

			// Save to database
			if err := db.UpdateTimeEntry(editedEntry.ID, editedEntry); err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			db.SaveLastAction(storage.UndoAction{
				Type:        storage.ActionEdit,
				EntryID:     selectedEntry.ID,
				ProjectName: selectedEntry.ProjectName,
				Entry:       selectedEntry,
			})

			fmt.Println()
			ui.PrintSuccess(ui.EmojiSuccess, "Entry updated successfully")
			ui.NewlineBelow()
		},
	}

	cmd.Flags().BoolVar(&showAllProjects, "show-all-projects", false, "Show project selection before entry selection")
	cmd.Flags().StringVarP(&editProjectFlag, "project", "p", "", "Edit entries for a specific global project")

	return cmd
}

func formatEntryLabel(entry *storage.TimeEntry) string {
	startStr := settings.FormatDateTimeDashed(entry.StartTime)
	endStr := settings.FormatTime(*entry.EndTime)
	duration := entry.Duration()
	durationStr := ui.FormatDuration(duration)

	description := entry.Description
	if description == "" {
		description = "(no description)"
	}

	return fmt.Sprintf("%s → %s (%s) - %s", startStr, endStr, durationStr, description)
}

func validateDateOptional(input, layout, displayFormat string) error {
	if input == "" {
		return nil
	}
	return validateDate(input, layout, displayFormat)
}

func validateTimeOptional(input string) error {
	if input == "" {
		return nil
	}
	return validateTime(input)
}
