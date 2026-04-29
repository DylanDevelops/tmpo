package backups

import (
	"github.com/spf13/cobra"
)

func ListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lists all existing backups",
		Long:  `Lists all existing backups which can be used to restore.`,
		Run: func(cmd *cobra.Command, args []string) {

		},
	}

	return cmd
}
