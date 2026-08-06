package config

import (
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"path/filepath"
	"slices"

	"github.com/spf13/viper"
)

type Config struct {
	Port        int    `mapstructure:"port"`
	LogLevel    string `mapstructure:"log_level"`
	DatabaseURL string `mapstructure:"database_url"`
	SupabaseURL string `mapstructure:"supabase_url"`
}

var validLogLevels = []string{"debug", "info", "warn", "error"}

// Load resolves config from, in increasing order of precedence: defaults,
// config.yaml, .env, then the process environment. Both files are optional.
func Load(dir string) (*Config, error) {
	settings := viper.New()
	settings.SetDefault("port", 8080)
	settings.SetDefault("log_level", "info")
	settings.AutomaticEnv()

	// AutomaticEnv only resolves keys viper already knows, and these have no default.
	if err := settings.BindEnv("database_url"); err != nil {
		return nil, fmt.Errorf("bind DATABASE_URL: %w", err)
	}
	if err := settings.BindEnv("supabase_url"); err != nil {
		return nil, fmt.Errorf("bind SUPABASE_URL: %w", err)
	}
	if err := readOptionalYAML(settings, dir); err != nil {
		return nil, err
	}
	if err := mergeOptionalDotenv(settings, dir); err != nil {
		return nil, err
	}

	var config Config
	if err := settings.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	if err := config.validate(); err != nil {
		return nil, err
	}
	return &config, nil
}

func readOptionalYAML(settings *viper.Viper, dir string) error {
	settings.SetConfigName("config")
	settings.SetConfigType("yaml")
	settings.AddConfigPath(dir)

	var missing viper.ConfigFileNotFoundError
	if err := settings.ReadInConfig(); err != nil && !errors.As(err, &missing) {
		return fmt.Errorf("read config.yaml: %w", err)
	}
	return nil
}

func mergeOptionalDotenv(settings *viper.Viper, dir string) error {
	settings.SetConfigFile(filepath.Join(dir, ".env"))
	settings.SetConfigType("dotenv")

	var missing *fs.PathError
	if err := settings.MergeInConfig(); err != nil && !errors.As(err, &missing) {
		return fmt.Errorf("read .env: %w", err)
	}
	return nil
}

func (c *Config) validate() error {
	if c.DatabaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}
	parsed, err := url.Parse(c.DatabaseURL)
	if err != nil {
		return fmt.Errorf("DATABASE_URL is not a valid URL: %w", err)
	}
	if parsed.Scheme != "postgres" && parsed.Scheme != "postgresql" {
		return fmt.Errorf("DATABASE_URL must be a postgres:// URL, got scheme %q", parsed.Scheme)
	}
	if c.SupabaseURL == "" {
		return errors.New("SUPABASE_URL is required")
	}
	if u, err := url.Parse(c.SupabaseURL); err != nil || (u.Scheme != "https" && u.Scheme != "http") {
		return fmt.Errorf("SUPABASE_URL must be an http(s) URL, got %q", c.SupabaseURL)
	}
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", c.Port)
	}
	if !slices.Contains(validLogLevels, c.LogLevel) {
		return fmt.Errorf("log_level must be one of %v, got %q", validLogLevels, c.LogLevel)
	}
	return nil
}
