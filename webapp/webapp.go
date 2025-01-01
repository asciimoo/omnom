// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

package webapp

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/asciimoo/omnom/config"
	"github.com/asciimoo/omnom/model"

	"github.com/gin-gonic/gin"

	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/contrib/sessions"
)

type NotificationType int

const (
	ServerAddr string = ":7331"
	SID        string = "sid"
)

const (
	nInfo NotificationType = iota
	nError
)

var e *gin.Engine
var baseURL func(string) string
var URLFor func(string) string

var tplFuncMap = template.FuncMap{
	"HasPrefix":  strings.HasPrefix,
	"ToHTML":     func(s string) template.HTML { return template.HTML(s) },         //nolint: gosec // HTML is well formed.
	"ToAttr":     func(s string) template.HTMLAttr { return template.HTMLAttr(s) }, //nolint: gosec // HTML is well formed.
	"ToURL":      func(s string) template.URL { return template.URL(s) },           //nolint: gosec // HTML is well formed.
	"ToDate":     func(t time.Time) string { return t.Format("2006-01-02") },
	"ToDateTime": func(t time.Time) string { return t.Format("2006-01-02 15:04:05") },
	"Replace":    strings.ReplaceAll,
	"ToLower":    strings.ToLower,
	"inc":        func(i int64) int64 { return i + 1 },
	"dec":        func(i int64) int64 { return i - 1 },
	"SnapshotURL": func(key string) string {
		return fmt.Sprintf("%s%s/%s.gz", baseURL("/static/data/snapshots/"), key[:2], key)
	},
	"AddURLParam": func(base string, param string) string {
		if strings.Contains(base, "?") {
			return base + "&" + param
		}
		return base + "?" + param
	},
	"Truncate": func(s string, maxLen int) string {
		if len(s) > maxLen {
			return s[:maxLen] + "[..]"
		} else {
			return s
		}
	},
	"KVData": func(values ...interface{}) (map[string]interface{}, error) {
		if len(values)%2 != 0 {
			return nil, errors.New("invalid dict call")
		}
		dict := make(map[string]interface{}, len(values)/2)
		for i := 0; i < len(values); i += 2 {
			key, ok := values[i].(string)
			if !ok {
				return nil, errors.New("dict keys must be strings")
			}
			dict[key] = values[i+1]
		}
		return dict, nil
	},
	"FormatSize": formatSize,
}

var resultsPerPage int64 = 20

func getFullURLPrefix(c *gin.Context) string {
	fullURLPrefix := ""
	if strings.HasPrefix(baseURL("/"), "/") {
		fullURLPrefix = "http://"
		if c.Request.TLS != nil {
			fullURLPrefix = "https://"
		}
		fullURLPrefix += c.Request.Host
	}
	return fullURLPrefix
}

func addTemplate(r multitemplate.DynamicRender, rootDir string, hasBase bool, name, filename string) {
	if hasBase {
		r.AddFromFilesFuncs(name, tplFuncMap, filepath.Join(rootDir, "layout/base.tpl"), filepath.Join(rootDir, filename))
	} else {
		r.AddFromFilesFuncs(name, tplFuncMap, filepath.Join(rootDir, filename))
	}
}

func createRenderer(rootDir string) multitemplate.Renderer {
	r := multitemplate.DynamicRender{}
	addTemplate(r, rootDir, true, "index", "index.tpl")
	addTemplate(r, rootDir, true, "dashboard", "dashboard.tpl")
	addTemplate(r, rootDir, true, "signup", "signup.tpl")
	addTemplate(r, rootDir, true, "signup-confirm", "signup_confirm.tpl")
	addTemplate(r, rootDir, true, "login", "login.tpl")
	addTemplate(r, rootDir, true, "login-confirm", "login_confirm.tpl")
	addTemplate(r, rootDir, true, "bookmarks", "bookmarks.tpl")
	addTemplate(r, rootDir, true, "snapshots", "snapshots.tpl")
	addTemplate(r, rootDir, true, "my-bookmarks", "my_bookmarks.tpl")
	addTemplate(r, rootDir, true, "profile", "profile.tpl")
	addTemplate(r, rootDir, true, "snapshotWrapper", "snapshot_wrapper.tpl")
	addTemplate(r, rootDir, true, "view-bookmark", "view_bookmark.tpl")
	addTemplate(r, rootDir, true, "edit-bookmark", "edit_bookmark.tpl")
	addTemplate(r, rootDir, true, "api", "api.tpl")
	addTemplate(r, rootDir, true, "error", "error.tpl")
	addTemplate(r, rootDir, false, "rss", "rss.xml")
	return r
}

