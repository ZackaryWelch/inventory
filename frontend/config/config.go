package config

import (
	"fmt"
	"log/slog"

	"github.com/spf13/viper"
)

// Config holds application configuration
type Config struct {
	BackendURL  string `mapstructure:"backend_url"`
	AuthURL     string `mapstructure:"auth_url"`
	ClientID    string `mapstructure:"client_id"`
	RedirectURL string `mapstructure:"redirect_url"`
	Port        string `mapstructure:"port"`
}

// LoadConfig loads configuration from filesystem (for desktop/serve commands)
func LoadConfig() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Environment variable bindings
	viper.SetEnvPrefix("NISHIKI")
	viper.AutomaticEnv()

	slog.Info("Loading config from filesystem")
	if err := viper.ReadInConfig(); err != nil {
		slog.Warn("config file not found, using defaults", "error", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		slog.Error("unmarshaling config", "error", err)
	}

	// Auto-generate redirect URL based on port if not explicitly set
	if config.RedirectURL == "" && config.Port != "" {
		config.RedirectURL = fmt.Sprintf("http://localhost:%s/auth/callback", config.Port)
	}

	slog.Info("Loaded config",
		"auth_url", config.AuthURL,
		"client_id", config.ClientID,
		"redirect_url", config.RedirectURL,
		"port", config.Port)

	return &config
}
