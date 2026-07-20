package backups

import (
	"fmt"
	"strconv"

	"github.com/DylanDevelops/tmpo/internal/settings"
	"github.com/DylanDevelops/tmpo/internal/storage"
	"github.com/DylanDevelops/tmpo/internal/ui"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var (
	restoreIDFlag string
)

func RestoreCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore from a backup",
		Long:  `Restore your database from an existing backup.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ui.NewlineAbove()

			db, err := storage.Initialize()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				ui.NewlineBelow()
				return ui.ErrHandled
			}

			running, err := db.GetRunningEntry()
			if err != nil {
				db.Close()
				ui.PrintError(ui.EmojiError, fmt.Sprintf("checking for active timer: %v", err))
				ui.NewlineBelow()
				return ui.ErrHandled
			}
			if running != nil {
				db.Close()
				ui.PrintError(ui.EmojiError, fmt.Sprintf(`timer is running for %s — stop it before restoring a backup`, ui.Bold(running.ProjectName)))
				ui.NewlineBelow()
				return ui.ErrHandled
			}

			if err := db.Close(); err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("closing database: %v", err))
				ui.NewlineBelow()
				return ui.ErrHandled
			}

			backups, err := storage.ListBackups()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("listing backups: %v", err))
				ui.NewlineBelow()
				return ui.ErrHandled
			}

			if len(backups) == 0 {
				ui.PrintInfo(0, ui.EmojiInfo+"  No backups found", "")
				ui.PrintMuted(2, "Run 'tmpo backup create' to create one.")
				ui.NewlineBelow()
				return nil
			}

			var selected *storage.BackupInfo

			if restoreIDFlag != "" {
				if id, err := strconv.Atoi(restoreIDFlag); err == nil {
					for i := range backups {
						if backups[i].ID == id {
							selected = &backups[i]
							break
						}
					}
					if selected == nil {
						ui.PrintError(ui.EmojiError, fmt.Sprintf("no backup found with ID %d", id))
						ui.NewlineBelow()
						return ui.ErrHandled
					}
				} else {
					for i := range backups {
						if backups[i].Filename == restoreIDFlag {
							selected = &backups[i]
							break
						}
					}
					if selected == nil {
						ui.PrintError(ui.EmojiError, fmt.Sprintf("no backup found with filename %q", restoreIDFlag))
						ui.NewlineBelow()
						return ui.ErrHandled
					}
				}
			} else {
				items := make([]string, len(backups))
				for i, b := range backups {
					versionTag := fmt.Sprintf("v%d, current", b.SchemaVersion)
					if !b.IsUpToDate {
						versionTag = fmt.Sprintf("v%d, outdated", b.SchemaVersion)
					}
					items[i] = fmt.Sprintf("[%d]  %s  %s  (%s)",
						b.ID,
						settings.FormatDateTime(b.CreatedAt),
						ui.FormatFileSize(b.Size),
						versionTag,
					)
				}

				prompt := promptui.Select{
					Label: "Select backup to restore",
					Items: items,
				}

				idx, _, err := prompt.Run()
				if err != nil {
					ui.NewlineBelow()
					return nil
				}

				selected = &backups[idx]
			}

			if !selected.IsUpToDate {
				fmt.Printf("\n  %s%s  This backup uses schema v%d; the current binary expects v%d.%s\n",
					ui.ColorYellow, ui.EmojiWarning, selected.SchemaVersion, storage.CurrentSchemaVersion, ui.ColorReset)
				fmt.Printf("  %sMigrations will run automatically after restore.%s\n\n",
					ui.ColorGray, ui.ColorReset)
			}

			confirmPrompt := promptui.Prompt{
				Label:     fmt.Sprintf("Restore from %s? This will overwrite your current database [y/N]", selected.Filename),
				IsConfirm: true,
			}

			if _, err := confirmPrompt.Run(); err != nil {
				ui.PrintInfo(0, ui.EmojiInfo+"  Restore cancelled", "")
				ui.NewlineBelow()
				return nil
			}

			if err := storage.RestoreBackup(selected.Path); err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("restoring backup: %v", err))
				ui.NewlineBelow()
				return ui.ErrHandled
			}

			ui.PrintSuccess(ui.EmojiBackup, fmt.Sprintf("Restored from %s", selected.Filename))
			ui.NewlineBelow()

			return nil
		},
	}

	cmd.Flags().StringVarP(&restoreIDFlag, "id", "i", "", "backup ID or filename to restore (skips interactive selection)")

	return cmd
}
