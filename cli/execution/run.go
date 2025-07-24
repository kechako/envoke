// Package execution provides functionality to run commands in a specified environment.
package execution

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"

	"entgo.io/ent/dialect/sql"
	"github.com/kechako/envoke/cli/clierrors"
	"github.com/kechako/envoke/cli/util"
	"github.com/kechako/envoke/ent"
	varpred "github.com/kechako/envoke/ent/variable"
	"github.com/spf13/cobra"
)

const GroupID = "execution"

func Command() *cobra.Command {
	cmd := &cobra.Command{
		GroupID: GroupID,
		Use:     "run [flags] [--] <command> [<args>...]",
		Short:   "Run a command in the specified environment",
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

			environ, err := makeEnviron(ctx, globalEnv, env)
			if err != nil {
				return clierrors.Exit(err, 1)
			}

			cmdName := args[0]
			cmdArgs := args[1:]

			err = execCommand(ctx, cmdName, cmdArgs, environ)
			if err != nil {
				return err
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

func execCommand(ctx context.Context, name string, args []string, env []string) error {
	// On Windows, there is no syscall.Exec, so the best we can do
	// is run a subprocess and exit with the same status.
	// Doing the same on Unix would be a problem because it wouldn't
	// propagate signals and such, but there are no signals on Windows.
	// We also use the exec case when ENVOKE_DEBUG_EXEC=0,
	// to allow testing this code even when not on Windows.
	if os.Getenv("ENVOKE_DEBUG_EXEC") == "0" || runtime.GOOS == "windows" {
		command := exec.CommandContext(ctx, name, args...)
		command.Stdin = os.Stdin
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		command.Env = env

		err := command.Run()
		if err != nil {
			if e, ok := err.(*exec.ExitError); ok && e.ProcessState != nil {
				if e.Exited() {
					return e
				}
				return clierrors.Exit(fmt.Errorf("exec %s: %w", name, e), 1)
			}
			return clierrors.Exit(fmt.Errorf("exec %s: %w", name, err), 1)
		}
		return nil
	}

	client := ent.FromContext(ctx)
	if client != nil {
		if err := client.Close(); err != nil {
			return clierrors.Exit(fmt.Errorf("failed to close database connection: %w", err), 1)
		}
	}

	path := name
	if filepath.Base(name) == name {
		lp, err := exec.LookPath(name)
		if err != nil {
			return clierrors.Exit(fmt.Errorf("lookPath %s: %w", name, err), 1)
		}
		path = lp
	}

	argv := append([]string{name}, args...)
	err := syscall.Exec(path, argv, env)
	if err != nil {
		return clierrors.Exit(fmt.Errorf("exec %s: %w", name, err), 1)
	}

	return nil
}
