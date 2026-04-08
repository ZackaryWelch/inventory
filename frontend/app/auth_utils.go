package app

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"golang.org/x/oauth2"
)

func newOAuth2Config(config *Config) *oauth2.Config {
	return &oauth2.Config{
		ClientID:    config.ClientID,
		RedirectURL: config.RedirectURL,
		Scopes:      []string{"openid", "profile", "email", "groups", "offline_access"},
		Endpoint: oauth2.Endpoint{
			AuthURL:   config.AuthURL + "/application/o/authorize/",
			TokenURL:  config.BackendURL + "/auth/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}
}

func generateRandomString(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		fmt.Printf("Warning: crypto/rand failed, using fallback: %v\n", err)
		return fmt.Sprintf("%d_%d", time.Now().UnixNano(), length)
	}
	return base64.RawURLEncoding.EncodeToString(bytes)
}

func generateCodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}
