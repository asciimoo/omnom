// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

package webapp

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/asciimoo/omnom/config"

	"github.com/gin-gonic/gin"
)

func oauthHandler(c *gin.Context) {
	cfg, _ := c.Get("config")
	providers := cfg.(*config.Config).OAuth
	provider, found := providers[c.Query("provider")]
	if !found {
		setNotification(c, nError, "Unknown OAuth provider", false)
		c.Redirect(http.StatusFound, URLFor("Login"))
		return
	}

	params := url.Values{}
	params.Add("client_id", provider.ClientID)
	params.Add("scope", strings.Join(provider.Scopes, ","))
	params.Add("response_type", "code")
	params.Add("redirect_uri", getFullURLPrefix(c)+URLFor("Oauth verification")+"?provider="+c.Query("provider"))

	reqURL := fmt.Sprintf("%s?%s", provider.AuthURL, params.Encode())

	log.Println("New OAUTH verification:", reqURL, provider)
	c.Redirect(http.StatusFound, reqURL)
}

func oauthRedirectHandler(c *gin.Context) {
	cfg, _ := c.Get("config")
	providers := cfg.(*config.Config).OAuth
	provider, found := providers[c.Query("provider")]
	if !found {
		setNotification(c, nError, "Failed OAuth login", false)
		c.Redirect(http.StatusFound, URLFor("Login"))
		return
	}
	params := url.Values{}
	reqURL := fmt.Sprintf("%s?%s", provider.TokenURL, params.Encode())
	log.Println("New OAUTH verification:", reqURL)
	// TODO
}
