// Package oauth provides OAuth 2.0 authentication support for external providers.
//
// This package implements OAuth 2.0 authentication flows for multiple identity providers:
//   - GitHub
//   - Google
//   - Generic OIDC (OpenID Connect) providers
//
// Each provider implements a common OAuth interface that handles the authorization
// code flow. The package supports:
//   - Authorization URL generation with PKCE
//   - Token exchange
//   - User profile fetching
//   - Automatic provider discovery via OIDC
//
// OAuth providers are configured in the application config file with client IDs,
// secrets, and scopes. Users can sign in using any configured provider, and their
// OAuth ID is linked to their Omnom account.
//
// The Providers map contains factory functions for creating provider instances:
//
//	provider := oauth.Providers["github"](oauthConfig)
//	authURL := provider.GetAuthURL(state)
//
// Example configuration:
//
//	oauth:
//	  github:
//	    client_id: "your-client-id"
//	    client_secret: "your-client-secret"
//	    scopes: ["user:email"]
package oauth

import (
	"context"
	"net/http"
)

const (
	responseTypeCode           ResponseType = "code"
	grantTypeAuthorizationCode GrantType    = "authorization_code"
	scopeName                  ScopeName    = "scope"
)

type (
	// ScopeName represents the name of an OAuth scope parameter.
	ScopeName string
	// ScopeValue represents the value of an OAuth scope.
	ScopeValue string
	// ResponseType represents the OAuth response type (e.g., "code").
	ResponseType string
	// GrantType represents the OAuth grant type (e.g., "authorization_code").
	GrantType string

	// TokenResponse contains the raw token response from an OAuth provider.
	TokenResponse []byte

	// PrepareRequest contains parameters for preparing an OAuth provider.
	PrepareRequest struct {
		configurationURL string
	}

	// RedirectURIRequest contains parameters for building an OAuth redirect URI.
	RedirectURIRequest struct {
		clientID    string
		redirectURI string
	}

	// TokenRequest contains parameters for exchanging an authorization code for a token.
	TokenRequest struct {
		clientID     string
		clientSecret string
		code         string
		redirectURI  string
	}

	// UserInfoResponse contains user information retrieved from an OAuth provider.
	UserInfoResponse struct {
		UID      string
		Email    string
		Username string
	}
)

// String returns the string representation of ScopeName.
func (sn ScopeName) String() string {
	return string(sn)
}

// String returns the string representation of ScopeValue.
func (sv ScopeValue) String() string {
	return string(sv)
}

// String returns the string representation of ResponseType.
func (rt ResponseType) String() string {
	return string(rt)
}

// String returns the string representation of GrantType.
func (gt GrantType) String() string {
	return string(gt)
}

// NewPrepareRequest creates a new PrepareRequest with the given configuration URL.
func NewPrepareRequest(cURL string) *PrepareRequest {
	return &PrepareRequest{configurationURL: cURL}
}

// NewRedirectURIRequest creates a new RedirectURIRequest with the given parameters.
func NewRedirectURIRequest(clientID string, redirectURI string) *RedirectURIRequest {
	return &RedirectURIRequest{clientID: clientID, redirectURI: redirectURI}
}

// NewTokenRequest creates a new TokenRequest with the given parameters.
func NewTokenRequest(clientID string, clientSecret string, code string, redirectURI string) *TokenRequest {
	return &TokenRequest{clientID: clientID, clientSecret: clientSecret, code: code, redirectURI: redirectURI}
}

type oauthProvider interface {
	Prepare(context.Context, *PrepareRequest) error
	GetRedirectURL(*RedirectURIRequest) string
	GetToken(context.Context, *TokenRequest) (*http.Response, error)
	GetUserInfo(context.Context, TokenResponse) (*UserInfoResponse, error)
	GetScope() (ScopeName, ScopeValue)
}

type (
	tokenData struct {
		AccessToken string `json:"access_token"`
	}

	userData struct {
		Email             string `json:"email"`              // all
		Login             string `json:"login"`              // github
		Name              string `json:"name"`               // google
		PreferredUsername string `json:"preferred_username"` // oidc
	}
)

// Providers maps provider names to their OAuth implementation instances.
// Supported providers: github, google, oidc.
var Providers = map[string]oauthProvider{
	"github": GitHubOAuth{
		AuthURL:  "https://github.com/login/oauth/authorize",
		TokenURL: "https://github.com/login/oauth/access_token",
	},
	"google": GoogleOAuth{
		AuthURL:  "https://accounts.google.com/o/oauth2/auth",
		TokenURL: "https://accounts.google.com/o/oauth2/token",
	},
	"oidc": &OIDCOAuth{},
}
