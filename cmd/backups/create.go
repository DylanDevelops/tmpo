package backups

import (
	"fmt"

	"github.com/DylanDevelops/tmpo/internal/storage"
	"github.com/DylanDevelops/tmpo/internal/ui"
	"github.com/spf13/cobra"
)

func CreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new backup",
		Long:  `Create a new backup of your entire database to save all your data to be restored from later.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ui.NewlineAbove()

			db, err := storage.Initialize()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				ui.NewlineBelow()
				return ui.ErrHandled
			}
			defer db.Close()

			running, err := db.GetRunningEntry()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("checking for active timer: %v", err))
				ui.NewlineBelow()
				return ui.ErrHandled
			}
			if running != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf(`timer is running for %s — stop it before creating a backup`, ui.Bold(running.ProjectName)))
				ui.NewlineBelow()
				return ui.ErrHandled
			}

			backup, err := db.CreateBackup()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("creating backup: %v", err))
				ui.NewlineBelow()
				return ui.ErrHandled
			}

			ui.PrintSuccess(ui.EmojiBackup, "Backup created successfully")
			fmt.Println()
			ui.PrintInfo(2, "File", backup.Filename)
			ui.PrintInfo(2, "Path", backup.Path)
			ui.PrintInfo(2, "Size", ui.FormatFileSize(backup.Size))
			ui.NewlineBelow()

			return nil
		},
	}

	return cmd
}
