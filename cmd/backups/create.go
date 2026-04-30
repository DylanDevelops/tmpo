package backups

import (
	"fmt"
	"os"

	"github.com/DylanDevelops/tmpo/internal/storage"
	"github.com/DylanDevelops/tmpo/internal/ui"
	"github.com/spf13/cobra"
)

func CreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new backup",
		Long:  `Create a new backup of your entire database to save all your data to be restored from later.`,
		Run: func(cmd *cobra.Command, args []string) {
			ui.NewlineAbove()

			db, err := storage.Initialize()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				ui.NewlineBelow()
				os.Exit(1)
			}
			defer db.Close()

			backup, err := db.CreateBackup()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("creating backup: %v", err))
				ui.NewlineBelow()
				os.Exit(1)
			}

			ui.PrintSuccess(ui.EmojiBackup, "Backup created successfully")
			fmt.Println()
			ui.PrintInfo(2, "File", backup.Filename)
			ui.PrintInfo(2, "Path", backup.Path)
			ui.PrintInfo(2, "Size", ui.FormatFileSize(backup.Size))
			ui.NewlineBelow()
		},
	}

	return cmd
}