func render(c *gin.Context, status int, page string, vars map[string]interface{}) {
	session := sessions.Default(c)
	u, _ := c.Get("user")
	cfg, _ := c.Get("config")
	csrf, _ := c.Get("_csrf")
	tplVars := gin.H{
		"Page":          page,
		"User":          u,
		"DisableSignup": cfg.(*config.Config).App.DisableSignup,
		"CSRF":          csrf,
		"OAuth":         cfg.(*config.Config).OAuth,
	}
	sessChanged := false
	if s := session.Get("Error"); s != nil {
		tplVars["Error"], _ = s.(string)
		session.Delete("Error")
		sessChanged = true
	}
	if s := session.Get("Warning"); s != nil {
		tplVars["Warning"], _ = s.(string)
		session.Delete("Warning")
		sessChanged = true
	}
	if s := session.Get("Info"); s != nil {
		tplVars["Info"], _ = s.(string)
		session.Delete("Info")
		sessChanged = true
	}
	if s, ok := c.Get("Error"); ok {
		tplVars["Error"], _ = s.(string)
	}
	if s, ok := c.Get("Warning"); ok {
		tplVars["Warning"], _ = s.(string)
	}
	if s, ok := c.Get("Info"); ok {
		tplVars["Info"], _ = s.(string)
	}
	if sessChanged {
		err := session.Save()
		if err != nil {
			_ = c.Error(fmt.Errorf("error saving context: %w", err))
		}
	}
	fullURL := baseURL(c.FullPath())
	if c.Request.URL.RawQuery != "" {
		fullURL += "?" + c.Request.URL.RawQuery
	}
	tplVars["URL"] = fullURL
	for k, v := range vars {
		tplVars[k] = v
	}
	switch c.Query("format") {
	case "json":
		renderJSON(c, status, tplVars)
		return
	case "rss":
		renderRSS(c, status, tplVars)
		return
	}
	c.HTML(status, page, tplVars)
}

func renderJSON(c *gin.Context, status int, vars map[string]interface{}) {
	delete(vars, "CSRF")
	delete(vars, "DisableSignup")
	c.IndentedJSON(status, vars)
}

func renderRSS(c *gin.Context, status int, vars map[string]interface{}) {
	k, ok := c.Get("RSS")
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"Title":   "Not found.",
			"Message": "This page does not exist.",
		})
		return
	}
	c.Header("Content-Type", "application/rss+xml; charset=utf-8")
	tplVars := map[string]interface{}{
		"RSS":  vars[k.(string)],
		"Type": k.(string),
	}
	tplVars["FullURLPrefix"] = getFullURLPrefix(c)
	c.HTML(status, "rss", tplVars)
}

func registerEndpoint(r *gin.RouterGroup, e *Endpoint) {
	var h gin.HandlerFunc
	if e.RSS != "" {
		h = RSSEndpointWrapper(e.Handler, e.RSS)
	} else {
		h = e.Handler
	}
	switch e.Method {
	case GET:
		r.GET(e.Path, h)
	case POST:
		r.POST(e.Path, h)
	case PUT:
		r.PUT(e.Path, h)
	case PATCH:
		r.PATCH(e.Path, h)
	case HEAD:
		r.HEAD(e.Path, h)
	}
}

func Run(cfg *config.Config) {
	e = gin.Default()
	if !cfg.App.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	if cfg.App.ResultsPerPage > 0 {
		resultsPerPage = cfg.App.ResultsPerPage
	}
	_ = e.SetTrustedProxies([]string{"127.0.0.1"})
	sess := sessions.NewCookieStore([]byte("secret"))
	sess.Options(sessions.Options{
		Secure: cfg.Server.SecureCookie,
	})
	e.Use(sessions.Sessions("SID", sess))
	e.Use(SessionMiddleware())
	e.Use(ConfigMiddleware(cfg))
	e.Use(CSRFMiddleware())
	e.Use(ErrorLoggerMiddleware())
	e.Use(GzipMiddleware())
	authorized := e.Group("/")
	authorized.Use(authRequiredMiddleware)

	bu := cfg.Server.BaseURL
	baseURL = func(u string) string {
		if strings.HasPrefix(u, "/") && strings.HasSuffix(bu, "/") {
			u = u[1:]
		}
		return bu + u
	}
	// TODO handle GET arguments as well
	URLFor = func(e string) string {
		for _, ep := range Endpoints {
			if ep.Name == e {
				return baseURL(ep.Path)
			}
		}
		log.Printf("Error: Endpoint '%s' not found", e)
		return "/"
	}
	tplFuncMap["BaseURL"] = baseURL
	tplFuncMap["URLFor"] = URLFor
	e.HTMLRender = createRenderer(cfg.App.TemplateDir)

	// ROUTES
	e.Static("/static", cfg.App.StaticDir)
	for _, ep := range Endpoints {
		if ep.AuthRequired {
			registerEndpoint(authorized, ep)
		} else {
			registerEndpoint(&e.RouterGroup, ep)
		}
	}
	e.NoRoute(notFoundView)

	log.Println("Starting server")
	err := e.Run(cfg.Server.Address)
	if err != nil {
		log.Printf("Error running server: %+v\n", err)
	}
}

