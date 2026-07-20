package tracking

import (
	"fmt"

	"github.com/DylanDevelops/tmpo/internal/storage"
	"github.com/DylanDevelops/tmpo/internal/ui"
	"github.com/spf13/cobra"
)

func CancelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel",
		Short: "Cancel the running time entry",
		Long:  "Stops and cancels the running time tracking session.",
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
				ui.PrintWarning(ui.EmojiWarning, "No active time tracking session.")
				ui.NewlineBelow()
				return nil
			}

			err = db.CancelEntry(running.ID)
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				return err
			}

			db.SaveLastAction(storage.UndoAction{
				Type:        storage.ActionCancel,
				ProjectName: running.ProjectName,
				Entry:       running,
			})

			ui.PrintSuccess(ui.EmojiCancel, fmt.Sprintf("Cancelled tracking %s", ui.Bold(running.ProjectName)))
			ui.PrintMuted(4, "If this was a mistake, run `tmpo undo`.")

			ui.NewlineBelow()

			return nil
		},
	}

	return cmd
}
