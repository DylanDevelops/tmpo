package backups

import (
	"fmt"
	"os"

	"github.com/DylanDevelops/tmpo/internal/settings"
	"github.com/DylanDevelops/tmpo/internal/storage"
	"github.com/DylanDevelops/tmpo/internal/ui"
	"github.com/spf13/cobra"
)

func ListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lists all existing backups",
		Long:  `Lists all existing backups which can be used to restore.`,
		Run: func(cmd *cobra.Command, args []string) {
			ui.NewlineAbove()

			backups, err := storage.ListBackups()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("listing backups: %v", err))
				ui.NewlineBelow()
				os.Exit(1)
			}

			if len(backups) == 0 {
				ui.PrintInfo(0, ui.EmojiInfo+"  No backups found", "")
				ui.PrintMuted(2, "Run 'tmpo backup create' to create one.")
				ui.NewlineBelow()
				return
			}

			fmt.Printf("  %s%-4s  %-28s  %-8s  %s%s\n",
				ui.FormatBold, "ID", "Created", "Size", "Schema", ui.ColorReset)
			fmt.Println()

			for i, b := range backups {
				var schemaTag string
				if b.IsUpToDate {
					schemaTag = fmt.Sprintf("%s%s v%d (current)%s", ui.ColorGreen, ui.EmojiSuccess, b.SchemaVersion, ui.ColorReset)
				} else {
					schemaTag = fmt.Sprintf("%s%s v%d (outdated)%s", ui.ColorYellow, ui.EmojiWarning, b.SchemaVersion, ui.ColorReset)
				}

				var latestTag string
				if i == len(backups)-1 {
					latestTag = fmt.Sprintf("  %s← latest%s", ui.ColorCyan, ui.ColorReset)
				}

				fmt.Printf("  %-4d  %-28s  %-8s  %s%s\n",
					b.ID,
					settings.FormatDateTime(b.CreatedAt),
					ui.FormatFileSize(b.Size),
					schemaTag,
					latestTag,
				)
			}

			ui.NewlineBelow()
		},
	}

	return cmd
}
