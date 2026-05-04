//go:build js && wasm

package config

import (
	_ "embed"
	"log/slog"
	"strings"
	"syscall/js"

	"github.com/spf13/viper"
)

//go:embed config.toml
var embeddedConfig []byte

// LoadConfig loads configuration from the embedded TOML (for WebAssembly builds).
func LoadConfig() *Config {
	viper.SetConfigType("toml")
	viper.SetEnvPrefix("NISHIKI")
	viper.AutomaticEnv()

	slog.Info("Loading embedded config for WebAssembly")
	if err := viper.ReadConfig(strings.NewReader(string(embeddedConfig))); err != nil {
		slog.Error("reading embedded config", "error", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		slog.Error("unmarshaling config", "error", err)
	}

	// Auto-generate redirect URL from the browser's actual origin so it matches
	// the URL the user is accessing regardless of port mapping or reverse proxy.
	if cfg.RedirectURL == "" {
		origin := js.Global().Get("window").Get("location").Get("origin").String()
		cfg.RedirectURL = origin + "/auth/callback"
	}

	slog.Info("Loaded config",
		"auth_url", cfg.AuthURL,
		"client_id", cfg.ClientID,
		"redirect_url", cfg.RedirectURL,
		"port", cfg.Port)

	return &cfg
}
