package utils_config

import (
	"context"
)

// ConfigChangeCallback is called when configuration changes
type ConfigChangeCallback func(oldConfig, newConfig *Config) error

// ConfigLoader interface for dependency injection
type ConfigLoader interface {
	Load(ctx context.Context, configPath string) (*Config, error)
	Reload(ctx context.Context) error
	Watch(ctx context.Context) error
	GetConfig() *Config
	RegisterCallback(callback ConfigChangeCallback)
	Validate() error
	Stop() error
}