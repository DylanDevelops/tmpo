package backups

import (
	"fmt"
	"os"
	"strconv"

	"github.com/DylanDevelops/tmpo/internal/settings"
	"github.com/DylanDevelops/tmpo/internal/storage"
	"github.com/DylanDevelops/tmpo/internal/ui"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var (
	deleteIDFlag string
)

func DeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a backup",
		Long:  `Permanently delete an existing backup. This action cannot be undone.`,
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

			var selected *storage.BackupInfo

			if deleteIDFlag != "" {
				if id, err := strconv.Atoi(deleteIDFlag); err == nil {
					for i := range backups {
						if backups[i].ID == id {
							selected = &backups[i]
							break
						}
					}
					if selected == nil {
						ui.PrintError(ui.EmojiError, fmt.Sprintf("no backup found with ID %d", id))
						ui.NewlineBelow()
						os.Exit(1)
					}
				} else {
					for i := range backups {
						if backups[i].Filename == deleteIDFlag {
							selected = &backups[i]
							break
						}
					}
					if selected == nil {
						ui.PrintError(ui.EmojiError, fmt.Sprintf("no backup found with filename %q", deleteIDFlag))
						ui.NewlineBelow()
						os.Exit(1)
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
					Label: "Select backup to delete",
					Items: items,
				}

				idx, _, err := prompt.Run()
				if err != nil {
					ui.NewlineBelow()
					return
				}

				selected = &backups[idx]
			}

			confirmPrompt := promptui.Prompt{
				Label:     fmt.Sprintf("Permanently delete %s? This cannot be undone [y/N]", selected.Filename),
				IsConfirm: true,
			}

			if _, err := confirmPrompt.Run(); err != nil {
				ui.PrintInfo(0, ui.EmojiInfo+"  Deletion cancelled", "")
				ui.NewlineBelow()
				return
			}

			if err := os.Remove(selected.Path); err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("deleting backup: %v", err))
				ui.NewlineBelow()
				os.Exit(1)
			}

			ui.PrintSuccess(ui.EmojiSuccess, fmt.Sprintf("Deleted %s", selected.Filename))
			ui.NewlineBelow()
		},
	}

	cmd.Flags().StringVarP(&deleteIDFlag, "id", "i", "", "backup ID or filename to delete (skips interactive selection)")

	return cmd
}
