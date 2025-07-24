package environment

import (
	"context"
	"errors"
	"fmt"

	"github.com/kechako/envoke/cli/clierrors"
	"github.com/kechako/envoke/cli/util"
	"github.com/kechako/envoke/ent"
	"github.com/spf13/cobra"
)

func UpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		GroupID: GroupID,
		Use:     "update [flags] <name>",
		Short:   "Update an environment",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return err
			}

			if args[0] == "" {
				return clierrors.Exit(errors.New("environment name cannot be empty"), 1)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			name := args[0]

			if name == "global" {
				return clierrors.Exit(errors.New("cannot update environment 'global' (protected)"), 1)
			}

			client := ent.FromContext(ctx)

			client.DoTransaction(ctx, nil, func(ctx context.Context, tx *ent.Tx) error {
				env, err := util.FindEnvironment(ctx, tx.Client(), name)
				if err != nil {
					return clierrors.Exit(err, 1)
				}

				update := tx.Environment.UpdateOne(env)
				setEnvironmentMutation(update.Mutation(), cmd)

				err = update.Exec(ctx)
				if err != nil {
					return clierrors.Exit(fmt.Errorf("failed to update environment '%s': %w", name, err), 1)
				}

				return nil
			})

			fmt.Printf("Environment '%s' updated successfully!\n", name)

			return nil
		},
	}

	cmd.Flags().StringP("description", "d", "", "Description of the environment")

	return cmd
}
