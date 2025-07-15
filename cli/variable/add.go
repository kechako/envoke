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

func addCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add [flags] <name> [<value>]",
		Aliases: []string{"new"},
		Short:   "Add a new environment variable",
		Long: `Add a new environment variable to the specified environment.

If no value is provided, an empty string will be used as the value.
Use --expand flag to enable variable expansion with ${VAR} syntax.`,
		Example: `  # Add a simple variable
  envoke var add -e development DATABASE_URL "postgres://localhost/myapp_dev"

  # Add a variable with comment
  envoke var add -e development API_KEY "dev-key-123" --comment "Development API key"

  # Add an expandable variable
  envoke var add -e development API_URL '${BASE_URL}/api/v1' --expand

  # Add or update a variable
  envoke var add -e development DEBUG "true" --update`,
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

			update, _ := cmd.Flags().GetBool("update")

			env, err := util.LoadEnvironment(ctx, cmd)
			if err != nil {
				return err
			}

			client := ent.FromContext(ctx)

			create := client.Variable.Create().
				SetEnvironment(env).
				SetName(name)
			setVariableMutation(create.Mutation(), cmd, args)

			var executor interface {
				Exec(context.Context) error
			} = create
			if update {
				executor = create.
					OnConflict().
					UpdateNewValues()
			}

			err = executor.Exec(ctx)
			if err != nil {
				if ent.IsConstraintError(err) {
					return clierrors.Exit(fmt.Errorf("variable '%s' already exists in environment '%s' (use --update to modify)", name, env.Name), 1)
				}
				return clierrors.Exit(err, 1)
			}

			fmt.Printf("Environment variable '%s' added successfully!\n", name)

			return nil
		},
	}

	cmd.Flags().String("comment", "", "Comment for the variable")
	cmd.Flags().Bool("expand", false, "Expand the variable (default: false)")
	cmd.Flags().Bool("update", false, "Update the variable if it already exists (default: false)")

	return cmd
}

func setVariableMutation(m *ent.VariableMutation, cmd *cobra.Command, args []string) {
	if len(args) > 1 || m.Op().Is(ent.OpCreate) {
		if len(args) > 1 {
			m.SetValue(args[1])
		} else {
			m.SetValue("")
		}
	}

	comment, err := cmd.Flags().GetString("comment")
	if err == nil {
		if comment != "" {
			m.SetComment(comment)
		} else {
			m.ClearComment()
		}
	}

	expand, err := cmd.Flags().GetBool("expand")
	if err == nil {
		if expand {
			m.SetExpand(true)
		} else {
			m.ClearExpand()
		}
	}
}
