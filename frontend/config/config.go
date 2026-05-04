package config

// Config holds application configuration
type Config struct {
	BackendURL  string `mapstructure:"backend_url"`
	AuthURL     string `mapstructure:"auth_url"`
	ClientID    string `mapstructure:"client_id"`
	RedirectURL string `mapstructure:"redirect_url"`
	Port        string `mapstructure:"port"`
}
