package backups

import (
	"github.com/spf13/cobra"
)

func RestoreCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore from a backup",
		Long:  `Restore your database from an existing backup.`,
		Run: func(cmd *cobra.Command, args []string) {

		},
	}

	return cmd
}
