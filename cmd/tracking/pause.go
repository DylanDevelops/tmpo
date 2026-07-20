package tracking

import (
	"fmt"
	"time"

	"github.com/DylanDevelops/tmpo/internal/storage"
	"github.com/DylanDevelops/tmpo/internal/ui"
	"github.com/spf13/cobra"
)

func PauseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pause",
		Short: "Pause time tracking",
		Long:  `Pause the currently running time tracking session. Use 'tmpo resume' to continue tracking.`,
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
				ui.PrintWarning(ui.EmojiWarning, "No active time tracking session to pause.")
				ui.NewlineBelow()
				return nil
			}

			err = db.StopEntry(running.ID)
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				return err
			}

			db.SaveLastAction(storage.UndoAction{Type: storage.ActionPause, EntryID: running.ID, ProjectName: running.ProjectName})

			duration := time.Since(running.StartTime)

			ui.PrintSuccess(ui.EmojiStop, fmt.Sprintf("Paused tracking %s", ui.Bold(running.ProjectName)))
			ui.PrintInfo(4, ui.Bold("Session Duration"), ui.FormatDuration(duration))
			ui.PrintMuted(4, "Use 'tmpo resume' to continue tracking")

			ui.NewlineBelow()

			return nil
		},
	}

	return cmd
}
