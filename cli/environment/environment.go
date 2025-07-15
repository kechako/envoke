// Package environment provides functionality to manage environments.
package environment

import (
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "environment",
		Aliases: []string{"env"},
		Short:   "Manage environments",
	}

	cmd.AddCommand(
		copyCommand(),
		createCommand(),
		listCommand(),
		removeCommand(),
		renameCommand(),
		updateCommand(),
	)

	return cmd
}
