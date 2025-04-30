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
	ScopeName    string
	ScopeValue   string
	ResponseType string
	GrantType    string

	TokenResponse []byte

	PrepareRequest struct {
		configurationURL string
	}

	RedirectURIRequest struct {
		clientID    string
		redirectURI string
	}

	TokenRequest struct {
		clientID     string
		clientSecret string
		code         string
		redirectURI  string
	}

	UserInfoResponse struct {
		UID      string
		Email    string
		Username string
	}
)

func (sn ScopeName) String() string {
	return string(sn)
}

func (sv ScopeValue) String() string {
	return string(sv)
}

func (rt ResponseType) String() string {
	return string(rt)
}

func (gt GrantType) String() string {
	return string(gt)
}

func NewPrepareRequest(cURL string) *PrepareRequest {
	return &PrepareRequest{configurationURL: cURL}
}

func NewRedirectURIRequest(clientID string, redirectURI string) *RedirectURIRequest {
	return &RedirectURIRequest{clientID: clientID, redirectURI: redirectURI}
}

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
