package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/DylanDevelops/tmpo/internal/storage"
	"github.com/DylanDevelops/tmpo/internal/ui"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current tracking status",
	Long:  `Display information about the currently running time tracking session.`,

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
			ui.PrintWarning(ui.EmojiWarning, "Not currently tracking time")
			ui.NewlineBelow()
			ui.PrintMuted(0, "Use 'tmpo start' to begin tracking")
			ui.NewlineBelow()
			return
		}

		duration := time.Since(running.StartTime)

		ui.PrintSuccess(ui.EmojiStatus, fmt.Sprintf("Currently tracking: %s", ui.Bold(running.ProjectName)))
		ui.PrintInfo(4, ui.Bold("Started"), running.StartTime.Format("3:04 PM"))
		ui.PrintInfo(4, ui.Bold("Duration"), ui.FormatDuration(duration))

		if running.Description != "" {
			ui.PrintInfo(4, ui.Bold("Description"), running.Description)
		}

		ui.NewlineBelow()
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
