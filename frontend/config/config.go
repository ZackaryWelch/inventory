package config

import (
	"fmt"

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

	fmt.Println("Loading config from filesystem")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Config file not found, using defaults: %v\n", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		fmt.Printf("Error unmarshaling config: %v\n", err)
	}

	// Auto-generate redirect URL based on port if not explicitly set
	if config.RedirectURL == "" && config.Port != "" {
		config.RedirectURL = fmt.Sprintf("http://localhost:%s/auth/callback", config.Port)
	}

	// Debug: Print loaded config
	fmt.Printf("Loaded config: AuthURL=%s, ClientID=%s, RedirectURL=%s, Port=%s\n",
		config.AuthURL, config.ClientID, config.RedirectURL, config.Port)

	return &config
}
