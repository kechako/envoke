// Package variable provides functionality to manage environment variables.
package variable

import (
	"github.com/spf13/cobra"
)

const GroupID = "variable"

func Command() *cobra.Command {
	cmd := &cobra.Command{
		GroupID: GroupID,
		Use:     "variable",
		Aliases: []string{"var"},
		Short:   "Manage environment variables",
	}

	cmd.AddCommand(
		addCommand(),
		exportCommand(),
		importCommand(),
		listCommand(),
		removeCommand(),
		updateCommand(),
	)

	cmd.PersistentFlags().StringP("env", "e", "", "Specify the environment to manage")

	return cmd
}
