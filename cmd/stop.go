package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/DylanDevelops/tmpo/internal/storage"
	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop tracking time",
	Long:  `Stop the currently running time tracking session.`,
	Run: func(cmd* cobra.Command, args []string) {
		db, err := storage.Initialize()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)

			os.Exit(1)
		}

		defer db.Close()

		running, err := db.GetRunningEntry()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)

			os.Exit(1)
		}
		
		if running == nil {
			fmt.Println("No active time tracking session.")

			os.Exit(0)
		}

		err = db.StopEntry(running.ID)
		if(err != nil) {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)

			os.Exit(1)
		}

		duration := time.Since(running.StartTime)

		fmt.Printf("[tmpo] Stopped tracking '%s'\n", running.ProjectName)
		fmt.Printf("	Total Duration: %s\n", formatDuration(duration))
	},
}

// formatDuration formats d into a concise, human-readable string using hours, minutes and seconds.
// It returns "<h>h <m>m <s>s" when the duration is at least one hour, "<m>m <s>s" when the duration
// is at least one minute but less than an hour, and "<s>s" for durations under one minute.
// Hours, minutes and seconds are derived from d using integer truncation (no fractional parts).
// This function is intended for non-negative durations; behavior for negative durations is unspecified.
func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}

	return fmt.Sprintf("%ds", seconds)
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