func notFoundView(c *gin.Context) {
	render(c, http.StatusNotFound, "error", gin.H{
		"Title":   "Not found.",
		"Message": "This page does not exist.",
	})
}

func index(c *gin.Context) {
	if u, ok := c.Get("user"); ok && u != nil && u.(*model.User) != nil {
		dashboard(c, u.(*model.User))
		return
	}
	render(c, http.StatusOK, "index", nil)
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

func authRequiredMiddleware(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(SID)
	if user == nil {
		setNotification(c, nError, "Unauthorized", false)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}
	if u, _ := c.Get("user"); u.(*model.User) == nil {
		setNotification(c, nError, "Unauthorized", false)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}
	c.Next()
}

func SessionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		uname := session.Get(SID)
		if uname != nil {
			u := model.GetUser(uname.(string))
			c.Set("user", u)
		} else {
			tok := c.PostForm("token")
			if tok == "" {
				tok = c.Query("token")
			}
			if tok != "" {
				// can be nil in case of invalid token
				c.Set("user", model.GetUserBySubmissionToken(tok))
			}
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

func CSRFMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasSuffix(c.HandlerName(), ".addBookmark") || strings.HasSuffix(c.HandlerName(), ".addResource") {
			c.Next()
			return
		}
		newToken := model.GenerateToken()
		c.Set("_csrf", newToken)
		session := sessions.Default(c)
		prevToken := session.Get("_csrf")
		session.Set("_csrf", newToken)
		err := session.Save()
		if err != nil {
			_ = c.Error(fmt.Errorf("error saving context: %w", err))
		}
		if c.Request.Method != POST {
			c.Next()
			return
		}
		uname := session.Get(SID)
		if uname != nil {
			if t := c.Request.FormValue("_csrf"); t == "" || prevToken != t {
				tok := c.PostForm("token")
				if tok == "" {
					tok = c.Query("token")
				}
				u := model.GetUserBySubmissionToken(tok)
				if u == nil {
					_, _ = gin.DefaultWriter.Write([]byte("\033[31m[ERROR] CSRF token mismatch\033[0m\n"))
					c.String(400, "CSRF token mismatch")
					c.Abort()
					return
				}
			}
		}
	}
}

func ErrorLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		err, ok := c.Get("Error")
		if ok {
			_, _ = gin.DefaultWriter.Write([]byte(fmt.Sprintf("\033[31m[ERROR] %s\033[0m\n", err)))
		}
	}
}

func GzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.Contains(c.Request.URL.Path, "/static/data/") {
			if strings.Contains(c.Request.URL.Path, "/static/data/snapshots") {
				c.Header("Content-Type", "text/html; charset=utf-8")
			}
			c.Header("Content-Encoding", "gzip")
		}
		c.Next()
	}
}

func formatSize(s uint) string {
	if s > 1024*1024*1024*1024 {
		return fmt.Sprintf("%.2fTb", float64(s)/(1024*1024*1024*1024))
	}
	if s > 1024*1024*1024 {
		return fmt.Sprintf("%.2fGb", float64(s)/(1024*1024*1024))
	}
	if s > 1024*1024 {
		return fmt.Sprintf("%.2fMb", float64(s)/(1024*1024))
	}
	if s > 1024 {
		return fmt.Sprintf("%.2fKb", float64(s)/1024)
	}
	return fmt.Sprintf("%.2fb", float64(s))
}

func setNotification(c *gin.Context, t NotificationType, n string, persist bool) {
	session := sessions.Default(c)
	if persist {
		defer func() {
			_ = session.Save()
		}()
	}
	switch t {
	case nInfo:
		c.Set("Info", n)
		if persist {
			session.Set("Info", n)
		}
	case nError:
		c.Set("Error", n)
		if persist {
			session.Set("Error", n)
		}
	}
}

func RSSEndpointWrapper(f gin.HandlerFunc, rssVar string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("RSS", rssVar)
		f(c)
	}
}
