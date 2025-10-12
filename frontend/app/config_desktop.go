//go:build !js || !wasm

package app

import (
	"fmt"

	"github.com/nishiki/frontend/config"
	"github.com/spf13/viper"
)

// LoadConfig loads configuration for desktop from filesystem
func LoadConfig() *config.Config {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Environment variable bindings
	viper.SetEnvPrefix("NISHIKI")
	viper.AutomaticEnv()

	fmt.Println("Loading config from filesystem for desktop")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Config file not found, using defaults: %v\n", err)
	}

	var cfg config.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		fmt.Printf("Error unmarshaling config: %v\n", err)
	}

	// Auto-generate redirect URL based on port if not explicitly set
	if cfg.RedirectURL == "" && cfg.Port != "" {
		cfg.RedirectURL = fmt.Sprintf("http://localhost:%s/auth/callback", cfg.Port)
	}

	// Debug: Print loaded config
	fmt.Printf("Loaded config: AuthURL=%s, ClientID=%s, RedirectURL=%s, Port=%s\n", 
		cfg.AuthURL, cfg.ClientID, cfg.RedirectURL, cfg.Port)

	return &cfg
}