package variable

import (
	"context"
	"errors"
	"fmt"

	"github.com/kechako/envoke/cli/clierrors"
	"github.com/kechako/envoke/cli/util"
	"github.com/kechako/envoke/ent"
	"github.com/spf13/cobra"
)

func removeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove [flags] <name>",
		Aliases: []string{"rm"},
		Short:   "Remove an environment variable",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return err
			}

			if args[0] == "" {
				return clierrors.Exit(errors.New("variable name cannot be empty"), 1)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			name := args[0]

			env, err := util.LoadEnvironment(ctx, cmd)
			if err != nil {
				return err
			}

			client := ent.FromContext(ctx)

			var v *ent.Variable

			err = client.DoTransaction(ctx, nil, func(ctx context.Context, tx *ent.Tx) error {
				var err error
				v, err = util.FindVariable(ctx, tx.Client(), env.ID, name)
				if err != nil {
					clierrors.Exit(err, 1)
				}

				// Delete the environment variable
				if err := tx.Variable.DeleteOne(v).Exec(ctx); err != nil {
					clierrors.Exit(err, 1)
				}

				return nil
			})
			if err != nil {
				return err
			}

			fmt.Printf("Environment variable '%s' removed successfully!\n", v.Name)

			return nil
		},
	}
	return cmd
}
