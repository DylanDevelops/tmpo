package backups

import (
	"github.com/spf13/cobra"
)

func CreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new backup",
		Long:  `Create a new backup of your entire database to save all your data to be restored from later.`,
		Run: func(cmd *cobra.Command, args []string) {

		},
	}

	return cmd
}
