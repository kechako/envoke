// Package util provides utility functions for managing environments and variables in the application.
package util

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"iter"
	"os"

	"github.com/kechako/envoke/cli/clierrors"
	"github.com/kechako/envoke/ent"
	envpred "github.com/kechako/envoke/ent/environment"
	varpred "github.com/kechako/envoke/ent/variable"
	"github.com/spf13/cobra"
)

func LoadGlobalEnvironment(ctx context.Context) (*ent.Environment, error) {
	client := ent.FromContext(ctx)

	env, err := FindEnvironment(ctx, client, "global")
	if err != nil {
		return nil, clierrors.Exit(err, 1)
	}

	return env, nil
}

func LoadEnvironment(ctx context.Context, cmd *cobra.Command) (*ent.Environment, error) {
	name, err := cmd.Flags().GetString("env")
	if err != nil {
		return nil, clierrors.Exit(errors.New("environment flag (-e) is required"), 1)
	}
	if name == "" {
		return nil, clierrors.Exit(errors.New("environment name cannot be empty"), 1)
	}

	client := ent.FromContext(ctx)

	env, err := FindEnvironment(ctx, client, name)
	if err != nil {
		return nil, clierrors.Exit(err, 1)
	}

	return env, nil
}

func MakeVariableMap(vars []*ent.Variable) map[string]*ent.Variable {
	var envMap = map[string]*ent.Variable{}
	for _, v := range vars {
		envMap[v.Name] = v
	}
	return envMap
}

func MergeVariables(globalVars, vars []*ent.Variable) iter.Seq2[int, *ent.Variable] {
	envMap := MakeVariableMap(vars)

	return func(yield func(int, *ent.Variable) bool) {
		i := 0
		for _, v := range globalVars {
			if _, exists := envMap[v.Name]; exists {
				continue
			}
			if !yield(i, v) {
				return
			}
			i++
		}

		for _, v := range vars {
			if !yield(i, v) {
				return
			}
			i++
		}

	}
}

func ExpandVariable(value string, globalEnvMap, envMap map[string]*ent.Variable, errorUndefined bool) (string, error) {
	var errs []error
	value = os.Expand(value, func(name string) string {
		if v, ok := envMap[name]; ok {
			if v.Expand {
				v, err := ExpandVariable(v.Value, globalEnvMap, envMap, errorUndefined)
				if err != nil {
					errs = append(errs, err)
					return ""
				}
				return v
			}

			return v.Value
		}

		if v, ok := globalEnvMap[name]; ok {
			if v.Expand {
				v, err := ExpandVariable(v.Value, globalEnvMap, envMap, errorUndefined)
				if err != nil {
					errs = append(errs, err)
					return ""
				}
				return v
			}

			return v.Value
		}

		v, ok := os.LookupEnv(name)
		if !ok {
			errs = append(errs, fmt.Errorf("undefined variable '%s'", name))
			return ""
		}
		return v
	})
	if len(errs) > 0 && errorUndefined {
		if len(errs) == 1 {
			return "", errs[0]
		}
		return "", errors.Join(errs...)
	}

	return value, nil
}

func FindEnvironment(ctx context.Context, client *ent.Client, name string) (*ent.Environment, error) {
	environment, err := client.Environment.Query().
		Where(envpred.Name(name)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("environment '%s' not found", name)
		}
		return nil, err
	}
	return environment, nil
}

func FindVariable(ctx context.Context, client *ent.Client, environmentID int, name string) (*ent.Variable, error) {
	variable, err := client.Variable.Query().
		Where(
			varpred.EnvironmentID(environmentID),
			varpred.Name(name),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("variable '%s' not found", name)
		}
		return nil, err
	}
	return variable, nil
}

func ConfirmPrompt(prompt string) (bool, error) {
	fmt.Printf("%s (y/N): ", prompt)

	s := bufio.NewScanner(os.Stdin)
	if !s.Scan() {
		if err := s.Err(); err != nil {
			return false, fmt.Errorf("failed to read input: %w", err)
		}
		return false, fmt.Errorf("no input provided")
	}

	input := s.Text()

	if len(input) > 0 && (input[0] == 'y' || input[0] == 'Y') {
		return true, nil
	}

	return false, nil
}
