// Package cli provides the command-line interface for the application.
package cli

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"

	"entgo.io/ent/dialect/sql/schema"
	"github.com/kechako/envoke/cli/clierrors"
	"github.com/kechako/envoke/cli/environment"
	"github.com/kechako/envoke/cli/run"
	"github.com/kechako/envoke/cli/variable"
	"github.com/kechako/envoke/config"
	"github.com/kechako/envoke/ent"
	envpred "github.com/kechako/envoke/ent/environment"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
)

const (
	appName    = "envoke"
	appVersion = "1.0.0"
)

func Main() {
	cmd := &cobra.Command{
		Use:   appName,
		Short: "A tool to manage environment variables",
		Long: `envoke is a CLI tool for efficiently managing environment variables across
multiple environments (development, staging, production, etc.).

You can create different environments, manage variables within each environment,
and run commands with environment-specific variables loaded.

Features:
  • Multiple environment configurations
  • Variable expansion with ${VAR} syntax
  • Import/export .env files
  • Global variables shared across environments
  • Run commands with environment variables loaded`,
		Version: appVersion,
		Example: `  # Create a development environment
  envoke env create development

  # Add environment variables
  envoke var add -e development DATABASE_URL "postgres://localhost/myapp_dev"
  envoke var add -e development API_KEY "dev-key-123"

  # Run commands with environment variables
  envoke run -e development npm start
  envoke run -e development ./migrate up

  # Export variables to .env file
  envoke var export -e development .env`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			configPath, _ := cmd.PersistentFlags().GetString("config")
			cfg, err := config.Load(ctx, configPath)
			if err != nil {
				return clierrors.Exit(err, 1)
			}
			ctx = config.NewContext(ctx, cfg)

			dbPath, err := cfg.GetDBPath()
			if err != nil {
				return clierrors.Exit(err, 1)
			}
			err = makeDataDir(dbPath)
			if err != nil {
				return clierrors.Exit(err, 1)
			}
			client, err := ent.Open("sqlite3", buildDataSourceName(dbPath))
			if err != nil {
				return clierrors.Exit(err, 1)
			}
			ctx = ent.NewContext(ctx, client)

			err = migrateDatabase(ctx, client)
			if err != nil {
				return clierrors.Exit(err, 1)
			}
			err = kitDatabase(ctx, client)
			if err != nil {
				return clierrors.Exit(err, 1)
			}

			cmd.SetContext(ctx)

			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			var errs []error

			client := ent.FromContext(ctx)
			if client != nil {
				if err := client.Close(); err != nil {
					errs = append(errs, fmt.Errorf("failed to close database connection: %w", err))
				}
			}

			switch len(errs) {
			case 0:
				return nil
			case 1:
				return clierrors.Exit(errs[0], 1)
			default:
				return clierrors.Exit(errors.Join(errs...), 1)
			}
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.PersistentFlags().StringP("config", "c", "", "Path to the configuration file")

	cmd.AddCommand(
		environment.Command(),
		variable.Command(),
		run.Command(),
	)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	err := cmd.ExecuteContext(ctx)
	if err != nil {
		code := 1

		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			code = exitErr.ExitCode()
		} else {
			var exitCoder clierrors.ExitCoder
			if errors.As(err, &exitCoder) {
				code = exitCoder.ExitCode()
			}

			if code == 0 {
				fmt.Fprintf(os.Stderr, "%v\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}
		}

		os.Exit(code)
	}
}

func makeDataDir(dbPath string) error {
	dir := filepath.Dir(dbPath)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create data directory %q: %w", dir, err)
	}
	return nil
}

func buildDataSourceName(dbPath string) string {
	if dbPath == "" {
		return "file::memory:?cache=shared"
	}

	query := url.Values{}
	query.Set("cache", "shared")
	query.Set("_fk", "1")

	sdn := &url.URL{
		Scheme:   "file",
		Opaque:   dbPath,
		RawQuery: query.Encode(),
	}

	return sdn.String()
}

func migrateDatabase(ctx context.Context, client *ent.Client) error {
	err := client.Schema.Create(ctx, schema.WithDropIndex(true), schema.WithDropColumn(true), schema.WithForeignKeys(true))
	if err != nil {
		return fmt.Errorf("failed to migrate database schema: %w", err)
	}

	return nil
}

func kitDatabase(ctx context.Context, client *ent.Client) error {
	err := client.DoTransaction(ctx, nil, func(ctx context.Context, tx *ent.Tx) error {
		{
			exists, err := tx.Environment.Query().Where(envpred.Name("global")).Exist(ctx)
			if err != nil {
				return fmt.Errorf("failed to check if global environment exists: %w", err)
			}
			if !exists {
				_, err := tx.Environment.Create().
					SetName("global").
					Save(ctx)
				if err != nil {
					return fmt.Errorf("failed to create global environment: %w", err)
				}
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to kit database: %w", err)
	}

	return nil
}
