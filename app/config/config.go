package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `toml:"server" mapstructure:"server"`
	Database DatabaseConfig `toml:"database" mapstructure:"database"`
	Auth     AuthConfig     `toml:"auth" mapstructure:"auth"`
	Logging  LoggingConfig  `toml:"logging" mapstructure:"logging"`
}

type ServerConfig struct {
	Port  int       `toml:"port" mapstructure:"port"`
	Debug bool      `toml:"debug" mapstructure:"debug"`
	TLS   TLSConfig `toml:"tls" mapstructure:"tls"`
}

type TLSConfig struct {
	Enabled  bool   `toml:"enabled" mapstructure:"enabled"`
	CertFile string `toml:"cert_file" mapstructure:"cert_file"`
	KeyFile  string `toml:"key_file" mapstructure:"key_file"`
}

type DatabaseConfig struct {
	URI      string `toml:"uri" mapstructure:"uri"`
	Database string `toml:"database" mapstructure:"database"`
	Timeout  int    `toml:"timeout" mapstructure:"timeout"`
}

type OAuthClient struct {
	ProviderName string `toml:"provider_name" mapstructure:"provider_name"`
	ClientID     string `toml:"client_id" mapstructure:"client_id"`
	ClientSecret string `toml:"client_secret" mapstructure:"client_secret"`
	RedirectURL  string `toml:"redirect_url" mapstructure:"redirect_url"`
}

type AuthConfig struct {
	AuthentikURL      string        `toml:"authentik_url" mapstructure:"authentik_url"`
	Clients           []OAuthClient `toml:"clients" mapstructure:"clients"`
	JWKSCacheDuration int           `toml:"jwks_cache_duration" mapstructure:"jwks_cache_duration"`
	AllowSelfSigned   bool          `toml:"allow_self_signed" mapstructure:"allow_self_signed"`
	APIToken          string        `toml:"api_token" mapstructure:"api_token"`
}

type LoggingConfig struct {
	Level       string `toml:"level" mapstructure:"level"`
	SeqEndpoint string `toml:"seq_endpoint" mapstructure:"seq_endpoint"`
	SeqAPIKey   string `toml:"seq_api_key" mapstructure:"seq_api_key"`
}

func Load() (*Config, error) {
	v := viper.New()

	// Set config name and paths
	v.SetConfigName("app")
	v.SetConfigType("toml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("/etc/nishiki/")

	// Set defaults
	setDefaults(v)

	// Enable environment variable support
	v.SetEnvPrefix("NISHIKI")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found, continue with defaults and env vars
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	if err := validate(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.port", 3001)
	v.SetDefault("server.debug", false)
	v.SetDefault("server.tls.enabled", true)
	v.SetDefault("server.tls.cert_file", "./certs/server.crt")
	v.SetDefault("server.tls.key_file", "./certs/server.key")

	// Database defaults
	v.SetDefault("database.uri", "mongodb://localhost:27017")
	v.SetDefault("database.database", "nishiki")
	v.SetDefault("database.timeout", 10)

	// Auth defaults
	v.SetDefault("auth.authentik_url", "")
	v.SetDefault("auth.jwks_cache_duration", 300)
	v.SetDefault("auth.allow_self_signed", false)
	v.SetDefault("auth.api_token", "")
	v.SetDefault("auth.clients", []OAuthClient{})

	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.seq_endpoint", "")
	v.SetDefault("logging.seq_api_key", "")
}

func validate(config *Config) error {
	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		return fmt.Errorf("server port must be between 1 and 65535")
	}

	if config.Database.URI == "" {
		return fmt.Errorf("database URI is required")
	}

	if config.Database.Database == "" {
		return fmt.Errorf("database name is required")
	}

	if config.Auth.AuthentikURL == "" {
		return fmt.Errorf("authentik URL is required")
	}

	if len(config.Auth.Clients) == 0 {
		return fmt.Errorf("at least one OAuth client must be configured")
	}

	// Validate each OAuth client
	providerNames := make(map[string]bool)
	for i, client := range config.Auth.Clients {
		if client.ProviderName == "" {
			return fmt.Errorf("client %d: provider name is required", i)
		}

		if providerNames[client.ProviderName] {
			return fmt.Errorf("duplicate client provider name: %s", client.ProviderName)
		}
		providerNames[client.ProviderName] = true

		if client.ClientID == "" {
			return fmt.Errorf("client %s: client ID is required", client.ProviderName)
		}

		if client.ClientSecret == "" {
			return fmt.Errorf("client %s: client secret is required", client.ProviderName)
		}

		if client.RedirectURL == "" {
			return fmt.Errorf("client %s: redirect URL is required", client.ProviderName)
		}
	}

	return nil
}
