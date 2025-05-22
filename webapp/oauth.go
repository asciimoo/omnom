// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

package webapp

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/asciimoo/omnom/config"
	"github.com/asciimoo/omnom/model"
	"github.com/asciimoo/omnom/oauth"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func oauthHandler(c *gin.Context) {
	cfg, _ := c.Get("config")
	pCfgs := cfg.(*config.Config).OAuth
	pCfg, cf := pCfgs[c.Query("provider")]
	p, pf := oauth.Providers[c.Query("provider")]

	if !cf || !pf {
		setNotification(c, nError, "Unknown OAuth provider", false)
		c.Redirect(http.StatusFound, URLFor("Login"))

		return
	}

	session := sessions.Default(c)
	token := model.GenerateToken()

	session.Set("oauth_token", token)
	if session.Save() != nil {
		setNotification(c, nError, "Session persist error", false)
		c.Redirect(http.StatusFound, URLFor("Login"))

		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	req := oauth.NewPrepareRequest(pCfg.ConfigurationURL)
	if err := p.Prepare(ctx, req); err != nil {
		setNotification(c, nError, err.Error(), true)
		c.Redirect(http.StatusFound, URLFor("Login"))

		return
	}

	sName, sValue := p.GetScope()
	reqURI := oauth.NewRedirectURIRequest(
		pCfg.ClientID,
		fmt.Sprintf("%s?provider=%s", getFullURL(c, URLFor("Oauth verification")), c.Query("provider")),
	)

	c.Redirect(http.StatusFound, fmt.Sprintf("%s&%s=%s&state=%s", p.GetRedirectURL(reqURI), sName, sValue, token))
}

func oauthRedirectHandler(c *gin.Context) {
	cfg, _ := c.Get("config")
	pCfgs := cfg.(*config.Config).OAuth
	pCfg, cf := pCfgs[c.Query("provider")]
	p, pf := oauth.Providers[c.Query("provider")]

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
			log.Error().Msg("OAuth handler: token mismatch")
			c.Redirect(http.StatusFound, URLFor("Login"))

			return
		}
	} else {
		setNotification(c, nError, "Invalid OAuth response", false)
		log.Error().Msg("OAuth handler: missing token")
		c.Redirect(http.StatusFound, URLFor("Login"))

		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	redirectURI := fmt.Sprintf("%s?provider=%s", getFullURL(c, URLFor("Oauth verification")), c.Query("provider"))
	code := c.Query("code")

	resp, err := p.GetToken(ctx, oauth.NewTokenRequest(pCfg.ClientID, pCfg.ClientSecret, code, redirectURI))
	if err != nil {
		setNotification(c, nError, "Invalid OAuth response", true)
		log.Error().Err(err).Msg("Invalid OAuth HTTP response")
		c.Redirect(http.StatusFound, URLFor("Login"))

		return
	}
	defer resp.Body.Close()

	tokenBody, err := io.ReadAll(resp.Body)
	if err != nil {
		setNotification(c, nError, "Invalid OAuth response", false)
		log.Error().Err(err).Msg("Invalid OAuth response body")
		c.Redirect(http.StatusFound, URLFor("Login"))

		return
	}

	userInfo, err := p.GetUserInfo(ctx, tokenBody)
	if err != nil {
		setNotification(c, nError, "Invalid OAuth response", false)
		log.Error().Err(err).Msg("Invalid OAuth response data")
		c.Redirect(http.StatusFound, URLFor("Login"))

		return
	}

	u := model.GetUserByOAuthID(userInfo.UID)
	if u == nil {
		session.Set("oauth_id", userInfo.UID)
		session.Set("oauth_email", userInfo.Email)
		session.Set("oauth_username", userInfo.Username)

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
