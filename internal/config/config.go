package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds user preferences that persist between CLI invocations.
type Config struct {
	// DefaultAuthorName is pre-filled when rewriting author info.
	DefaultAuthorName string `mapstructure:"author_name"`
	// DefaultAuthorEmail similarly.
	DefaultAuthorEmail string `mapstructure:"author_email"`
	// BackupBranch controls whether we create a backup before destructive ops.
	BackupBranch bool `mapstructure:"backup_branch"`
	// Verbose globally enables verbose git output.
	Verbose bool `mapstructure:"verbose"`
}

// Load reads config from viper (which has already parsed file + flags + env).
func Load() (*Config, error) {
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("config parse error: %w", err)
	}
	return &cfg, nil
}

// Write persists a config struct back to the user's home config file.
func Write(cfg *Config) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	viper.Set("author_name", cfg.DefaultAuthorName)
	viper.Set("author_email", cfg.DefaultAuthorEmail)
	viper.Set("backup_branch", cfg.BackupBranch)

	configPath := filepath.Join(home, ".commitcraft.yaml")
	return viper.WriteConfigAs(configPath)
}
