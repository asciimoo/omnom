package webapp

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/asciimoo/omnom/config"
	"github.com/asciimoo/omnom/model"

	"github.com/gin-gonic/gin"

	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/contrib/sessions"
)

const (
	SERVER_ADDR string = ":7331"
	SID         string = "sid"
)

var e *gin.Engine

var tplFuncMap = template.FuncMap{
	"HasPrefix": strings.HasPrefix,
	"ToHTML":    func(s string) template.HTML { return template.HTML(s) },
	"ToAttr":    func(s string) template.HTMLAttr { return template.HTMLAttr(s) },
	"ToURL":     func(s string) template.URL { return template.URL(s) },
	"ToDate":    func(t time.Time) string { return t.Format("2006-01-02") },
	"inc":       func(i int64) int64 { return i + 1 },
	"dec":       func(i int64) int64 { return i - 1 },
	"Truncate": func(s string, maxLen int) string {
		if len(s) > maxLen {
			return s[:maxLen] + "[..]"
		} else {
			return s
		}
	},
}

var bookmarksPerPage int64 = 20

func createRenderer() multitemplate.Renderer {
	r := multitemplate.DynamicRender{}
	r.AddFromFilesFuncs("index", tplFuncMap, "templates/layout/base.tpl", "templates/index.tpl")
	r.AddFromFilesFuncs("dashboard", tplFuncMap, "templates/layout/base.tpl", "templates/dashboard.tpl")
	r.AddFromFilesFuncs("signup", tplFuncMap, "templates/layout/base.tpl", "templates/signup.tpl")
	r.AddFromFilesFuncs("signup-confirm", tplFuncMap, "templates/layout/base.tpl", "templates/signup_confirm.tpl")
	r.AddFromFilesFuncs("login", tplFuncMap, "templates/layout/base.tpl", "templates/login.tpl")
	r.AddFromFilesFuncs("login-confirm", tplFuncMap, "templates/layout/base.tpl", "templates/login_confirm.tpl")
	r.AddFromFilesFuncs("bookmarks", tplFuncMap, "templates/layout/base.tpl", "templates/bookmarks.tpl")
	r.AddFromFilesFuncs("my-bookmarks", tplFuncMap, "templates/layout/base.tpl", "templates/my_bookmarks.tpl")
	r.AddFromFilesFuncs("profile", tplFuncMap, "templates/layout/base.tpl", "templates/profile.tpl")
	r.AddFromFilesFuncs("snapshotWrapper", tplFuncMap, "templates/layout/base.tpl", "templates/snapshot_wrapper.tpl")
	return r
}

func renderHTML(c *gin.Context, status int, page string, vars map[string]interface{}) {
	session := sessions.Default(c)
	u, _ := c.Get("user")
	tplVars := gin.H{
		"Page": page,
		"User": u,
	}
	sessChanged := false
	if s := session.Get("Error"); s != nil {
		tplVars["Error"] = s.(string)
		session.Delete("Error")
		sessChanged = true
	}
	if s := session.Get("Warning"); s != nil {
		tplVars["Warning"] = s.(string)
		session.Delete("Warning")
		sessChanged = true
	}
	if s := session.Get("Info"); s != nil {
		tplVars["Info"] = s.(string)
		session.Delete("Info")
		sessChanged = true
	}
	if sessChanged {
		session.Save()
	}
	for k, v := range vars {
		tplVars[k] = v
	}
	c.HTML(status, page, tplVars)
}

func Run(cfg *config.Config) {
	e = gin.Default()
	if !cfg.App.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	if cfg.App.BookmarksPerPage > 0 {
		bookmarksPerPage = cfg.App.BookmarksPerPage
	}
	e.Use(sessions.Sessions("SID", sessions.NewCookieStore([]byte("secret"))))
	e.Use(SessionMiddleware())
	e.Use(ConfigMiddleware(cfg))
	authorized := e.Group("/")
	authorized.Use(authRequired)

	bu := cfg.Server.BaseURL
	if bu == "" {
		bu = fmt.Sprintf("http://%s/", cfg.Server.Address)
	}
	baseURL, err := url.Parse(bu)
	if err != nil {
		panic(err)
	}
	tplFuncMap["BaseURL"] = func(u string) string {
		b, err := url.Parse(u)
		if err != nil {
			return ""
		}
		return baseURL.ResolveReference(b).String()
	}
	e.HTMLRender = createRenderer()

	// ROUTES
	e.Static("/static", "./static")
	e.GET("/", index)
	e.GET("/signup", signup)
	e.POST("/signup", signup)
	e.GET("/login", login)
	e.POST("/login", login)
	e.GET("/logout", logout)
	e.GET("/bookmarks", bookmarks)
	e.GET("/snapshot", snapshotWrapper)
	e.GET("/viewSnapshot", snapshot)
	e.POST("/add_bookmark", addBookmark)

	authorized.GET("/profile", profile)
	authorized.GET("/generate_addon_token", generateAddonToken)
	authorized.GET("/delete_addon_token", deleteAddonToken)
	authorized.GET("/my_bookmarks", myBookmarks)

	log.Println("Starting server")
	e.Run(cfg.Server.Address)
}

func index(c *gin.Context) {
	if u, ok := c.Get("user"); ok && u != nil {
		dashboard(c, u.(*model.User))
		return
	}
	renderHTML(c, http.StatusOK, "index", nil)
}

func authRequired(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(SID)
	if user == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}
	c.Next()
}

func getPageno(c *gin.Context) int64 {
	var pageno int64 = 1
	if pagenoStr, ok := c.GetQuery("pageno"); ok {
		if userPageno, err := strconv.Atoi(pagenoStr); err == nil && userPageno > 0 {
			pageno = int64(userPageno)
		}
	}
	return pageno
}

func SessionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		uname := session.Get(SID)
		if uname != nil {
			c.Set("user", model.GetUser(uname.(string)))
		} else {
			c.Set("user", nil)
		}
		c.Next()
	}
}

func ConfigMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("config", cfg)
		c.Next()
	}
}
