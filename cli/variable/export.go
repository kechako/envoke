package variable

import (
	"bufio"
	"errors"
	"fmt"
	"io"
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
		Use:   "export [flags] [<envfile>]",
		Short: "Export environment variables to a file",
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

			var envfileName string
			var w io.Writer
			if len(args) == 0 {
				envfileName = "<stdout>"
				w = os.Stdout
			} else {
				envfileName = args[0]
				file, err := os.Create(envfileName)
				if err != nil {
					return fmt.Errorf("failed to create environment file '%s': %w", envfileName, err)
				}
				defer file.Close()
				w = file
			}

			err = exportVariables(w, globalVars, vars, comment, global)
			if err != nil {
				return clierrors.Exit(fmt.Errorf("failed to write environment file: %w", err), 1)
			}

			return nil
		},
	}

	cmd.Flags().Bool("comment", false, "Include comments in the export (default: false)")
	cmd.Flags().Bool("global", false, "Export global variables (default: false)")

	return cmd
}

func exportVariables(w io.Writer, globalVars, vars []*ent.Variable, comment, global bool) error {
	globalEnvMap := util.MakeVariableMap(globalVars)
	envMap := util.MakeVariableMap(vars)

	ew := newEnvWriter(w)

	var exportGlobalVars []*ent.Variable
	if global {
		exportGlobalVars = globalVars
	}

	for _, v := range util.MergeVariables(exportGlobalVars, vars) {
		if comment && v.Comment != "" {
			err := ew.WriteComment(v.Comment)
			if err != nil {
				return err
			}
		}

		value := v.Value
		if v.Expand {
			value, _ = util.ExpandVariable(value, globalEnvMap, envMap, false)
		}
		err := ew.WriteVariable(v.Name, value)
		if err != nil {
			return err
		}
	}

	if err := ew.Flush(); err != nil {
		return err
	}

	return nil
}

type envWriter struct {
	w   *bufio.Writer
	err error
	n   int
}

func newEnvWriter(w io.Writer) *envWriter {
	bw, ok := w.(*bufio.Writer)
	if !ok {
		bw = bufio.NewWriter(w)
	}
	return &envWriter{w: bw}
}

func (w *envWriter) Flush() error {
	return w.w.Flush()
}

func (w *envWriter) WriteComment(comment string) error {
	if w.err != nil {
		return w.err
	}

	if w.n > 0 {
		w.writeByte('\n')
	}

	w.writeString("# ")
	w.writeString(comment)

	if w.err != nil {
		return w.err
	}

	return nil
}

func (w *envWriter) WriteVariable(name, value string) error {
	if w.err != nil {
		return w.err
	}

	if w.n > 0 {
		w.writeByte('\n')
	}

	w.writeString(name)
	w.writeByte('=')
	w.writeString(value)

	if w.err != nil {
		return w.err
	}

	return nil
}

func (w *envWriter) write(p []byte) {
	if w.err != nil {
		return
	}
	var n int
	n, w.err = w.w.Write(p)
	w.n += n
}

func (w *envWriter) writeByte(b byte) {
	if w.err != nil {
		return
	}
	w.err = w.w.WriteByte(b)
	w.n++
}

func (w *envWriter) writeString(s string) {
	if w.err != nil {
		return
	}
	var n int
	n, w.err = w.w.WriteString(s)
	w.n += n
}
