//go:build js && wasm

package app

import (
	_ "embed"
	"fmt"
	"strings"
	"syscall/js"

	"github.com/spf13/viper"

	"github.com/nishiki/frontend/config"
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

	// Auto-generate redirect URL from the browser's actual origin
	// This ensures the redirect URL matches the URL the user is accessing,
	// regardless of Docker port mapping or reverse proxy configuration
	if cfg.RedirectURL == "" {
		origin := js.Global().Get("window").Get("location").Get("origin").String()
		cfg.RedirectURL = origin + "/auth/callback"
	}

	// Debug: Print loaded config
	fmt.Printf("Loaded config: AuthURL=%s, ClientID=%s, RedirectURL=%s, Port=%s\n",
		cfg.AuthURL, cfg.ClientID, cfg.RedirectURL, cfg.Port)

	return &cfg
}
