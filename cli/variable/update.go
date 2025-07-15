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

func updateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [flags] <name> [<value>]",
		Short: "Update an existing environment variable",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.RangeArgs(1, 2)(cmd, args); err != nil {
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

			client.DoTransaction(ctx, nil, func(ctx context.Context, tx *ent.Tx) error {
				v, err := util.FindVariable(ctx, tx.Client(), env.ID, name)
				if err != nil {
					return clierrors.Exit(err, 1)
				}

				update := tx.Variable.UpdateOne(v)
				setVariableMutation(update.Mutation(), cmd, args)

				err = update.Exec(ctx)
				if err != nil {
					return clierrors.Exit(fmt.Errorf("failed to update variable '%s': %w", name, err), 1)
				}

				return nil
			})

			fmt.Printf("Environment variable '%s' updated successfully!\n", name)

			return nil
		},
	}

	cmd.Flags().String("comment", "", "Comment for the variable")
	cmd.Flags().Bool("expand", false, "Expand the variable (default: false)")

	return cmd
}
