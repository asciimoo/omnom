package webapp

import (
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/asciimoo/omnom/mail"
	"github.com/asciimoo/omnom/model"

	"github.com/gin-gonic/gin"

	"github.com/gin-gonic/contrib/sessions"
)

var userRe = regexp.MustCompile(`[a-zA-Z0-9_]+`)

func signup(c *gin.Context) {
	if c.Request.Method == "POST" {
		username := c.PostForm("username")
		email := c.PostForm("email")
		if username == "" || email == "" {
			renderHTML(c, http.StatusOK, "signup", map[string]interface{}{
				"Error": "Missing data",
			})
			return
		}
		if strings.ToLower(username) == "admin" {
			renderHTML(c, http.StatusOK, "signup", map[string]interface{}{
				"Error": "Reserved username",
			})
			return
		}
		if match := userRe.MatchString(username); !match {
			renderHTML(c, http.StatusOK, "signup", map[string]interface{}{
				"Error": "Invalid username. Use only letters, numbers and underscore",
			})
			return
		}
		u := model.GetUser(username)
		if u != nil {
			renderHTML(c, http.StatusOK, "signup", map[string]interface{}{
				"Error": "Username already exists",
			})
			return
		}
		err := model.CreateUser(username, email)
		if err != nil {
			renderHTML(c, http.StatusOK, "signup", map[string]interface{}{
				"Error": err.Error(),
			})
			return
		}
		u = model.GetUser(username)
		err = mail.Send(u.Email, "login", map[string]interface{}{
			"Token":   u.LoginToken,
			"BaseURL": baseURL("/login"),
		})
		if err != nil {
			renderHTML(c, http.StatusOK, "signup", map[string]interface{}{
				"Error": err.Error(),
			})
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
			renderHTML(c, http.StatusOK, "login", map[string]interface{}{
				"Error": "Unknown user",
			})
			return
		}
		u.LoginToken = model.GenerateToken()
		err := model.DB.Save(u).Error
		if err != nil {
			renderHTML(c, http.StatusOK, "login", map[string]interface{}{
				"Error": err.Error(),
			})
			return
		}
		err = mail.Send(u.Email, "login", map[string]interface{}{
			"Token":   u.LoginToken,
			"BaseURL": baseURL("/login"),
		})
		if err != nil {
			renderHTML(c, http.StatusOK, "login", map[string]interface{}{
				"Error": err.Error(),
			})
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
			renderHTML(c, http.StatusOK, "login", map[string]interface{}{
				"Error": "Invalid token",
			})
			return
		}
		u.LoginToken = ""
		err := model.DB.Save(u).Error
		if err != nil {
			renderHTML(c, http.StatusOK, "login", map[string]interface{}{
				"Error": err.Error(),
			})
			return
		}
		session := sessions.Default(c)
		session.Set(SID, u.Username)
		err = session.Save()
		if err != nil {
			renderHTML(c, http.StatusOK, "login", map[string]interface{}{
				"Error": err.Error(),
			})
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
	session.Save()
	c.Redirect(http.StatusFound, baseURL("/"))
}

func profile(c *gin.Context) {
	u, _ := c.Get("user")
	tplData := map[string]interface{}{}
	if u == nil {
		c.Redirect(http.StatusFound, baseURL("/"))
		return
	}
	var ts []*model.Token
	err := model.DB.Where("user_id = ?", u.(*model.User).ID).Find(&ts).Error
	if err != nil {
		tplData["Error"] = err.Error()
	}
	tplData["AddonTokens"] = ts
	renderHTML(c, http.StatusOK, "profile", tplData)
}

func generateAddonToken(c *gin.Context) {
	session := sessions.Default(c)
	u, _ := c.Get("user")
	tok := &model.Token{
		Text:   model.GenerateToken(),
		UserID: u.(*model.User).ID,
	}
	err := model.DB.Create(tok).Error
	if err != nil {
		session.Set("Error", err.Error())
	} else {
		session.Set("Info", "Token created")
	}
	session.Save()
	c.Redirect(http.StatusFound, baseURL("/profile"))
}

func deleteAddonToken(c *gin.Context) {
	session := sessions.Default(c)
	id := c.PostForm("id")
	u, _ := c.Get("user")
	err := model.DB.Where("user_id = ? AND id = ?", u.(*model.User).ID, id).Delete(&model.Token{}).Error
	if err != nil {
		session.Set("Error", err.Error())
	} else {
		session.Set("Info", "Token deleted")
	}
	session.Save()
	c.Redirect(http.StatusFound, baseURL("/profile"))
}
