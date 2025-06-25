// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

package webapp

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/asciimoo/omnom/config"
	"github.com/asciimoo/omnom/localization"
	"github.com/asciimoo/omnom/model"
	"github.com/asciimoo/omnom/static"
	"github.com/asciimoo/omnom/storage"
	"github.com/asciimoo/omnom/templates"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"filippo.io/csrf"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
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

var baseURL func(string) string
var URLFor func(string, ...string) string

var tplFuncMap = template.FuncMap{
	"HasPrefix":  strings.HasPrefix,
	"ToHTML":     func(s string) template.HTML { return template.HTML(s) },         //nolint: gosec // HTML is well formed.
	"ToAttr":     func(s string) template.HTMLAttr { return template.HTMLAttr(s) }, //nolint: gosec // HTML is well formed.
	"ToURL":      func(s string) template.URL { return template.URL(s) },           //nolint: gosec // HTML is well formed.
	"ToDate":     func(t time.Time) string { return t.Format("2006-01-02") },
	"ToDateTime": func(t time.Time) string { return t.Format("2006-01-02 15:04:05") },
	"Replace":    strings.ReplaceAll,
	"ToLower":    strings.ToLower,
	"Capitalize": strings.Title,
	"inc":        func(i int64) int64 { return i + 1 },
	"dec":        func(i int64) int64 { return i - 1 },
	"SnapshotURL": func(key string) string {
		return fmt.Sprintf("%s%s/%s.gz", baseURL("/static/data/snapshots/"), key[:2], key)
	},
	"AddURLParam": addURLParam,
	"Truncate":    truncate,
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
	"ResourceURL": func(s string) string {
		u, full := storage.GetResourceURL(s)
		if full {
			return u
		}
		return baseURL(u)
	},
	"HasAttr": func(v interface{}, name string) bool {
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
		}
		if rv.Kind() != reflect.Struct {
			return false
		}
		return rv.FieldByName(name).IsValid()
	},
}

var resultsPerPage int64 = 20

func addURLParam(base string, param string) string {
	if strings.Contains(base, "?") {
		u, err := url.Parse(base)
		if err != nil {
			return base + "&" + param
		}
		kv := strings.SplitN(param, "=", 2)
		q := u.Query()
		q.Set(kv[0], kv[1])
		u.RawQuery = q.Encode()
		return u.String()
	}
	return base + "?" + param
}

func getFullURLPrefix(c *gin.Context) string {
	ccfg, _ := c.Get("config")
	cfg := ccfg.(*config.Config)
	if cfg.Server.BaseURL != "" {
		return cfg.Server.BaseURL
	}
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

func getFullURL(c *gin.Context, u string) string {
	if strings.HasPrefix(u, "/") {
		return getFullURLPrefix(c) + u
	}
	return u
}

func truncate(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen] + "[..]"
	} else {
		return s
	}
}

func addTemplate(r multitemplate.DynamicRender, root fs.FS, hasBase bool, name, filename string) {
	if hasBase {
		r.AddFromFSFuncs(name, tplFuncMap, root, "layout/base.tpl", filename)
	} else {
		r.AddFromFSFuncs(name, tplFuncMap, root, filename)
	}
}

func createRenderer(tplFS fs.FS) multitemplate.Renderer {
	r := multitemplate.DynamicRender{}
	addTemplate(r, tplFS, true, "index", "index.tpl")
	addTemplate(r, tplFS, true, "dashboard", "dashboard.tpl")
	addTemplate(r, tplFS, true, "signup", "signup.tpl")
	addTemplate(r, tplFS, true, "signup-confirm", "signup_confirm.tpl")
	addTemplate(r, tplFS, true, "login", "login.tpl")
	addTemplate(r, tplFS, true, "login-confirm", "login_confirm.tpl")
	addTemplate(r, tplFS, true, "bookmarks", "bookmarks.tpl")
	addTemplate(r, tplFS, true, "snapshots", "snapshots.tpl")
	addTemplate(r, tplFS, true, "my-bookmarks", "my_bookmarks.tpl")
	addTemplate(r, tplFS, true, "profile", "profile.tpl")
	addTemplate(r, tplFS, true, "snapshot-wrapper", "snapshot_wrapper.tpl")
	addTemplate(r, tplFS, true, "snapshot-details", "snapshot_details.tpl")
	addTemplate(r, tplFS, true, "view-bookmark", "view_bookmark.tpl")
	addTemplate(r, tplFS, true, "edit-bookmark", "edit_bookmark.tpl")
	addTemplate(r, tplFS, true, "create-bookmark", "create_bookmark.tpl")
	addTemplate(r, tplFS, true, "snapshot-diff-form", "snapshot_diff_form.tpl")
	addTemplate(r, tplFS, true, "snapshot-diff", "snapshot_diff.tpl")
	addTemplate(r, tplFS, true, "snapshot-diff-side-by-side", "snapshot_diff_side_by_side.tpl")
	addTemplate(r, tplFS, true, "edit-collection", "edit_collection.tpl")
	addTemplate(r, tplFS, true, "feeds", "feeds.tpl")
	addTemplate(r, tplFS, true, "edit-feed", "edit_feed.tpl")
	addTemplate(r, tplFS, true, "user", "user.tpl")
	addTemplate(r, tplFS, true, "api", "api.tpl")
	addTemplate(r, tplFS, true, "error", "error.tpl")
	addTemplate(r, tplFS, false, "rss", "rss.xml")
	return r
}

