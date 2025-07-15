package variable

import (
	"fmt"

	"entgo.io/ent/dialect/sql"
	"github.com/fatih/color"
	"github.com/kechako/envoke/cli/clierrors"
	"github.com/kechako/envoke/cli/util"
	"github.com/kechako/envoke/ent/variable"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
)

func listCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all environment variables",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(0)(cmd, args); err != nil {
				return err
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			env, err := util.LoadEnvironment(ctx, cmd)
			if err != nil {
				return err
			}

			vars, err := env.QueryVariables().
				Order(variable.ByName(sql.OrderAsc())).
				All(ctx)
			if err != nil {
				return clierrors.Exit(err, 1)
			}

			if len(vars) == 0 {
				fmt.Println("(No environment variables found)")
				return nil
			}

			headerFmt := color.New(color.ResetUnderline, color.Bold).SprintfFunc()

			tbl := table.New("Name", "Value", "Expand", "Comment")
			tbl.WithHeaderFormatter(headerFmt)

			for _, v := range vars {
				tbl.AddRow(v.Name, v.Value, v.Expand, v.Comment)
			}

			tbl.Print()

			return nil
		},
	}
	return cmd
}
