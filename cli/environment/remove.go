package environment

import (
	"context"
	"errors"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/kechako/envoke/cli/clierrors"
	"github.com/kechako/envoke/cli/util"
	"github.com/kechako/envoke/ent"
	"github.com/spf13/cobra"
)

func RemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		GroupID: GroupID,
		Use:     "remove [flags] <name>",
		Aliases: []string{"rm"},
		Short:   "Remove an environment",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return err
			}

			if args[0] == "" {
				return clierrors.Exit(errors.New("environment name cannot be empty"), 1)
			}
			if args[0] == "global" {
				return clierrors.Exit(errors.New("cannot remove environment 'global' (protected)"), 1)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			name := args[0]

			client := ent.FromContext(ctx)

			var env *ent.Environment
			err := client.DoTransaction(ctx, nil, func(ctx context.Context, tx *ent.Tx) error {
				var err error
				env, err = util.FindEnvironment(ctx, tx.Client(), name)
				if err != nil {
					return clierrors.Exit(err, 1)
				}

				confirm, err := confirmRemoval()
				if err != nil {
					return clierrors.Exit(err, 1)
				}
				if !confirm {
					return clierrors.Exit(errors.New("environment removal cancelled"), 0)
				}

				// Delete the environment
				if err := tx.Environment.DeleteOne(env).Exec(ctx); err != nil {
					return clierrors.Exit(err, 1)
				}

				return nil
			})
			if err != nil {
				return err
			}

			fmt.Printf("Environment '%s' removed successfully!\n", env.Name)

			return nil
		},
	}
	return cmd
}

func confirmRemoval() (bool, error) {
	confirm := false
	err := huh.NewConfirm().
		Title("Are you sure to remove this environment?").
		Value(&confirm).
		Run()
	if err != nil {
		return false, fmt.Errorf("failed to get confirmation: %w", err)
	}
	return confirm, nil
}
