package oauth

type oauthProvider interface {
	GetRedirectURL(string, string) string
	GetTokenURL(string, string, string) string
	GetUniqueUserID([]byte) (string, error)
}

var Providers = map[string]oauthProvider{
	"github": GitHubOAuth{
		AuthURL:  "https://github.com/login/oauth/authorize",
		TokenURL: "https://github.com/login/oauth/access_token",
	},
}
