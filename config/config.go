// Package config provides functionality to manage application configuration.
package config

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

type Config struct {
	DBPath string `yaml:"db_path"`
}

func (cfg *Config) validate() error {
	return nil
}

func (cfg *Config) GetDBPath() (string, error) {
	if cfg.DBPath != "" {
		return cfg.DBPath, nil
	}

	dataDir, err := dataDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dataDir, "data.db"), nil
}

func Load(ctx context.Context, name string) (*Config, error) {
	if name == "" {
		cfgDir, err := configDir()
		if err != nil {
			return nil, fmt.Errorf("unable to determine config directory: %w", err)
		}
		name = filepath.Join(cfgDir, "config.yaml")
	}

	file, err := os.Open(name)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("unable to open config file '%s': %w", name, err)
	}
	defer file.Close()

	return load(ctx, file)
}

func load(ctx context.Context, r io.Reader) (*Config, error) {
	var cfg Config
	err := yaml.NewDecoder(r).DecodeContext(ctx, &cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid config file format: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

func configDir() (string, error) {
	var err error
	dir := os.Getenv("XDG_DATA_HOME")
	if dir == "" {
		dir, err = os.UserHomeDir()
		if err != nil {
			return "", err
		}

		dir += "/.config"
	} else if !filepath.IsAbs(dir) {
		return "", errors.New("XDG_CONFIG_HOME must be an absolute path")
	}

	return filepath.Join(dir, "envoke"), nil
}

func dataDir() (string, error) {
	var err error
	dir := os.Getenv("XDG_DATA_HOME")
	if dir == "" {
		dir, err = os.UserHomeDir()
		if err != nil {
			return "", err
		}

		dir += "/.local/share"
	} else if !filepath.IsAbs(dir) {
		return "", errors.New("XDG_DATA_HOME must be an absolute path")
	}

	return filepath.Join(dir, "envoke"), nil
}
