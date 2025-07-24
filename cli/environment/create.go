package environment

import (
	"errors"
	"fmt"

	"github.com/kechako/envoke/cli/clierrors"
	"github.com/kechako/envoke/ent"
	"github.com/spf13/cobra"
)

func CreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		GroupID: GroupID,
		Use:     "create [flags] <name>",
		Aliases: []string{"new"},
		Short:   "Create a new environment",
		Long: `Create a new environment for managing environment variables.

Each environment maintains its own set of variables that can be used
when running commands. Variables from the 'global' environment are
always included and can be overridden by environment-specific values.`,
		Example: `  # Create a basic environment
  envoke create development

  # Create an environment with description
  envoke create production --description "Production environment"`,
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return err
			}

			if args[0] == "" {
				return clierrors.Exit(errors.New("environment name cannot be empty"), 1)
			}
			if args[0] == "global" {
				return clierrors.Exit(errors.New("cannot create environment 'global' (reserved name)"), 1)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			name := args[0]

			client := ent.FromContext(ctx)

			create := client.Environment.Create().
				SetName(name)
			setEnvironmentMutation(create.Mutation(), cmd)

			env, err := create.Save(ctx)
			if err != nil {
				if ent.IsConstraintError(err) {
					return clierrors.Exit(fmt.Errorf("environment '%s' already exists", name), 1)
				}
				return clierrors.Exit(err, 1)
			}

			// Print success message
			fmt.Printf("Environment '%s' created successfully!\n", env.Name)

			return nil
		},
	}

	cmd.Flags().StringP("description", "d", "", "Description of the environment")

	return cmd
}

func setEnvironmentMutation_(m *ent.EnvironmentMutation, cmd any) {
}

func setEnvironmentMutation(m *ent.EnvironmentMutation, cmd *cobra.Command) {
	description, err := cmd.Flags().GetString("description")
	if err == nil {
		if description != "" {
			m.SetDescription(description)
		} else {
			m.ClearDescription()
		}
	}
}
