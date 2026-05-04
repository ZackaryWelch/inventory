//go:build !js || !wasm

package config

import (
	"fmt"
	"log/slog"

	"github.com/spf13/viper"
)

// LoadConfig loads configuration from filesystem (for desktop/serve commands).
func LoadConfig() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	viper.SetEnvPrefix("NISHIKI")
	viper.AutomaticEnv()

	slog.Info("Loading config from filesystem")
	if err := viper.ReadInConfig(); err != nil {
		slog.Warn("config file not found, using defaults", "error", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		slog.Error("unmarshaling config", "error", err)
	}

	if cfg.RedirectURL == "" && cfg.Port != "" {
		cfg.RedirectURL = fmt.Sprintf("http://localhost:%s/auth/callback", cfg.Port)
	}

	slog.Info("Loaded config",
		"auth_url", cfg.AuthURL,
		"client_id", cfg.ClientID,
		"redirect_url", cfg.RedirectURL,
		"port", cfg.Port)

	return &cfg
}
