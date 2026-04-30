package backups

import "github.com/spf13/cobra"

func BackupCmds() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Manage backups",
		Long:  `Manage backups to allow for easy restoration when things go wrong.`,
	}

	cmd.AddCommand(CreateCmd())
	cmd.AddCommand(RestoreCmd())
	cmd.AddCommand(ListCmd())
	cmd.AddCommand(DeleteCmd())

	return cmd
}
