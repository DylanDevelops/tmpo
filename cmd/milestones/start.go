package milestones

import (
	"fmt"
	"os"

	"github.com/DylanDevelops/tmpo/internal/project"
	"github.com/DylanDevelops/tmpo/internal/storage"
	"github.com/DylanDevelops/tmpo/internal/ui"
	"github.com/spf13/cobra"
)

var (
	startMilestoneProjectFlag string
)

func StartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start [name]",
		Short: "Start a new milestone",
		Long:  `Start a new milestone for the current project, or the one specified. Time entries created after starting a milestone will be automatically tagged with it.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ui.NewlineAbove()

			db, err := storage.Initialize()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}
			defer db.Close()

			projectName, err := project.DetectConfiguredProjectWithOverride(startMilestoneProjectFlag)
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("detecting project: %v", err))
				os.Exit(1)
			}

			milestoneName := args[0]

			// Check if there's already an active milestone
			activeMilestone, err := db.GetActiveMilestoneForProject(projectName)
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			if activeMilestone != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("Milestone '%s' is already active for %s", activeMilestone.Name, projectName))
				ui.PrintMuted(0, "Use 'tmpo milestone finish' to finish it first.")
				ui.NewlineBelow()
				os.Exit(1)
			}

			// check if this name is already in use currently or in the past
			existingMilestone, err := db.GetMilestoneByName(projectName, milestoneName)
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			if existingMilestone != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("Milestone '%s' already exists for %s", milestoneName, projectName))
				if existingMilestone.IsActive() {
					ui.PrintMuted(0, "This milestone is currently active. Use a different name or finish it first.")
				} else {
					ui.PrintMuted(0, "This milestone has already been finished. Use a different name for the new milestone.")
				}
				ui.NewlineBelow()
				os.Exit(1)
			}

			// Create the milestone
			milestone, err := db.CreateMilestone(projectName, milestoneName)
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("failed to create milestone: %v", err))
				os.Exit(1)
			}

			ui.PrintSuccess(ui.EmojiMilestone, fmt.Sprintf("Started milestone %s for %s", ui.Bold(milestone.Name), ui.Bold(projectName)))
			ui.PrintMuted(4, "└─ New time entries will be automatically tagged")
			ui.NewlineBelow()
		},
	}

	cmd.Flags().StringVarP(&startMilestoneProjectFlag, "project", "p", "", "Start a milestone for a specific global project")

	return cmd
}
