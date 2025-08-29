// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

package webapp

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/asciimoo/omnom/config"
	"github.com/asciimoo/omnom/mail"
	"github.com/asciimoo/omnom/model"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

var userRe = regexp.MustCompile(`[a-zA-Z0-9_]+`)

func isActivityPubHeader(s string) bool {
	return strings.HasPrefix(s, "application/activity+json") || strings.HasPrefix(s, `application/ld+json; profile="https://www.w3.org/ns/activitystreams"`)
}

func userProfile(c *gin.Context) {
	user := model.GetUser(c.Param("username"))
	if user == nil {
		setNotification(c, nError, "Unknown user", false)
		log.Debug().Str("username", c.Param("username")).Msg("Unknown user")
		notFoundView(c)
		return
	}

	if isActivityPubHeader(c.Request.Header.Get("Accept")) {
		apIdentityResponse(c, user)
		return
	}

	u, err := url.Parse(getFullURL(c, "/"))
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse URL")
		notFoundView(c)
		return
	}
	fName := fmt.Sprintf("%s@%s", user.Username, u.Host)
	var bc int64
	if err := model.DB.Model(&model.Bookmark{}).Where("bookmarks.user_id = ? and bookmarks.public = 1", user.ID).Count(&bc).Error; err != nil {
		log.Error().Err(err).Msg("Failed to count bookmarks")
	}
	var fc int64
	if err := model.DB.Model(&model.APFollower{}).Where("ap_followers.user_id = ?", user.ID).Count(&fc).Error; err != nil {
		log.Error().Err(err).Msg("Failed to count bookmarks")
	}
	render(c, http.StatusOK, "user", gin.H{
		"User":          user,
		"FediName":      fName,
		"BookmarkCount": bc,
		"FollowerCount": fc,
	})
}

func validateUsername(username string) error {
	if strings.ToLower(username) == "admin" {
		return errors.New("reserved username")
	}
	if match := userRe.MatchString(username); !match {
		return errors.New("invalid username. Use only letters, numbers and underscore")
	}
	u := model.GetUser(username)
	if u != nil {
		return errors.New("username already exists")
	}
	return nil
}

func signup(c *gin.Context) {
	cfg, _ := c.Get("config")
	if cfg.(*config.Config).App.DisableSignup {
		c.Redirect(http.StatusFound, baseURL("/"))
	}

	session := sessions.Default(c)
	tplVars := map[string]any{
		"OAuthID":       session.Get("oauth_id"),
		"OAuthEmail":    session.Get("oauth_email"),
		"OAuthUsername": session.Get("oauth_username"),
	}

	if c.Request.Method == http.MethodPost {
		username := c.PostForm("username")
		email := c.PostForm("email")
		if username == "" || email == "" {
			setNotification(c, nError, "Missing email", false)
			render(c, http.StatusOK, "signup", tplVars)
			return
		}
		err := validateUsername(username)
		if err != nil {
			setNotification(c, nError, err.Error(), false)
			render(c, http.StatusOK, "signup", tplVars)
			return
		}
		err = model.CreateUser(username, email)
		if err != nil {
			setNotification(c, nError, err.Error(), false)
			render(c, http.StatusOK, "signup", tplVars)
			return
		}
		u := model.GetUser(username)
		if session.Get("oauth_id") != nil {
			oauthID, _ := session.Get("oauth_id").(string)
			u.OAuthID = &oauthID
			err = model.DB.Save(u).Error
			if err != nil {
				setNotification(c, nError, err.Error(), false)
				render(c, http.StatusOK, "signup", tplVars)
				return
			}
			session.Delete("oauth_id")
			session.Delete("oauth_email")
			session.Delete("oauth_username")

			if err := session.Save(); err != nil {
				setNotification(c, nError, err.Error(), false)
				render(c, http.StatusOK, "signup", tplVars)
				return
			}

			setNotification(c, nInfo, "Successful registration", false)
			c.Redirect(http.StatusFound, baseURL("/"))

			return
		} else {
			err = mail.Send(
				*u.Email,
				"Successful registration to Omnom",
				"login",
				map[string]any{
					"Token":    u.LoginToken,
					"Username": u.Username,
					"URL":      fmt.Sprintf("%s?token=%s", getFullURL(c, URLFor("Login")), u.LoginToken),
				},
			)
			if err != nil {
				setNotification(c, nError, err.Error(), false)
				render(c, http.StatusOK, "signup", tplVars)
				return
			}
		}
		render(c, http.StatusOK, "signup-confirm", nil)
		return
	}
	render(c, http.StatusOK, "signup", tplVars)
}

