package environment

import (
	"context"
	"errors"
	"fmt"

	"github.com/kechako/envoke/cli/clierrors"
	"github.com/kechako/envoke/ent"
	envpred "github.com/kechako/envoke/ent/environment"
	"github.com/spf13/cobra"
)

func RenameCommand() *cobra.Command {
	cmd := &cobra.Command{
		GroupID: GroupID,
		Use:     "rename [flags] <name> <new-name>",
		Aliases: []string{"mv"},
		Short:   "Rename an environment",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(2)(cmd, args); err != nil {
				return err
			}

			if args[0] == "" {
				return clierrors.Exit(errors.New("current environment name cannot be empty"), 1)
			}
			if args[0] == "global" {
				return clierrors.Exit(errors.New("cannot rename environment 'global' (protected)"), 1)
			}
			if args[1] == "" {
				return clierrors.Exit(errors.New("new environment name cannot be empty"), 1)
			}
			if args[1] == "global" {
				return clierrors.Exit(errors.New("cannot rename to 'global' (reserved name)"), 1)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			name := args[0]
			newName := args[1]

			client := ent.FromContext(ctx)

			var env *ent.Environment

			err := client.DoTransaction(ctx, nil, func(ctx context.Context, tx *ent.Tx) error {
				var err error
				env, err = tx.Environment.Query().
					Where(envpred.Name(name)).
					Only(ctx)
				if err != nil {
					if ent.IsNotFound(err) {
						return clierrors.Exit(fmt.Errorf("environment '%s' not found", name), 1)
					}
					return err
				}

				env, err = tx.Environment.UpdateOne(env).
					SetName(newName).
					Save(ctx)
				if err != nil {
					if ent.IsConstraintError(err) {
						return clierrors.Exit(fmt.Errorf("environment '%s' already exists", newName), 1)
					}
					return err
				}

				return nil
			})
			if err != nil {
				return err
			}

			fmt.Printf("Environment '%s' renamed to '%s' successfully!\n", name, env.Name)

			return nil
		},
	}
	return cmd
}
