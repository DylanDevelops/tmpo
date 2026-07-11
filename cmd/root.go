package cmd

import (
	"os"

	"github.com/DylanDevelops/tmpo/cmd/backups"
	"github.com/DylanDevelops/tmpo/cmd/config"
	"github.com/DylanDevelops/tmpo/cmd/entries"
	"github.com/DylanDevelops/tmpo/cmd/history"
	"github.com/DylanDevelops/tmpo/cmd/milestones"
	"github.com/DylanDevelops/tmpo/cmd/setup"
	"github.com/DylanDevelops/tmpo/cmd/tracking"
	"github.com/DylanDevelops/tmpo/cmd/utilities"
	"github.com/spf13/cobra"
)

func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tmpo",
		Short: "Minimal CLI time tracker for developers",
		Long: `tmpo - Set the tmpo

A minimal, developer-friendly time tracking tool that lives in your terminal.
Track time effortlessly with automatic project detection and simple commands.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Check if version flag was set
			versionFlag, _ := cmd.Flags().GetBool("version")

			if versionFlag {
				utilities.DisplayVersionWithUpdateCheck()
				return
			}

			// Otherwise show help
			cmd.Help()
		},
	}

	cmd.Flags().BoolP("version", "v", false, "version for tmpo")

	// Utilities
	cmd.AddCommand(utilities.UndoCmd())
	cmd.AddCommand(utilities.VersionCmd())

	// Tracking
	cmd.AddCommand(tracking.StartCmd())
	cmd.AddCommand(tracking.StopCmd())
	cmd.AddCommand(tracking.CancelCmd())
	cmd.AddCommand(tracking.PauseCmd())
	cmd.AddCommand(tracking.ResumeCmd())
	cmd.AddCommand(tracking.StatusCmd())

	// History
	cmd.AddCommand(history.LogCmd())
	cmd.AddCommand(history.StatsCmd())
	cmd.AddCommand(history.ExportCmd())

	// Entries
	cmd.AddCommand(entries.EditCmd())
	cmd.AddCommand(entries.DeleteCmd())
	cmd.AddCommand(entries.ManualCmd())

	// Setup
	cmd.AddCommand(setup.InitCmd())

	// Configuration
	cmd.AddCommand(config.ConfigCmd())

	// Milestones
	cmd.AddCommand(milestones.MilestoneCmds())

	// Backups
	cmd.AddCommand(backups.BackupCmds())

	return cmd
}

func Execute() {
	err := RootCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}
