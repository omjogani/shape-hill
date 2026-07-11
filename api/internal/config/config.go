package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Port        int    `mapstructure:"port"`
	LogLevel    string `mapstructure:"log_level"`
	DatabaseURL string `mapstructure:"database_url"`
}

// Load reads config.yaml, then overrides from the environment (PORT, LOG_LEVEL,
// DATABASE_URL). Secrets live only in env — never in config.yaml.
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	// AutomaticEnv only sees keys viper already knows about, so declare the
	// env-only ones by binding them explicitly.
	if err := viper.BindEnv("database_url"); err != nil {
		return nil, err
	}

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var c Config
	if err := viper.Unmarshal(&c); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	if c.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}
	return &c, nil
}