func render(c *gin.Context, status int, page string, vars map[string]interface{}) {
	session := sessions.Default(c)
	u, _ := c.Get("user")
	cfg, _ := c.Get("config")
	l, _ := c.Get("localizer")
	tplVars := gin.H{
		"Page":                  page,
		"User":                  u,
		"DisableSignup":         cfg.(*config.Config).App.DisableSignup,
		"AllowBookmarkCreation": cfg.(*config.Config).App.CreateBookmarkFromWebapp,
		"OAuth":                 cfg.(*config.Config).OAuth,
		"Tr":                    l.(*localization.Localizer),
		"FullURL": func(u string) string {
			return getFullURL(c, u)
		},
	}
	sessChanged := false
	if c.Query("theme") == "dark" || c.Query("theme") == "light" {
		session.Set("Theme", c.Query("theme"))
		sessChanged = true
	}
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
	if s := session.Get("Theme"); s != nil {
		tplVars["Theme"], _ = s.(string)
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
	allowManualLogin := true
	if cfg.(*config.Config).Server.RemoteUserHeader != "" {
		allowManualLogin = false
	}
	tplVars["AllowManualLogin"] = allowManualLogin
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
	delete(vars, "DisableSignup")
	delete(vars, "OAuth")
	delete(vars, "Tr")
	delete(vars, "FullURL")
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
		"FullURL": func(u string) string {
			return getFullURL(c, u)
		},
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

func resolveDynamicPath(p string, v []string) string {
	if len(v) == 0 {
		return p
	}
	if !strings.Contains(p, ":") {
		return p
	}
	pParts := strings.Split(p, "/")
	vRef := 0
	for i, f := range pParts {
		if vRef >= len(v) {
			break
		}
		if strings.HasPrefix(f, ":") {
			pParts[i] = v[vRef]
			vRef += 1
		}
	}
	return strings.Join(pParts, "/")
}

func createEngine(cfg *config.Config) *gin.Engine {
	e := gin.Default()
	store := cookie.NewStore([]byte("secret"))
	store.Options(sessions.Options{
		Secure: cfg.Server.SecureCookie,
	})
	if cfg.App.ResultsPerPage > 0 {
		resultsPerPage = cfg.App.ResultsPerPage
	}
	_ = e.SetTrustedProxies([]string{"127.0.0.1"})
	e.Use(sessions.Sessions("SID", store))
	e.Use(SessionMiddleware(cfg))
	e.Use(LocalizationMiddleware(cfg))
	e.Use(ConfigMiddleware(cfg))
	e.Use(CSRFMiddleware())
	e.Use(ErrorLoggerMiddleware())
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
	URLFor = func(e string, paths ...string) string {
		for _, ep := range Endpoints {
			if strings.ToLower(ep.Name) == strings.ToLower(e) {
				return baseURL(resolveDynamicPath(ep.Path, paths))
			}
		}
		log.Error().Str("Endpoint", e).Msg("Not found")
		return "/"
	}
	tplFuncMap["BaseURL"] = baseURL
	tplFuncMap["URLFor"] = URLFor
	// ROUTES
	staticFS(e, "/static", static.FS, storage.FS())
	for _, ep := range Endpoints {
		if ep.AuthRequired {
			registerEndpoint(authorized, ep)
		} else {
			registerEndpoint(&e.RouterGroup, ep)
		}
	}
	e.NoRoute(notFoundView)
	e.HTMLRender = createRenderer(templates.FS)
	return e
}

func openStaticFS(name string, staticfs fs.FS, snapshotfs fs.FS) (fs.File, bool, error) {
	if strings.HasPrefix(name, "data/") {
		name := strings.TrimPrefix(name, "data/")
		f, err := snapshotfs.Open(name)
		return f, true, err
	}
	f, err := staticfs.Open(name)
	return f, false, err
}

// staticFS returns files without any directory indexing and can apply
// various content settings for snapshots. This exists because Gin's
// filesystem handling isn't a good match for our needs and webapp statics
// that come from "embed" don't play nicely with filesytem snapshot content
// that's only known at run-time.
func staticFS(e *gin.Engine, prefix string, staticfs fs.FS, snapshotfs fs.FS) {
	handler := func(c *gin.Context) {
		name := strings.TrimPrefix(c.Param("filepath"), "/")
		f, snapshotContent, err := openStaticFS(name, staticfs, snapshotfs)
		if err != nil {
			notFoundView(c)
			return
		}
		defer f.Close()
		info, err := f.Stat()
		if err != nil || info.IsDir() {
			notFoundView(c)
			return
		}
		seeker, ok := f.(io.ReadSeeker)
		if !ok {
			notFoundView(c)
			return
		}
		// Don't add gzip or content-type headers until after we've handled
		// all of the error cases so that 404 pages get rendered correctly.
		if snapshotContent {
			if strings.HasPrefix(name, "data/snapshots/") {
				c.Header("Content-Type", "text/html; charset=utf-8")
			}
			c.Header("Content-Encoding", "gzip")
		}
		http.ServeContent(c.Writer, c.Request, name, info.ModTime(), seeker)
	}
	pattern := path.Join(prefix, "/*filepath")
	e.GET(pattern, handler)
	e.HEAD(pattern, handler)
}

func Run(cfg *config.Config) {
	gin.SetMode(gin.ReleaseMode)

	engine := createEngine(cfg)
	log.Info().Str("Address", cfg.Server.Address).Msg("Starting server")
	err := engine.Run(cfg.Server.Address)
	if err != nil {
		log.Error().Err(err).Msg("Cannot start server")
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

func SessionMiddleware(cfg *config.Config) gin.HandlerFunc {
	if cfg.Server.RemoteUserHeader != "" {
		// Always trust the username sent in the RemoteUserHeader. The user set
		// in the session is ignored.
		header := cfg.Server.RemoteUserHeader
		return func(c *gin.Context) {
			// Set the user in the context
			hUname := c.GetHeader(header)
			u := model.GetUser(hUname)
			// Create a user if it doesn't already exist
			if hUname == "" {
				log.Error().Msgf("remote user header %q was empty or not present, unable to log user in", header)
			} else if u == nil {
				log.Debug().Msgf("Automatically creating user '%s' based on remote user header", hUname)
				err := validateUsername(hUname)
				if err == nil {
					err = model.CreateUser(hUname, "")
				}
				if err == nil {
					u = model.GetUser(hUname)
				} else {
					log.Error().Err(err).Msg("Failed to automatically create user")
				}
			}
			c.Set("user", u)

			// Update the session if the username wasn't present
			session := sessions.Default(c)
			sUname := session.Get(SID)
			if sUname == nil || sUname.(string) != hUname {
				session.Set(SID, hUname)
				_ = session.Save()
			}

			c.Next()
		}
	}
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

func LocalizationMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		if c.PostForm("lang") != "" {
			session.Set("lang", c.PostForm("lang"))
			_ = session.Save()
		}
		lang := ""
		if session.Get("lang") != nil {
			lang = session.Get("lang").(string)
		}
		if c.Query("lang") != "" {
			lang = c.Query("lang")
		}
		aLang := c.Request.Header.Get("Accept-Language")
		if lang != "" {
			c.Set("localizer", localization.NewLocalizer(lang, aLang))
		} else {
			c.Set("localizer", localization.NewLocalizer(aLang))
		}
		c.Next()
	}
}

func CSRFMiddleware() gin.HandlerFunc {
	protection := csrf.New()
	return func(c *gin.Context) {
		if strings.HasSuffix(c.HandlerName(), ".addBookmark") || strings.HasSuffix(c.HandlerName(), ".addResource") {
			c.Next()
			return
		}
		err := protection.Check(c.Request)
		if err != nil {
			c.String(403, err.Error())
			c.Abort()
			return
		}
	}
}

func ErrorLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		err, ok := c.Get("Error")
		if ok {
			log.Error().Str("Error", err.(string)).Msg("webapp error")
		}
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
