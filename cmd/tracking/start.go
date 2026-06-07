package tracking

import (
	"fmt"
	"os"

	"github.com/DylanDevelops/tmpo/internal/project"
	"github.com/DylanDevelops/tmpo/internal/settings"
	"github.com/DylanDevelops/tmpo/internal/storage"
	"github.com/DylanDevelops/tmpo/internal/ui"
	"github.com/spf13/cobra"
)

var (
	startProjectFlag string
)

func StartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start [description]",
		Short: "Start tracking time",
		Long:  `Start a new time tracking session for the current project.`,
		Run: func(cmd *cobra.Command, args []string) {
			ui.NewlineAbove()

			db, err := storage.Initialize()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			defer db.Close()

			running, err := db.GetRunningEntry()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			if running != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("Already tracking time for `%s`", running.ProjectName))
				ui.PrintMuted(0, "Use 'tmpo stop' to stop the current session first.")
				ui.NewlineBelow()
				os.Exit(1)
			}

			projectName, err := project.DetectConfiguredProjectWithOverride(startProjectFlag)
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("detecting project: %v", err))
				os.Exit(1)
			}

			description := ""
			if len(args) > 0 {
				description = ui.SanitizeSingleLine(args[0])
			}

			var hourlyRate *float64
			configRate, _, err := project.GetProjectConfig(projectName)
			if err == nil && configRate != nil {
				hourlyRate = configRate
			}

			var milestoneName *string
			activeMilestone, err := db.GetActiveMilestoneForProject(projectName)

			if activeMilestone != nil {
				milestoneName = &activeMilestone.Name
			}

			entry, err := db.CreateEntry(projectName, description, hourlyRate, milestoneName)
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			db.SaveLastAction(storage.UndoAction{Type: storage.ActionStart, EntryID: entry.ID, ProjectName: entry.ProjectName})

			ui.PrintSuccess(ui.EmojiStart, fmt.Sprintf("Started tracking time for %s", ui.Bold(entry.ProjectName)))

			// communicate config source to user
			if startProjectFlag != "" {
				ui.PrintMuted(4, "└─ Config Source: global project")
			} else if cfg, _, err := settings.FindAndLoad(); err == nil && cfg != nil {
				ui.PrintMuted(4, "└─ Config Source: .tmporc")
			} else if project.IsInGitRepo() {
				ui.PrintMuted(4, "└─ Config Source: git repository")
			} else {
				ui.PrintMuted(4, "└─ Config Source: directory name")
			}

			if description != "" {
				ui.PrintInfo(4, "Description", description)
			}

			if milestoneName != nil {
				ui.PrintInfo(4, "Milestone", *milestoneName)
			}

			ui.NewlineBelow()
		},
	}

	cmd.Flags().StringVarP(&startProjectFlag, "project", "p", "", "Track time for a specific global project")

	return cmd
}
