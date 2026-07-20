package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// walkCommands visits root and every command registered beneath it.
func walkCommands(root *cobra.Command, visit func(*cobra.Command)) {
	visit(root)
	for _, sub := range root.Commands() {
		walkCommands(sub, visit)
	}
}

func TestAllCommandsUseRunEInsteadOfRun(t *testing.T) {
	var checked int

	walkCommands(RootCmd(), func(cmd *cobra.Command) {
		checked++
		assert.Nilf(t, cmd.Run,
			"%q uses Run; use RunE so deferred cleanup runs on error paths", cmd.CommandPath())
	})

	// Guard against the walk silently visiting nothing and passing vacuously.
	assert.Greater(t, checked, 20, "expected the command tree to be walked")
}

func TestRunnableCommandsDeclareRunE(t *testing.T) {
	walkCommands(RootCmd(), func(cmd *cobra.Command) {
		if !cmd.Runnable() {
			return
		}

		assert.NotNilf(t, cmd.RunE, "%q is runnable but declares no RunE", cmd.CommandPath())
	})
}

func TestRootDoesNotSilenceCobraBeforeRun(t *testing.T) {
	root := RootCmd()

	assert.False(t, root.SilenceErrors, "silencing errors on the root hides unknown-flag messages")
	assert.False(t, root.SilenceUsage, "silencing usage on the root hides usage for flag errors")
	assert.NotNil(t, root.PersistentPreRun, "expected PersistentPreRun to silence Cobra for runtime errors")
}
