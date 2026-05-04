package mcpserver

import (
	"encoding/json"
	"net/http"
	"strings"
)

// OAuthProtectedResource advertises OAuth 2.0 Protected Resource Metadata
// (RFC 9728) for the MCP endpoints and emits a WWW-Authenticate challenge
// on unauthenticated requests so that clients like mcp-remote can discover
// the authorization server automatically.
type OAuthProtectedResource struct {
	// Issuer is the authorization server issuer URL, e.g.
	// https://authentik.example.com/application/o/nishiki/ (trailing slash preserved).
	Issuer string
	// Scopes advertised to clients; defaults to the OIDC basics if empty.
	Scopes []string
}

// Wrap installs the discovery endpoints and auth challenge in front of next.
//
//   - GET /.well-known/oauth-protected-resource → RFC 9728 JSON metadata
//   - GET /.well-known/oauth-authorization-server → 307 redirect to the issuer's OIDC config
//   - requests without a bearer token → 401 with WWW-Authenticate pointing at the metadata URL
//   - otherwise pass through to next
func (o *OAuthProtectedResource) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/oauth-protected-resource":
			if r.Method == http.MethodGet {
				o.serveMetadata(w, r)
				return
			}
		case "/.well-known/oauth-authorization-server":
			if r.Method == http.MethodGet {
				http.Redirect(w, r,
					strings.TrimRight(o.Issuer, "/")+"/.well-known/openid-configuration",
					http.StatusTemporaryRedirect)
				return
			}
		}

		if !hasBearerToken(r) {
			w.Header().Set("WWW-Authenticate", `Bearer realm="mcp", resource_metadata="`+resourceMetadataURL(r)+`"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (o *OAuthProtectedResource) serveMetadata(w http.ResponseWriter, r *http.Request) {
	scopes := o.Scopes
	if len(scopes) == 0 {
		scopes = []string{"openid", "profile", "email"}
	}
	body := map[string]any{
		"resource":                 baseURL(r),
		"authorization_servers":    []string{o.Issuer},
		"bearer_methods_supported": []string{"header"},
		"scopes_supported":         scopes,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(body)
}

func hasBearerToken(r *http.Request) bool {
	if strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
		return true
	}
	return r.URL.Query().Get("token") != ""
}

func baseURL(r *http.Request) string {
	scheme := "http://"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https://"
	}
	return scheme + r.Host
}

func resourceMetadataURL(r *http.Request) string {
	return baseURL(r) + "/.well-known/oauth-protected-resource"
}
