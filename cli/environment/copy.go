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

func copyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "copy [flags] <name> <new-name>",
		Aliases: []string{"cp"},
		Short:   "Copy an environment",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(2)(cmd, args); err != nil {
				return err
			}

			if args[0] == "" {
				return clierrors.Exit(errors.New("source environment name cannot be empty"), 1)
			}
			if args[1] == "" {
				return clierrors.Exit(errors.New("destination environment name cannot be empty"), 1)
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
					WithVariables().
					Only(ctx)
				if err != nil {
					if ent.IsNotFound(err) {
						return clierrors.Exit(fmt.Errorf("environment '%s' not found", name), 1)
					}
					return err
				}

				create := tx.Environment.Create().
					SetName(newName)
				if env.Description != "" {
					create.SetDescription(env.Description)
				}

				newEnv, err := create.Save(ctx)
				if err != nil {
					if ent.IsConstraintError(err) {
						return clierrors.Exit(fmt.Errorf("environment '%s' already exists", newName), 1)
					}
					return err
				}

				builders := make([]*ent.VariableCreate, len(env.Edges.Variables))
				for i, v := range env.Edges.Variables {
					create := tx.Variable.Create().
						SetEnvironment(newEnv).
						SetName(v.Name).
						SetValue(v.Value)
					if v.Expand {
						create.SetExpand(true)
					}
					if v.Comment != "" {
						create.SetComment(v.Comment)
					}

					builders[i] = create
				}
				_, err = tx.Variable.CreateBulk(builders...).Save(ctx)
				if err != nil {
					return clierrors.Exit(fmt.Errorf("failed to copy variables: %w", err), 1)
				}

				return nil
			})
			if err != nil {
				return err
			}

			fmt.Printf("Environment '%s' copied to '%s' successfully!\n", name, newName)

			return nil
		},
	}
	return cmd
}
