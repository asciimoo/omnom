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
	ClientID         string
	ClientSecret     string
	ConfigurationURL string
	AuthURL          string
	TokenURL         string
	UserInfoURL      string
	Icon             string
	ScopeName        string
	ScopeValue       string
	Code             string
	ResponseType     string
	GrantType        string
	RedirectURI      string
)

func (c ClientID) String() string {
	return string(c)
}

func (c ClientSecret) String() string {
	return string(c)
}

func (c ConfigurationURL) String() string {
	return string(c)
}

func (a AuthURL) String() string {
	return string(a)
}

func (t TokenURL) String() string {
	return string(t)
}

func (ui UserInfoURL) String() string {
	return string(ui)
}

func (sn ScopeName) String() string {
	return string(sn)
}

func (sv ScopeValue) String() string {
	return string(sv)
}

func (c Code) String() string {
	return string(c)
}

func (rt ResponseType) String() string {
	return string(rt)
}

func (gt GrantType) String() string {
	return string(gt)
}

func (ri RedirectURI) String() string {
	return string(ri)
}

type oauthProvider interface {
	Prepare(context.Context, ConfigurationURL) error
	GetRedirectURL(ClientID, RedirectURI) string
	GetTokenRequest(context.Context, ClientID, ClientSecret, Code, RedirectURI) (*http.Request, error)
	GetUniqueUserID(context.Context, []byte) (string, error)
	GetScope() (ScopeName, ScopeValue)
}

type (
	tokenData struct {
		AccessToken string `json:"access_token"`
	}

	userData struct {
		Email string `json:"email"`
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
