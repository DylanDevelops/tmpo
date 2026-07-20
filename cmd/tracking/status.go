package tracking

import (
	"fmt"
	"time"

	"github.com/DylanDevelops/tmpo/internal/settings"
	"github.com/DylanDevelops/tmpo/internal/storage"
	"github.com/DylanDevelops/tmpo/internal/ui"
	"github.com/spf13/cobra"
)

func StatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show current tracking status",
		Long:  `Display information about the currently running time tracking session.`,

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

			if running == nil {
				ui.PrintWarning(ui.EmojiWarning, "Not currently tracking time")
				ui.NewlineBelow()
				ui.PrintMuted(0, "Use 'tmpo start' to begin tracking")
				ui.NewlineBelow()
				return nil
			}

			duration := time.Since(running.StartTime)

			ui.PrintSuccess(ui.EmojiStatus, fmt.Sprintf("Currently tracking: %s", ui.Bold(running.ProjectName)))
			ui.PrintInfo(4, ui.Bold("Started"), settings.FormatTime(running.StartTime))
			ui.PrintInfo(4, ui.Bold("Duration"), ui.FormatDuration(duration))

			if running.Description != "" {
				ui.PrintInfo(4, ui.Bold("Description"), running.Description)
			}

			if running.MilestoneName != nil && *running.MilestoneName != "" {
				ui.PrintInfo(4, ui.Bold("Milestone"), *running.MilestoneName)
			}

			ui.NewlineBelow()

			return nil
		},
	}

	return cmd
}
