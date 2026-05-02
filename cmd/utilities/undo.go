package utilities

import (
	"github.com/spf13/cobra"
)

func UndoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "undo",
		Short: "Undo the previous action",
		Long:  `Undo the previous action in case of a mistake or in need of a rollback.`,
		Run: func(cmd *cobra.Command, args []string) {

		},
	}

	return cmd
}
