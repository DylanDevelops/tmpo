package cmd

import (
	"fmt"
	"os"

	"github.com/DylanDevelops/tmpo/internal/storage"
	"github.com/DylanDevelops/tmpo/internal/ui"
	"github.com/spf13/cobra"
)

var resumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "Resume time tracking",
	Long:  `Resume time tracking by starting a new session with the same project and description as the last paused session.`,
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

		lastStopped, err := db.GetLastStoppedEntry()
		if err != nil {
			ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
			os.Exit(1)
		}

		if lastStopped == nil {
			ui.PrintError(ui.EmojiError, "No previous session found to resume.")
			ui.PrintMuted(0, "Use 'tmpo start' to begin a new session.")
			ui.NewlineBelow()
			os.Exit(1)
		}

		entry, err := db.CreateEntry(lastStopped.ProjectName, lastStopped.Description, lastStopped.HourlyRate)
		if err != nil {
			ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
			os.Exit(1)
		}

		ui.PrintSuccess(ui.EmojiStart, fmt.Sprintf("Resumed tracking time for %s", ui.Bold(entry.ProjectName)))

		if entry.Description != "" {
			ui.PrintInfo(4, "Description", entry.Description)
		}

		ui.NewlineBelow()
	},
}

func init() {
	rootCmd.AddCommand(resumeCmd)
}
