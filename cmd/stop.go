package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/DylanDevelops/tmpo/internal/storage"
	"github.com/DylanDevelops/tmpo/internal/ui"
	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop tracking time",
	Long:  `Stop the currently running time tracking session.`,
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

		if running == nil {
			ui.PrintWarning(ui.EmojiWarning, "No active time tracking session.")
			os.Exit(0)
		}

		err = db.StopEntry(running.ID)
		if(err != nil) {
			ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
			os.Exit(1)
		}

		duration := time.Since(running.StartTime)

		ui.PrintSuccess(ui.EmojiStop, fmt.Sprintf("Stopped tracking %s", ui.Bold(running.ProjectName)))
		ui.PrintInfo(4, ui.Bold("Total Duration"), ui.FormatDuration(duration))

		ui.NewlineBelow()
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
