package tracking

import (
	"fmt"
	"os"
	"time"

	"github.com/DylanDevelops/tmpo/internal/export"
	"github.com/DylanDevelops/tmpo/internal/settings"
	"github.com/DylanDevelops/tmpo/internal/storage"
	"github.com/DylanDevelops/tmpo/internal/ui"
	"github.com/spf13/cobra"
)

var (
	statusJson bool
)

type statusOutput struct {
	Tracking bool                `json:"tracking"`
	Entry    *export.ExportEntry `json:"entry,omitempty"`
}

func StatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show current tracking status",
		Long:  `Display information about the currently running time tracking session.`,

		RunE: func(cmd *cobra.Command, args []string) error {
			if !statusJson {
				ui.NewlineAbove()
			}

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

			if statusJson {
				output := statusOutput{Tracking: running != nil}
				if running != nil {
					entry := export.BuildExportEntries([]*storage.TimeEntry{running}, false)[0]
					output.Entry = &entry
				}
				return export.EncodeJson(os.Stdout, output)
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

	cmd.Flags().BoolVar(&statusJson, "json", false, "Output status as JSON")

	return cmd
}
