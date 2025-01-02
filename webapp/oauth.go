// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

package webapp

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/asciimoo/omnom/config"
	"github.com/asciimoo/omnom/model"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

type oauthProvider interface {
	GetRedirectURL(string, string) string
	GetTokenURL(string, string, string) string
	GetUniqueUserID([]byte) (string, error)
}

var oauthProviders = map[string]oauthProvider{
	"github": GitHubOAuth{
		AuthURL:  "https://github.com/login/oauth/authorize",
		TokenURL: "https://github.com/login/oauth/access_token",
		Scope:    "read:user",
	},
}

func oauthHandler(c *gin.Context) {
	cfg, _ := c.Get("config")
	pCfgs := cfg.(*config.Config).OAuth
	pCfg, cf := pCfgs[c.Query("provider")]
	p, pf := oauthProviders[c.Query("provider")]

	if !cf || !pf {
		setNotification(c, nError, "Unknown OAuth provider", false)
		c.Redirect(http.StatusFound, URLFor("Login"))
		return
	}

	session := sessions.Default(c)
	tok := model.GenerateToken()
	session.Set("oauth_token", tok)
	err := session.Save()
	if err != nil {
		setNotification(c, nError, "Session persist error", false)
		c.Redirect(http.StatusFound, URLFor("Login"))
		return
	}

	handlerURL := fmt.Sprintf("%s?provider=%s&state=%s", getFullURLPrefix(c)+URLFor("Oauth verification"), c.Query("provider"), tok)

	reqURL := p.GetRedirectURL(pCfg.ClientID, handlerURL)

	c.Redirect(http.StatusFound, reqURL)
}

func oauthRedirectHandler(c *gin.Context) {
	cfg, _ := c.Get("config")
	pCfgs := cfg.(*config.Config).OAuth
	pCfg, cf := pCfgs[c.Query("provider")]
	p, pf := oauthProviders[c.Query("provider")]

	if !cf || !pf {
		setNotification(c, nError, "Unknown OAuth provider", false)
		c.Redirect(http.StatusFound, URLFor("Login"))
		return
	}

	oauthToken := c.Query("state")
	session := sessions.Default(c)
	if t := session.Get("oauth_token"); t != nil {
		tok, _ := t.(string)
		if tok != oauthToken {
			setNotification(c, nError, "Invalid OAuth response", false)
			log.Println("OAuth handler: token mismatch ")
			c.Redirect(http.StatusFound, URLFor("Login"))
			return
		}
	} else {
		setNotification(c, nError, "Invalid OAuth response", false)
		log.Println("OAuth handler: missing token")
		c.Redirect(http.StatusFound, URLFor("Login"))
		return
	}

	code := c.Query("code")
	reqURL := p.GetTokenURL(pCfg.ClientID, pCfg.ClientSecret, code)

	resp, err := http.Get(reqURL) // #nosec G107
	if err != nil {
		setNotification(c, nError, "Invalid OAuth response", false)
		log.Println("OAuth handler http response error:", err)
		c.Redirect(http.StatusFound, URLFor("Login"))
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		setNotification(c, nError, "Invalid OAuth response", false)
		log.Println("OAuth handler response error:", err)
		c.Redirect(http.StatusFound, URLFor("Login"))
		return
	}

	uid, err := p.GetUniqueUserID(body)
	if err != nil {
		setNotification(c, nError, "Invalid OAuth response", false)
		log.Println("OAuth provider response parse error:", err)
		c.Redirect(http.StatusFound, URLFor("Login"))
		return
	}

	u := model.GetUserByOAuthID(uid)
	if u == nil {
		session.Set("oauth_id", uid)
		err = session.Save()
		if err != nil {
			setNotification(c, nError, err.Error(), false)
			c.Redirect(http.StatusFound, URLFor("Login"))
			return
		}
		c.Redirect(http.StatusFound, URLFor("Signup"))
		return
	}

	session.Set(SID, u.Username)
	err = session.Save()
	if err != nil {
		setNotification(c, nError, err.Error(), false)
		c.Redirect(http.StatusFound, URLFor("Login"))
		return
	}
	c.Redirect(http.StatusFound, baseURL("/"))
}

type GitHubOAuth struct {
	AuthURL  string
	TokenURL string
	Scope    string
}

func (g GitHubOAuth) GetRedirectURL(clientID, handlerURL string) string {
	params := url.Values{}
	params.Add("client_id", clientID)
	params.Add("scope", g.Scope)
	params.Add("response_type", "code")
	params.Add("redirect_uri", handlerURL)
	return fmt.Sprintf("%s?%s", g.AuthURL, params.Encode())
}

func (g GitHubOAuth) GetTokenURL(clientID, clientSecret, code string) string {
	params := url.Values{}
	params.Add("client_id", clientID)
	params.Add("client_secret", clientSecret)
	params.Add("code", code)
	return fmt.Sprintf("%s?%s", g.TokenURL, params.Encode())
}

func (g GitHubOAuth) GetUniqueUserID(body []byte) (string, error) {
	v, err := url.ParseQuery(string(body))
	if err != nil {
		return "", err
	}

	t := v.Get("access_token")
	if t == "" {
		return "", errors.New("no access token found")
	}

	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "bearer "+t)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	uBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	log.Println("BODY:", string(uBody))
	var j map[string]interface{}

	err = json.Unmarshal(uBody, &j)
	if err != nil {
		return "", err
	}

	l, ok := j["login"].(string)
	if !ok {
		return "", errors.New("failed to get user login data")
	}

	return fmt.Sprintf("gh-%s", l), nil
}
