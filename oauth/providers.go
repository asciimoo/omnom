package oauth

import "net/http"

type oauthProvider interface {
	GetRedirectURL(string, string) string
	GetTokenRequest(string, string, string, string) (*http.Request, error)
	GetUniqueUserID([]byte) (string, error)
	GetScope() (string, string)
}

var Providers = map[string]oauthProvider{
	"github": GitHubOAuth{
		AuthURL:  "https://github.com/login/oauth/authorize",
		TokenURL: "https://github.com/login/oauth/access_token",
	},
	"google": GoogleOAuth{
		AuthURL:  "https://accounts.google.com/o/oauth2/auth",
		TokenURL: "https://accounts.google.com/o/oauth2/token",
	},
}
