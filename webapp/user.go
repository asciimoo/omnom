package webapp

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/asciimoo/omnom/config"
	"github.com/asciimoo/omnom/mail"
	"github.com/asciimoo/omnom/model"

	"github.com/gin-gonic/gin"

	"github.com/gin-gonic/contrib/sessions"
)

var userRe = regexp.MustCompile(`[a-zA-Z0-9_]+`)

func signup(c *gin.Context) {
	cfg, _ := c.Get("config")
	if cfg.(*config.Config).App.DisableSignup {
		return
	}
	if c.Request.Method == http.MethodPost {
		username := c.PostForm("username")
		email := c.PostForm("email")
		if username == "" || email == "" {
			setNotification(c, nError, "Missing email", false)
			renderHTML(c, http.StatusOK, "signup", nil)
			return
		}
		if strings.ToLower(username) == "admin" {
			setNotification(c, nError, "Reserved username", false)
			renderHTML(c, http.StatusOK, "signup", nil)
			return
		}
		if match := userRe.MatchString(username); !match {
			setNotification(c, nError, "Invalid username. Use only letters, numbers and underscore", false)
			renderHTML(c, http.StatusOK, "signup", nil)
			return
		}
		u := model.GetUser(username)
		if u != nil {
			setNotification(c, nError, "Username already exists", false)
			renderHTML(c, http.StatusOK, "signup", nil)
			return
		}
		err := model.CreateUser(username, email)
		if err != nil {
			setNotification(c, nError, err.Error(), false)
			renderHTML(c, http.StatusOK, "signup", nil)
			return
		}
		u = model.GetUser(username)
		err = mail.Send(
			u.Email,
			"Successful registration to Omnom",
			"login",
			map[string]interface{}{
				"Token":    u.LoginToken,
				"Username": u.Username,
				"BaseURL":  baseURL("/login"),
			},
		)
		if err != nil {
			setNotification(c, nError, err.Error(), false)
			renderHTML(c, http.StatusOK, "signup", nil)
			return
		}
		renderHTML(c, http.StatusOK, "signup-confirm", nil)
		return
	}
	renderHTML(c, http.StatusOK, "signup", nil)
}

func login(c *gin.Context) {
	uname, ok := c.GetPostForm("username")
	if ok {
		u := model.GetUser(uname)
		if u == nil {
			setNotification(c, nError, "Unknown user", false)
			renderHTML(c, http.StatusOK, "login", nil)
			return
		}
		u.LoginToken = model.GenerateToken()
		err := model.DB.Save(u).Error
		if err != nil {
			setNotification(c, nError, err.Error(), false)
			renderHTML(c, http.StatusOK, "login", nil)
			return
		}
		err = mail.Send(
			u.Email,
			fmt.Sprintf("Omnom login verification for %s", u.Username),
			"login",
			map[string]interface{}{
				"Token":    u.LoginToken,
				"Username": u.Username,
				"BaseURL":  baseURL("/login"),
			},
		)
		if err != nil {
			setNotification(c, nError, err.Error(), false)
			renderHTML(c, http.StatusOK, "login", nil)
			return
		}
		log.Println("New login token generated:", u.LoginToken)
		renderHTML(c, http.StatusOK, "login-confirm", nil)
		return
	}

	tok, ok := c.GetQuery("token")
	if ok && tok != "" {
		u := model.GetUserByLoginToken(tok)
		if u == nil {
			setNotification(c, nError, "Unknown user", false)
			renderHTML(c, http.StatusOK, "login", nil)
			return
		}
		u.LoginToken = ""
		err := model.DB.Save(u).Error
		if err != nil {
			setNotification(c, nError, err.Error(), false)
			renderHTML(c, http.StatusOK, "login", nil)
			return
		}
		session := sessions.Default(c)
		session.Set(SID, u.Username)
		err = session.Save()
		if err != nil {
			setNotification(c, nError, err.Error(), false)
			renderHTML(c, http.StatusOK, "login", nil)
			return
		}
		c.Redirect(http.StatusFound, baseURL("/"))
		return
	}
	renderHTML(c, http.StatusOK, "login", nil)
}

func logout(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(SID)
	if user == nil {
		c.Redirect(http.StatusFound, baseURL("/"))
		return
	}
	session.Delete(SID)
	defer func() {
		_ = session.Save()
	}()
	c.Redirect(http.StatusFound, baseURL("/"))
}

func profile(c *gin.Context) {
	u, _ := c.Get("user")
	tplData := map[string]interface{}{}
	if u == nil {
		c.Redirect(http.StatusFound, baseURL("/"))
		return
	}
	uid := u.(*model.User).ID
	if c.Request.Method == "POST" {
		var ts []*model.Token
		err := model.DB.Where("user_id = ?", uid).Find(&ts).Error
		if err != nil {
			setNotification(c, nError, err.Error(), false)
		}
		tplData["AddonTokens"] = ts
		tplData["DisplayTokens"] = true
	}
	var sSize int64
	model.DB.
		Model(&model.Snapshot{}).
		Select("sum(snapshots.size)").
		Joins("join bookmarks on bookmarks.id = snapshots.bookmark_id").
		Joins("join users on bookmarks.user_id = users.id").
		Group("users.id").
		Where("users.id = ?", uid).First(&sSize)
	tplData["SnapshotsSize"] = uint(sSize)
	renderHTML(c, http.StatusOK, "profile", tplData)
}

func generateAddonToken(c *gin.Context) {
	u, _ := c.Get("user")
	tok := &model.Token{
		Text:   model.GenerateToken(),
		UserID: u.(*model.User).ID,
	}
	err := model.DB.Create(tok).Error
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
