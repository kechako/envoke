package environment

import (
	"fmt"
	"slices"
	"strings"

	"entgo.io/ent/dialect/sql"
	"github.com/fatih/color"
	"github.com/kechako/envoke/cli/clierrors"
	"github.com/kechako/envoke/ent"
	envpred "github.com/kechako/envoke/ent/environment"
	varpred "github.com/kechako/envoke/ent/variable"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
)

func listCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all environments",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client := ent.FromContext(ctx)

			type Env struct {
				ID             int    `json:"id"`
				Name           string `json:"name"`
				Description    string `json:"description"`
				VariablesCount int    `json:"variables_count"`
			}

			var envs []*Env

			err := client.Environment.Query().
				Order(envpred.ByName(sql.OrderAsc())).
				GroupBy(envpred.FieldID, envpred.FieldName, envpred.FieldDescription).
				Aggregate(func(s *sql.Selector) string {
					t := sql.Table(varpred.Table)
					s.LeftJoin(t).On(s.C(envpred.FieldID), t.C(varpred.FieldEnvironmentID))
					return sql.As(sql.Count(t.C(varpred.FieldID)), "variables_count")
				}).
				Scan(ctx, &envs)
			if err != nil {
				return clierrors.Exit(err, 1)
			}

			if len(envs) == 0 {
				fmt.Println("(No environments found)")
				return nil
			}

			slices.SortStableFunc(envs, func(a, b *Env) int {
				if a.Name == "global" {
					return -1 // Ensure 'global' is always first
				}
				if b.Name == "global" {
					return 1 // Ensure 'global' is always first
				}

				return strings.Compare(a.Name, b.Name)
			})

			headerFmt := color.New(color.ResetUnderline, color.Bold).SprintfFunc()

			tbl := table.New("Name", "Description", "Variables")
			tbl.WithHeaderFormatter(headerFmt)

			for _, env := range envs {
				tbl.AddRow(env.Name, env.Description, formatVariablesCount(env.VariablesCount))
			}

			tbl.Print()

			return nil
		},
	}

	return cmd
}

func formatVariablesCount(count int) string {
	return fmt.Sprintf("%8d", count)
}
