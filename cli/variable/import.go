package variable

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/kechako/envfile"
	"github.com/kechako/envoke/cli/clierrors"
	"github.com/kechako/envoke/cli/util"
	"github.com/kechako/envoke/ent"
	"github.com/kechako/envoke/ent/variable"
	"github.com/spf13/cobra"
)

func importCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import [flags] [<envfile>]",
		Short: "Import environment variables from a file",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.RangeArgs(0, 1)(cmd, args); err != nil {
				return err
			}

			if len(args) > 0 && args[0] == "" {
				return clierrors.Exit(errors.New("environment file path cannot be empty"), 1)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			merge, _ := cmd.Flags().GetBool("merge")

			var envfileName string
			var r io.Reader
			if len(args) == 0 {
				envfileName = "<stdin>"
				r = os.Stdin
			} else {
				envfileName = args[0]
				file, err := os.Open(envfileName)
				if err != nil {
					return clierrors.Exit(fmt.Errorf("failed to open environment file '%s': %w", envfileName, err), 1)
				}
				defer file.Close()
				r = file
			}

			envs, err := envfile.Parse(r)
			if err != nil {
				return clierrors.Exit(fmt.Errorf("failed to parse environment: %w", err), 1)
			}

			env, err := util.LoadEnvironment(ctx, cmd)
			if err != nil {
				return err
			}

			client := ent.FromContext(ctx)

			var imported = 0
			err = client.DoTransaction(ctx, nil, func(ctx context.Context, tx *ent.Tx) error {
				if merge {
					for name := range envs.Envs() {
						_, err := tx.Variable.Delete().
							Where(
								variable.EnvironmentID(env.ID),
								variable.Name(name),
							).
							Exec(ctx)
						if err != nil {
							return fmt.Errorf("failed to remove existing variable '%s': %w", name, err)
						}
					}
				} else {
					_, err := tx.Variable.Delete().
						Where(variable.EnvironmentID(env.ID)).
						Exec(ctx)
					if err != nil {
						return fmt.Errorf("failed to clear existing variables: %w", err)
					}
				}

				var builders []*ent.VariableCreate
				for name, value := range envs.Envs() {
					builders = append(builders, tx.Variable.Create().
						SetEnvironment(env).
						SetName(name).
						SetValue(value))
				}
				vars, err := tx.Variable.CreateBulk(builders...).Save(ctx)
				if err != nil {
					return fmt.Errorf("failed to create variables: %w", err)
				}
				imported = len(vars)

				return nil
			})
			if err != nil {
				return clierrors.Exit(err, 1)
			}

			fmt.Printf("Imported %d environment variables from '%s' into environment '%s'.\n", imported, envfileName, env.Name)

			return nil
		},
	}

	cmd.Flags().Bool("merge", false, "Merge with existing variables (default: false)")

	return cmd
}
