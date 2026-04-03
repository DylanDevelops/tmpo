package milestones

import (
	"fmt"
	"os"
	"time"

	"github.com/DylanDevelops/tmpo/internal/project"
	"github.com/DylanDevelops/tmpo/internal/settings"
	"github.com/DylanDevelops/tmpo/internal/storage"
	"github.com/DylanDevelops/tmpo/internal/ui"
	"github.com/spf13/cobra"
)

var (
	statusMilestoneProjectFlag string
)

func StatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show active milestone status",
		Long:  `Display information about the currently active milestone for the current project.`,
		Run: func(cmd *cobra.Command, args []string) {
			ui.NewlineAbove()

			db, err := storage.Initialize()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}
			defer db.Close()

			projectName, err := project.DetectConfiguredProjectWithOverride(statusMilestoneProjectFlag)
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("detecting project: %v", err))
				os.Exit(1)
			}

			// Get active milestone
			activeMilestone, err := db.GetActiveMilestoneForProject(projectName)
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			if activeMilestone == nil {
				ui.PrintWarning(ui.EmojiWarning, "No active milestone")
				ui.PrintMuted(0, "Use 'tmpo milestone start' to start a new milestone.")
				ui.NewlineBelow()
				return
			}

			// Get entries for this milestone
			entries, err := db.GetEntriesByMilestone(projectName, activeMilestone.Name)
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			// Calculate total time tracked
			var totalTime time.Duration
			for _, entry := range entries {
				if !entry.IsRunning() {
					totalTime += entry.Duration()
				}
			}

			ui.PrintSuccess(ui.EmojiMilestone, fmt.Sprintf("Active Milestone: %s", ui.Bold(activeMilestone.Name)))
			ui.PrintInfo(4, "Project", projectName)
			ui.PrintInfo(4, "Started", settings.FormatTime(activeMilestone.StartTime))
			ui.PrintInfo(4, "Duration", ui.FormatDuration(activeMilestone.Duration()))
			ui.PrintInfo(4, "Entries", fmt.Sprintf("%d", len(entries)))
			ui.PrintInfo(4, "Total Time", ui.FormatDuration(totalTime))
			ui.NewlineBelow()
		},
	}

	cmd.Flags().StringVarP(&statusMilestoneProjectFlag, "project", "p", "", "Get the milestone status for a specific global project")

	return cmd
}
