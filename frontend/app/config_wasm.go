//go:build js && wasm

package app

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/nishiki/frontend/config"
	"github.com/spf13/viper"
)

//go:embed config/config.toml
var embeddedConfig []byte

// LoadConfig loads configuration for WebAssembly from embedded file
func LoadConfig() *config.Config {
	viper.SetConfigType("toml")

	// Environment variable bindings
	viper.SetEnvPrefix("NISHIKI")
	viper.AutomaticEnv()

	// Use embedded config for WebAssembly
	fmt.Println("Loading embedded config for WebAssembly")
	if err := viper.ReadConfig(strings.NewReader(string(embeddedConfig))); err != nil {
		fmt.Printf("Error reading embedded config: %v\n", err)
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
