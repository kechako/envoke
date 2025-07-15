package config

import "context"

type contextKey struct{}

func NewContext(parent context.Context, cfg *Config) context.Context {
	return context.WithValue(parent, contextKey{}, cfg)
}

func FromContext(ctx context.Context) *Config {
	cfg, _ := ctx.Value(contextKey{}).(*Config)
	return cfg
}
