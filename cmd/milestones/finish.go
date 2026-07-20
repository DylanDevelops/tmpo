package milestones

import (
	"fmt"

	"github.com/DylanDevelops/tmpo/internal/project"
	"github.com/DylanDevelops/tmpo/internal/storage"
	"github.com/DylanDevelops/tmpo/internal/ui"
	"github.com/spf13/cobra"
)

var (
	finishMilestoneProjectFlag string
)

func FinishCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "finish",
		Short: "Finish the active milestone",
		Long:  `Finish the currently active milestone for the current project, or the one specified. This marks the milestone as completed and stops auto-tagging new time entries with it.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ui.NewlineAbove()

			db, err := storage.Initialize()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				return err
			}
			defer db.Close()

			projectName, err := project.DetectConfiguredProjectWithOverride(finishMilestoneProjectFlag)
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("detecting project: %v", err))
				return err
			}

			// Get active milestone
			activeMilestone, err := db.GetActiveMilestoneForProject(projectName)
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				return err
			}

			if activeMilestone == nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("No active milestone found for %s", projectName))
				ui.PrintMuted(0, "Use 'tmpo milestone start' to start a new milestone.")
				ui.NewlineBelow()
				return ui.ErrHandled
			}

			// Get entries for this milestone to show count
			entries, err := db.GetEntriesByMilestone(projectName, activeMilestone.Name)
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				return err
			}

			// Finish the milestone
			err = db.FinishMilestone(activeMilestone.ID)
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				return err
			}

			// Get updated milestone to show duration
			finishedMilestone, err := db.GetMilestone(activeMilestone.ID)
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				return err
			}

			ui.PrintSuccess(ui.EmojiMilestone, fmt.Sprintf("Finished milestone %s", ui.Bold(finishedMilestone.Name)))
			ui.PrintInfo(4, "Duration", ui.FormatDuration(finishedMilestone.Duration()))
			ui.PrintInfo(4, "Entries", fmt.Sprintf("%d", len(entries)))
			ui.NewlineBelow()

			return nil
		},
	}

	cmd.Flags().StringVarP(&finishMilestoneProjectFlag, "project", "p", "", "Finish a milestone for a specific global project")

	return cmd
}
