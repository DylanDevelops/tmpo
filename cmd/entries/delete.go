package entries

import (
	"fmt"
	"os"

	"github.com/DylanDevelops/tmpo/internal/settings"
	"github.com/DylanDevelops/tmpo/internal/project"
	"github.com/DylanDevelops/tmpo/internal/storage"
	"github.com/DylanDevelops/tmpo/internal/ui"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var showAllProjectsDelete bool

func DeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a time entry",
		Long:  `Delete a time entry using an interactive menu.`,
		Run: func(cmd *cobra.Command, args []string) {
			ui.NewlineAbove()
			ui.PrintSuccess("🗑️", "Delete Time Entry")
			fmt.Println()

			db, err := storage.Initialize()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}
			defer db.Close()

			var entries []*storage.TimeEntry
			var projectName string

			if showAllProjectsDelete {
				// Show project selection first
				projects, err := db.GetAllProjects()
				if err != nil {
					ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
					os.Exit(1)
				}

				if len(projects) == 0 {
					ui.PrintError(ui.EmojiError, "No time entries found")
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

			// Get all entries for the selected/detected project
			entries, err = db.GetEntriesByProject(projectName)
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			if len(entries) == 0 {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("No time entries found for project '%s'", projectName))
				if !showAllProjectsDelete {
					ui.PrintMuted(0, "Use 'tmpo delete --show-all-projects' to see entries from all projects")
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
				label := formatEntryLabelForDelete(entry)
				items = append(items, entryItem{Label: label, Entry: entry})
			}

			entryPrompt := promptui.Select{
				Label:     "Select entry to delete",
				Items:     items,
				Templates: templates,
			}

			idx, _, err := entryPrompt.Run()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			selectedEntry := items[idx].Entry

			// Show entry details and confirmation
			fmt.Println()
			ui.PrintWarning(ui.EmojiWarning, "You are about to delete this entry:")
			fmt.Println()
			ui.PrintInfo(4, ui.Bold("ID"), fmt.Sprintf("%d", selectedEntry.ID))
			ui.PrintInfo(4, ui.Bold("Project"), selectedEntry.ProjectName)
			ui.PrintInfo(4, ui.Bold("Start"), settings.FormatDateTimeLong(selectedEntry.StartTime))
			if selectedEntry.EndTime != nil {
				ui.PrintInfo(4, ui.Bold("End"), settings.FormatDateTimeLong(*selectedEntry.EndTime))
				ui.PrintInfo(4, ui.Bold("Duration"), ui.FormatDuration(selectedEntry.Duration()))
			} else {
				ui.PrintInfo(4, ui.Bold("Status"), ui.Warning("Running"))
			}
			if selectedEntry.Description != "" {
				ui.PrintInfo(4, ui.Bold("Description"), selectedEntry.Description)
			}
			fmt.Println()

			// Confirm deletion
			confirmPrompt := promptui.Select{
				Label: "Are you sure you want to delete this entry?",
				Items: []string{"No", "Yes"},
			}

			_, result, err := confirmPrompt.Run()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			if result == "No" {
				ui.PrintWarning(ui.EmojiWarning, "Deletion cancelled")
				ui.NewlineBelow()
				os.Exit(0)
			}

			// Delete from database
			if err := db.DeleteTimeEntry(selectedEntry.ID); err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			fmt.Println()
			ui.PrintSuccess(ui.EmojiSuccess, "Entry deleted successfully")
			ui.NewlineBelow()
		},
	}

	cmd.Flags().BoolVar(&showAllProjectsDelete, "show-all-projects", false, "Show project selection before entry selection")

	return cmd
}

// formatEntryLabelForDelete formats a time entry for display in the delete selection list
// Shows running entries differently than completed ones
func formatEntryLabelForDelete(entry *storage.TimeEntry) string {
	startStr := settings.FormatDateTimeDashed(entry.StartTime)

	if entry.EndTime == nil {
		// Running entry
		description := entry.Description
		if description == "" {
			description = "(no description)"
		}
		return fmt.Sprintf("%s → Running - %s", startStr, description)
	}

	// Completed entry
	endStr := settings.FormatTime(*entry.EndTime)
	duration := entry.Duration()
	durationStr := ui.FormatDuration(duration)

	description := entry.Description
	if description == "" {
		description = "(no description)"
	}

	return fmt.Sprintf("%s → %s (%s) - %s", startStr, endStr, durationStr, description)
}
