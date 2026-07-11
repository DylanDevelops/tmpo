package utilities

import (
	"fmt"
	"os"

	"github.com/DylanDevelops/tmpo/internal/storage"
	"github.com/DylanDevelops/tmpo/internal/ui"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var actionDescriptions = map[storage.ActionType]string{
	storage.ActionStop:   "Stopped tracking",
	storage.ActionCancel: "Cancelled tracking",
	storage.ActionPause:  "Paused tracking",
	storage.ActionStart:  "Started tracking",
	storage.ActionResume: "Resumed tracking",
	storage.ActionManual: "Created manual entry for",
	storage.ActionDelete: "Deleted entry for",
	storage.ActionEdit:   "Edited entry for",
}

func UndoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "undo",
		Short: "Undo the previous action",
		Long:  `Undo the previous action in case of a mistake or in need of a rollback.`,
		Run: func(cmd *cobra.Command, args []string) {
			ui.NewlineAbove()

			db, err := storage.Initialize()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}
			defer db.Close()

			action, err := db.GetLastAction()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			if action == nil {
				ui.PrintWarning(ui.EmojiWarning, "Nothing to undo.")
				ui.NewlineBelow()
				return
			}

			ui.PrintInfo(0, ui.EmojiUndo+"  Last action", undoActionDescription(action))
			fmt.Println()

			confirmPrompt := promptui.Prompt{
				Label:     "Undo this action? [y/N]",
				IsConfirm: true,
			}
			if _, err := confirmPrompt.Run(); err != nil {
				ui.PrintWarning(ui.EmojiWarning, "Undo cancelled.")
				ui.NewlineBelow()
				return
			}

			if err := applyUndo(db, action); err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("undo failed: %v", err))
				ui.NewlineBelow()
				os.Exit(1)
			}

			// not fatal if fails
			db.ClearLastAction()

			ui.PrintSuccess(ui.EmojiUndo, "Undo successful.")
			ui.NewlineBelow()
		},
	}

	return cmd
}

func undoActionDescription(action *storage.UndoAction) string {
	if prefix, ok := actionDescriptions[action.Type]; ok {
		return fmt.Sprintf("%s %s", prefix, ui.Bold(action.ProjectName))
	}
	return fmt.Sprintf("Unknown action: %s", action.Type)
}

func applyUndo(db *storage.Database, action *storage.UndoAction) error {
	switch action.Type {
	case storage.ActionStop, storage.ActionPause:
		running, err := db.GetRunningEntry()
		if err != nil {
			return fmt.Errorf("checking for running entry: %w", err)
		}
		if running != nil {
			return fmt.Errorf("a timer is already running for %s — stop it first with 'tmpo stop'", running.ProjectName)
		}
		return db.UncompleteEntry(action.EntryID)

	case storage.ActionStart, storage.ActionResume, storage.ActionManual:
		return db.DeleteTimeEntry(action.EntryID)

	case storage.ActionDelete, storage.ActionCancel:
		if action.Entry == nil {
			return fmt.Errorf("no entry snapshot available to restore")
		}
		return db.RestoreDeletedEntry(action.Entry)

	case storage.ActionEdit:
		if action.Entry == nil {
			return fmt.Errorf("no entry snapshot available to restore")
		}
		return db.UpdateTimeEntry(action.EntryID, action.Entry)

	default:
		return fmt.Errorf("unknown action type: %s", action.Type)
	}
}
