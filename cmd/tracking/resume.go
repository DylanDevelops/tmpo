package tracking

import (
	"fmt"

	"github.com/DylanDevelops/tmpo/internal/project"
	"github.com/DylanDevelops/tmpo/internal/storage"
	"github.com/DylanDevelops/tmpo/internal/ui"
	"github.com/spf13/cobra"
)

var (
	resumeProjectFlag string
)

func ResumeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resume",
		Short: "Resume time tracking",
		Long:  `Resume time tracking by starting a new session with the same project and description as the last stopped session for the current project.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ui.NewlineAbove()

			db, err := storage.Initialize()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				return err
			}

			defer db.Close()

			running, err := db.GetRunningEntry()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				return err
			}

			if running != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("Already tracking time for `%s`", running.ProjectName))
				ui.PrintMuted(0, "Use 'tmpo stop' to stop the current session first.")
				ui.NewlineBelow()
				return ui.ErrHandled
			}

			projectName, err := project.DetectConfiguredProjectWithOverride(resumeProjectFlag)
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("detecting project: %v", err))
				return err
			}

			lastStopped, err := db.GetLastStoppedEntryByProject(projectName)
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				return err
			}

			if lastStopped == nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("No previous session found for project '%s' to resume.", projectName))
				ui.PrintMuted(0, "Use 'tmpo start' to begin a new session.")
				ui.NewlineBelow()
				return ui.ErrHandled
			}

			entry, err := db.CreateEntry(lastStopped.ProjectName, lastStopped.Description, lastStopped.HourlyRate, lastStopped.MilestoneName)
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				return err
			}

			db.SaveLastAction(storage.UndoAction{Type: storage.ActionResume, EntryID: entry.ID, ProjectName: entry.ProjectName})

			ui.PrintSuccess(ui.EmojiStart, fmt.Sprintf("Resumed tracking time for %s", ui.Bold(entry.ProjectName)))

			if entry.Description != "" {
				ui.PrintInfo(4, "Description", entry.Description)
			}

			if entry.MilestoneName != nil {
				ui.PrintInfo(4, "Milestone", *entry.MilestoneName)
			}

			ui.NewlineBelow()

			return nil
		},
	}

	cmd.Flags().StringVarP(&resumeProjectFlag, "project", "p", "", "Resume tracking for a specific global project")

	return cmd
}