func login(c *gin.Context) {
	cfg, _ := c.Get("config")
	if cfg.(*config.Config).Server.RemoteUserHeader != "" {
		c.Redirect(http.StatusFound, baseURL("/"))
	}
	uname, ok := c.GetPostForm("username")
	if ok {
		u := model.GetUser(uname)
		if u == nil {
			setNotification(c, nError, "Unknown user", false)
			render(c, http.StatusOK, "login", nil)
			return
		}
		u.LoginToken = model.GenerateToken()
		err := model.DB.Save(u).Error
		if err != nil {
			setNotification(c, nError, err.Error(), false)
			render(c, http.StatusOK, "login", nil)
			return
		}
		if u.Email != nil {
			err = mail.Send(
				*u.Email,
				fmt.Sprintf("Omnom login verification for %s", u.Username),
				"login",
				map[string]any{
					"Token":    u.LoginToken,
					"Username": u.Username,
					"URL":      fmt.Sprintf("%s?token=%s", getFullURL(c, URLFor("Login")), u.LoginToken),
				},
			)
			if err != nil {
				setNotification(c, nError, err.Error(), false)
				render(c, http.StatusOK, "login", nil)
				return
			}
		}
		log.Debug().Str("username", u.Username).Msgf("Visit %s?token=%s to sign in", getFullURL(c, "login"), u.LoginToken)
		render(c, http.StatusOK, "login-confirm", nil)
		return
	}

	tok, ok := c.GetQuery("token")
	if ok && tok != "" {
		u := model.GetUserByLoginToken(tok)
		if u == nil {
			setNotification(c, nError, "Unknown user", false)
			render(c, http.StatusOK, "login", nil)
			return
		}
		u.LoginToken = ""
		err := model.DB.Save(u).Error
		if err != nil {
			setNotification(c, nError, err.Error(), false)
			render(c, http.StatusOK, "login", nil)
			return
		}
		session := sessions.Default(c)
		session.Set(SID, u.Username)
		err = session.Save()
		if err != nil {
			setNotification(c, nError, err.Error(), false)
			render(c, http.StatusOK, "login", nil)
			return
		}
		c.Redirect(http.StatusFound, baseURL("/"))
		return
	}
	render(c, http.StatusOK, "login", nil)
}

func logout(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(SID)
	if user == nil {
		c.Redirect(http.StatusFound, baseURL("/"))
		return
	}
	session.Delete(SID)
	_ = session.Save()
	c.Redirect(http.StatusFound, baseURL("/"))
}

func profile(c *gin.Context) {
	u, _ := c.Get("user")
	tplData := map[string]any{}
	if u == nil {
		c.Redirect(http.StatusFound, baseURL("/"))
		return
	}
	uid := u.(*model.User).ID
	if c.Request.Method == POST {
		var ts []*model.Token
		err := model.DB.Where("user_id = ?", uid).Find(&ts).Error
		if err != nil {
			setNotification(c, nError, err.Error(), false)
		}
		tplData["AddonTokens"] = ts
		tplData["DisplayTokens"] = true
	}
	var sSize uint
	model.DB.
		Model(&model.Snapshot{}).
		Select("sum(snapshots.size)").
		Joins("join bookmarks on bookmarks.id = snapshots.bookmark_id").
		Joins("join users on bookmarks.user_id = users.id").
		Group("users.id").
		Where("users.id = ?", uid).First(&sSize)
	tplData["SnapshotsSize"] = sSize
	render(c, http.StatusOK, "profile", tplData)
}

func generateAddonToken(c *gin.Context) {
	u, _ := c.Get("user")
	tok, err := model.CreateAddonToken(u.(*model.User).ID)
	if err != nil {
		setNotification(c, nError, err.Error(), true)
	} else {
		setNotification(c, nInfo, "Token created: "+tok.Text, true)
	}
	c.Redirect(http.StatusFound, baseURL("/profile"))
}

func deleteAddonToken(c *gin.Context) {
	id := c.PostForm("id")
	u, _ := c.Get("user")
	err := model.DB.Where("user_id = ? AND id = ?", u.(*model.User).ID, id).Delete(&model.Token{}).Error
	if err != nil {
		setNotification(c, nError, err.Error(), true)
	} else {
		setNotification(c, nInfo, "Token deleted", true)
	}
	c.Redirect(http.StatusFound, baseURL("/profile"))
}

func checkAddonToken(c *gin.Context) {
	tok := c.PostForm("token")
	var t model.Token
	err := model.DB.Where("text = ?", tok).First(&t).Error
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"message": "Invalid token. Check your addon tokens on your profile page of the webapp.",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "success",
	})
}
