//go:build !js || !wasm

package app

import "github.com/nishiki/frontend/config"

// LoadConfig loads configuration from the filesystem for desktop builds.
func LoadConfig() *Config {
	return config.LoadConfig()
}
