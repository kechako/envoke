package variable

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"entgo.io/ent/dialect/sql"
	"github.com/kechako/envoke/cli/clierrors"
	"github.com/kechako/envoke/cli/util"
	"github.com/kechako/envoke/ent"
	"github.com/kechako/envoke/ent/variable"
	"github.com/spf13/cobra"
)

func exportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export [flags] <envfile>",
		Short: "Export environment variables to a file",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return err
			}

			if args[0] == "" {
				return clierrors.Exit(errors.New("environment file path cannot be empty"), 1)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			envfileName := args[0]
			comment, _ := cmd.Flags().GetBool("comment")
			global, _ := cmd.Flags().GetBool("global")

			globalEnv, err := util.LoadGlobalEnvironment(ctx)
			if err != nil {
				return err
			}

			env, err := util.LoadEnvironment(ctx, cmd)
			if err != nil {
				return err
			}

			globalVars, err := globalEnv.QueryVariables().Order(variable.ByName(sql.OrderAsc())).All(ctx)
			if err != nil {
				return clierrors.Exit(err, 1)
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

			err = exportVariables(envfileName, globalVars, vars, comment, global)
			if err != nil {
				return clierrors.Exit(err, 1)
			}

			return nil
		},
	}

	cmd.Flags().Bool("comment", false, "Include comments in the export (default: false)")
	cmd.Flags().Bool("global", false, "Export global variables (default: false)")

	return cmd
}

func exportVariables(name string, globalVars, vars []*ent.Variable, comment, global bool) error {
	globalEnvMap := util.MakeVariableMap(globalVars)
	envMap := util.MakeVariableMap(vars)

	file, err := os.Create(name)
	if err != nil {
		return fmt.Errorf("failed to create environment file '%s': %w", name, err)
	}
	defer file.Close()

	w := bufio.NewWriter(file)

	var exportGlobalVars []*ent.Variable
	if global {
		exportGlobalVars = globalVars
	}

	for i, v := range util.MergeVariables(exportGlobalVars, vars) {
		if i > 0 {
			w.WriteByte('\n')
		}

		if comment && v.Comment != "" {
			w.WriteString("# ")
			w.WriteString(v.Comment)
			w.WriteByte('\n')
		}

		value := v.Value
		if v.Expand {
			value, _ = util.ExpandVariable(value, globalEnvMap, envMap, false)
		}
		w.WriteString(v.Name)
		w.WriteByte('=')
		w.WriteString(value)
	}

	if err := w.Flush(); err != nil {
		return fmt.Errorf("failed to write to environment file '%s': %w", name, err)
	}

	return nil
}
