package tracking

import (
	"fmt"
	"os"

	"github.com/DylanDevelops/tmpo/internal/storage"
	"github.com/DylanDevelops/tmpo/internal/ui"
	"github.com/spf13/cobra"
)

func CancelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel",
		Short: "Cancel the running time entry",
		Long:  "Stops and cancels the running time tracking session.",
		Run: func(cmd *cobra.Command, args []string) {
			ui.NewlineAbove()

			db, err := storage.Initialize()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			running, err := db.GetRunningEntry()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			if running == nil {
				ui.PrintWarning(ui.EmojiWarning, "No active time tracking session.")
				os.Exit(0)
			}

			err = db.CancelEntry(running.ID)
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			db.SaveLastAction(storage.UndoAction{
				Type:        storage.ActionCancel,
				ProjectName: running.ProjectName,
				Entry:       running,
			})

			ui.PrintSuccess(ui.EmojiCancel, fmt.Sprintf("Cancelled tracking %s", ui.Bold(running.ProjectName)))
			ui.PrintMuted(2, "Run 'tmpo start' to create a new time tracking session.")

			ui.NewlineBelow()
		},
	}

	return cmd
}
