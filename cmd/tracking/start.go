package tracking

import (
	"fmt"
	"os"

	"github.com/DylanDevelops/tmpo/internal/config"
	"github.com/DylanDevelops/tmpo/internal/project"
	"github.com/DylanDevelops/tmpo/internal/storage"
	"github.com/DylanDevelops/tmpo/internal/ui"
	"github.com/spf13/cobra"
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

			projectName, err := project.DetectConfiguredProject()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("detecting project: %v", err))
				os.Exit(1)
			}

			description := ""
			if len(args) > 0 {
				description = args[0]
			}

			// Load config to get hourly rate if available
			var hourlyRate *float64
			if cfg, _, err := config.FindAndLoad(); err == nil && cfg != nil && cfg.HourlyRate > 0 {
				hourlyRate = &cfg.HourlyRate
			}

			entry, err := db.CreateEntry(projectName, description, hourlyRate)
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			ui.PrintSuccess(ui.EmojiStart, fmt.Sprintf("Started tracking time for %s", ui.Bold(entry.ProjectName)))

			if cfg, _, err := config.FindAndLoad(); err == nil && cfg != nil {
				ui.PrintMuted(4, "└─ Config Source: .tmporc")
			} else if project.IsInGitRepo() {
				ui.PrintMuted(4, "└─ Config Source: git repository")
			} else {
				ui.PrintMuted(4, "└─ Config Source: directory name")
			}

			if description != "" {
				ui.PrintInfo(4, "Description", description)
			}

			ui.NewlineBelow()
		},
	}

	return cmd
}
