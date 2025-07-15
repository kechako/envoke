// Package run provides functionality to run commands in a specified environment.
package run

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"entgo.io/ent/dialect/sql"
	"github.com/kechako/envoke/cli/clierrors"
	"github.com/kechako/envoke/cli/util"
	"github.com/kechako/envoke/ent"
	varpred "github.com/kechako/envoke/ent/variable"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [flags] [--] <command> [<args>...]",
		Short: "Run a command in the specified environment",
		Long: `Run a command with environment variables loaded from the specified environment.

Variables are loaded in the following order of precedence:
1. System environment variables (lowest priority)
2. Global environment variables
3. Environment-specific variables (highest priority)

Variable expansion is performed for variables with the expand flag enabled.`,
		Example: `  # Run a Node.js application
  envoke run -e development npm start

  # Run database migrations
  envoke run -e production ./migrate up

  # Run a script with environment variables
  envoke run -e testing python manage.py test

  # Run with the global environment (no -e flag needed)
  envoke run python scripts/backup.py`,
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MinimumNArgs(1)(cmd, args); err != nil {
				return err
			}

			if args[0] == "" {
				return clierrors.Exit(errors.New("command cannot be empty"), 1)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			globalEnv, err := util.LoadGlobalEnvironment(ctx)
			if err != nil {
				return err
			}

			env, err := util.LoadEnvironment(ctx, cmd)
			if err != nil {
				return err
			}

			cmdName := args[0]
			cmdArgs := args[1:]

			command := exec.CommandContext(ctx, cmdName, cmdArgs...)

			command.Cancel = func() error {
				return command.Process.Signal(os.Interrupt)
			}

			command.Stdin = os.Stdin
			command.Stdout = os.Stdout
			command.Stderr = os.Stderr

			environ, err := makeEnviron(ctx, globalEnv, env)
			if err != nil {
				return clierrors.Exit(err, 1)
			}
			command.Env = environ

			if err := command.Run(); err != nil {
				var exitErr *exec.ExitError
				if errors.Is(err, exitErr) {
					return err
				}
				if !errors.Is(err, context.Canceled) {
					return clierrors.Exit(err, 1)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringP("env", "e", "", "Specify the environment to manage")

	return cmd
}

func makeEnviron(ctx context.Context, globalEnv, env *ent.Environment) ([]string, error) {
	globalVars, err := globalEnv.QueryVariables().Order(varpred.ByName(sql.OrderAsc())).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query environment variables: %w", err)
	}

	vars, err := env.QueryVariables().Order(varpred.ByName(sql.OrderAsc())).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query environment variables: %w", err)
	}

	globalEnvMap := util.MakeVariableMap(globalVars)
	envMap := util.MakeVariableMap(vars)

	environ := os.Environ()
	for _, v := range util.MergeVariables(globalVars, vars) {
		value := v.Value
		if v.Expand {
			value, _ = util.ExpandVariable(value, globalEnvMap, envMap, false)
		}
		environ = append(environ, v.Name+"="+value)
	}

	return environ, nil
}
