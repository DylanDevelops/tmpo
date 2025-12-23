package entries

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/DylanDevelops/tmpo/internal/project"
	"github.com/DylanDevelops/tmpo/internal/storage"
	"github.com/DylanDevelops/tmpo/internal/ui"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var showAllProjects bool

func EditCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit an existing time entry",
		Long:  `Edit an existing time entry using an interactive menu.`,
		Run: func(cmd *cobra.Command, args []string) {
			ui.NewlineAbove()
			ui.PrintSuccess("✏️", "Edit Time Entry")
			fmt.Println()

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
				// Use current project
				detectedProject, err := project.DetectConfiguredProject()
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

			// Create a copy of the entry for editing
			editedEntry := &storage.TimeEntry{
				ID:          selectedEntry.ID,
				ProjectName: selectedEntry.ProjectName,
				StartTime:   selectedEntry.StartTime,
				EndTime:     selectedEntry.EndTime,
				Description: selectedEntry.Description,
				HourlyRate:  selectedEntry.HourlyRate,
			}

			// Edit start date
			currentStartDate := selectedEntry.StartTime.Format("01-02-2006")
			startDatePrompt := promptui.Prompt{
				Label:     fmt.Sprintf("Start date (MM-DD-YYYY): (%s)", currentStartDate),
				Validate:  validateDateOptional,
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
			currentStartTime := selectedEntry.StartTime.Format("3:04 PM")
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
			currentEndDate := selectedEntry.EndTime.Format("01-02-2006")
			endDatePrompt := promptui.Prompt{
				Label:     fmt.Sprintf("End date (MM-DD-YYYY): (%s)", currentEndDate),
				Validate:  validateDateOptional,
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
			currentEndTime := selectedEntry.EndTime.Format("3:04 PM")
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
			if err := validateEndDateTime(startDateInput, startTimeInput, endDateInput, endTimeInput); err != nil {
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

			descriptionInput = strings.TrimSpace(descriptionInput)
			if descriptionInput == "" {
				descriptionInput = currentDescription
			}

			// Parse the new times
			newStartTime, err := parseDateTime(startDateInput, startTimeInput)
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("parsing start time: %v", err))
				os.Exit(1)
			}

			newEndTime, err := parseDateTime(endDateInput, endTimeInput)
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("parsing end time: %v", err))
				os.Exit(1)
			}

			// Update the edited entry
			editedEntry.StartTime = newStartTime
			editedEntry.EndTime = &newEndTime
			editedEntry.Description = descriptionInput

			// Show confirmation with diff
			fmt.Println()
			ui.PrintInfo(0, ui.Bold("Changes to entry"), fmt.Sprintf("#%d", selectedEntry.ID))
			fmt.Println()

			hasChanges := false

			selectedStartTrunc := selectedEntry.StartTime.Truncate(time.Minute)
			editedStartTrunc := editedEntry.StartTime.Truncate(time.Minute)

			if !selectedStartTrunc.Equal(editedStartTrunc) {
				hasChanges = true
				oldStr := selectedEntry.StartTime.Format("01-02-2006 3:04 PM")
				newStr := editedEntry.StartTime.Format("01-02-2006 3:04 PM")
				fmt.Printf("    %s %s → %s\n", ui.Bold("Start time:"), ui.Muted(oldStr), newStr)
			}

			selectedEndTrunc := selectedEntry.EndTime.Truncate(time.Minute)
			editedEndTrunc := editedEntry.EndTime.Truncate(time.Minute)

			if !selectedEndTrunc.Equal(editedEndTrunc) {
				hasChanges = true
				oldStr := selectedEntry.EndTime.Format("01-02-2006 3:04 PM")
				newStr := editedEntry.EndTime.Format("01-02-2006 3:04 PM")
				fmt.Printf("    %s %s → %s\n", ui.Bold("End time:"), ui.Muted(oldStr), newStr)
			}

			if selectedEntry.Description != editedEntry.Description {
				hasChanges = true
				fmt.Printf("    %s %s → %s\n", ui.Bold("Description:"), ui.Muted(fmt.Sprintf("%q", selectedEntry.Description)), fmt.Sprintf("%q", editedEntry.Description))
			}

			if !hasChanges {
				ui.PrintWarning(ui.EmojiWarning, "No changes detected")
				ui.NewlineBelow()
				os.Exit(0)
			}

			fmt.Println()

			// Confirm save
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

			fmt.Println()
			ui.PrintSuccess(ui.EmojiSuccess, "Entry updated successfully")
			ui.NewlineBelow()
		},
	}

	cmd.Flags().BoolVar(&showAllProjects, "show-all-projects", false, "Show project selection before entry selection")

	return cmd
}

// formatEntryLabel formats a time entry for display in the selection list
// Format: "2024-05-21 9:30 AM → 10:30 AM (1h) - Fixed bug in UI"
func formatEntryLabel(entry *storage.TimeEntry) string {
	startStr := entry.StartTime.Format("2006-01-02 3:04 PM")
	endStr := entry.EndTime.Format("3:04 PM")
	duration := entry.Duration()
	durationStr := ui.FormatDuration(duration)

	description := entry.Description
	if description == "" {
		description = "(no description)"
	}

	return fmt.Sprintf("%s → %s (%s) - %s", startStr, endStr, durationStr, description)
}

// validateDateOptional validates date input for edit mode, allowing empty input
// Empty input is valid (indicates keeping current value)
// Non-empty input is validated using the same rules as validateDate
func validateDateOptional(input string) error {
	if input == "" {
		return nil
	}
	return validateDate(input)
}

// validateTimeOptional validates time input for edit mode, allowing empty input
// Empty input is valid (indicates keeping current value)
// Non-empty input is validated using the same rules as validateTime
func validateTimeOptional(input string) error {
	if input == "" {
		return nil
	}
	return validateTime(input)
}
