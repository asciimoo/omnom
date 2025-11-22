// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

package webapp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"time"

	"github.com/asciimoo/omnom/config"
	"github.com/asciimoo/omnom/model"
	"github.com/asciimoo/omnom/static"
	"github.com/asciimoo/omnom/storage"
	"github.com/asciimoo/omnom/validator"

	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

const (
	dateAsc  = "date_asc"
	dateDesc = "date_desc"
)

var snapshotJS string
var shadowDOMScript = []byte(`<script>
// render shadow DOM nodes
(()=>{
    document.currentScript.remove();
    function processNode(node){
        node.querySelectorAll("[omnomshadowroot]").forEach(el => {
            let shadowRoot = el.shadowRoot;
            if (!shadowRoot) {
                try {
                    shadowRoot = el.attachShadow({mode: "open"});
                    let tpl = el.querySelector("template");
                    shadowRoot.innerHTML = tpl.innerHTML;
                    tpl.remove()
                } catch (error) {
                    console.log("Cannot populate shadow root element:", error);
                }
                if (shadowRoot) {
                    processNode(shadowRoot);
                }
            }
        });
    }
    processNode(document);
})()</script>`)

var bookmarkSubmenu = []struct {
	Name     string
	Endpoint string
	PageName string
}{
	{"my", "My bookmarks", "my-bookmarks"},
	{"public", "Public bookmarks", "bookmarks"},
	{"create", "Create bookmark form", "create-bookmark"},
}

type browserSnapshotResponse struct {
	DOM       string `json:"dom"`
	Favicon   string `json:"favicon"`
	Title     string `json:"string"`
	Text      string `json:"string"`
	Resources []struct {
		Content   []byte `json:"content"`
		Mimetype  string `json:"mimetype"`
		Filename  string `json:"filename"`
		Extension string `json:"extension"`
		Src       string `json:"src"`
	} `json:"resources"`
}

func bookmarks(c *gin.Context) {
	var bs []*model.Bookmark
	pageno := getPageno(c)
	offset := (pageno - 1) * resultsPerPage
	hasSearch := false
	sp := &searchParams{}
	var bookmarkCount int64
	if err := c.ShouldBind(sp); err != nil {
		setNotification(c, nError, err.Error(), false)
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	cq := model.DB.Model(&model.Bookmark{}).Where("bookmarks.public = 1")
	//nolint: gosec // uint -> int conversion is safe
	q := model.DB.Limit(int(resultsPerPage)).Offset(int(offset)).Where("bookmarks.public = 1").Preload("Snapshots").Preload("Tags").Preload("User").Preload("Collection")
	if !reflect.DeepEqual(*sp, searchParams{}) {
		hasSearch = true
		filterText(sp.Q, sp.SearchInNote, sp.SearchInSnapshot, q, cq)
		filterOwner(sp.Owner, q, cq)
		_ = filterFromDate(sp.FromDate, q, cq)
		_ = filterToDate(sp.ToDate, q, cq)
		filterDomain(sp.Domain, q, cq)
		filterTag(sp.Tag, q, cq)
	}
	q.Group("bookmarks.id")
	cq.Group("bookmarks.id")
	cq.Count(&bookmarkCount)
	orderBy, _ := c.GetQuery("order_by")
	switch orderBy {
	case dateAsc:
		q = q.Order("bookmarks.updated_at asc")
	case dateDesc:
		q = q.Order("bookmarks.updated_at desc")
	default:
		q = q.Order("bookmarks.updated_at desc")
	}
	q.Find(&bs)
	args := map[string]any{
		"Bookmarks":     bs,
		"Pageno":        pageno,
		"BookmarkCount": bookmarkCount,
		//nolint: gosec // conversion is safe
		"HasNextPage":  offset+resultsPerPage < uint(bookmarkCount),
		"SearchParams": sp,
		"HasSearch":    hasSearch,
		"OrderBy":      orderBy,
		"FrequentTags": model.GetFrequentPublicTags(20),
	}
	_, ok := c.Get("user")
	if ok {
		args["Submenu"] = bookmarkSubmenu
	}
	render(c, http.StatusOK, "bookmarks", args)
}

func myBookmarks(c *gin.Context) {
	u, _ := c.Get("user")
	uid := u.(*model.User).ID
	var bs []*model.Bookmark
	pageno := getPageno(c)
	offset := (pageno - 1) * resultsPerPage
	var bookmarkCount int64
	cq := model.DB.Model(&model.Bookmark{}).Where("bookmarks.user_id = ?", uid)
	//nolint: gosec // uint -> int conversion is safe
	q := model.DB.Limit(int(resultsPerPage)).Offset(int(offset)).Model(&model.Bookmark{}).Where("bookmarks.user_id = ?", u.(*model.User).ID).Preload("Snapshots").Preload("Tags").Preload("User").Preload("Collection")
	sp := &searchParams{}
	hasSearch := false
	if err := c.ShouldBind(sp); err != nil {
		setNotification(c, nError, err.Error(), false)
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	} else {
		if !reflect.DeepEqual(*sp, searchParams{}) {
			hasSearch = true
			filterText(sp.Q, sp.SearchInNote, sp.SearchInSnapshot, q, cq)
			_ = filterFromDate(sp.FromDate, q, cq)
			_ = filterToDate(sp.ToDate, q, cq)
			filterDomain(sp.Domain, q, cq)
			filterTag(sp.Tag, q, cq)
			filterCollection(sp.Collection, uid, q, cq)
			if sp.IsPublic {
				filterPublic(q, cq)
			}
			if sp.IsPrivate {
				filterPublic(q, cq)
			}
		}
	}
	cq.Count(&bookmarkCount)
	orderBy := c.Query("order_by")
	switch orderBy {
	case dateAsc:
		q = q.Order("bookmarks.updated_at asc")
	case dateDesc:
		q = q.Order("bookmarks.updated_at desc")
	default:
		q = q.Order("bookmarks.updated_at desc")
	}
	q.Find(&bs)

	cols := model.GetCollectionTree(uid)
	render(c, http.StatusOK, "my-bookmarks", map[string]any{
		"Bookmarks":     bs,
		"Pageno":        pageno,
		"Submenu":       bookmarkSubmenu,
		"BookmarkCount": bookmarkCount,
		//nolint: gosec // conversion is safe
		"HasNextPage":       offset+resultsPerPage < uint(bookmarkCount),
		"SearchParams":      sp,
		"HasSearch":         hasSearch,
		"OrderBy":           orderBy,
		"Collections":       cols,
		"CurrentCollection": c.Query("collection"),
	})
}

func createBookmarkForm(c *gin.Context) {
	cfg, _ := c.Get("config")
	u, _ := c.Get("user")
	uid := u.(*model.User).ID
	cols := model.GetCollectionTree(uid)
	render(c, http.StatusOK, "create-bookmark", map[string]any{
		"Collections":           cols,
		"Submenu":               bookmarkSubmenu,
		"AllowSnapshotCreation": cfg.(*config.Config).App.CreateSnapshotFromWebapp,
	})
}

func createBookmark(c *gin.Context) {
	cfg, _ := c.Get("config")
	cu, _ := c.Get("user")
	u, _ := cu.(*model.User)

	bs := &browserSnapshotResponse{}
	bsCreated := false
	var err error
	if cfg.(*config.Config).App.CreateSnapshotFromWebapp {
		bs, err = createSnapshot(c.PostForm("url"), cfg.(*config.Config).App.WebappSnapshotterTimeout)
		if err != nil {
			setNotification(c, nError, "Failed to create snapshot: "+err.Error(), true)
		} else {
			bsCreated = true
		}
	}
	b, new, err := model.GetOrCreateBookmark(
		u,
		c.PostForm("url"),
		c.PostForm("title"),
		c.PostForm("tags"),
		c.PostForm("notes"),
		c.PostForm("public"),
		bs.Favicon,
		c.PostForm("collection"),
		c.PostForm("unread"),
	)
	if err != nil {
		setNotification(c, nError, "Failed to create bookmark: "+err.Error(), true)
		c.Redirect(http.StatusFound, URLFor("Create bookmark form"))
		return
	}
	if new {
		go apNotifyFollowers(c, b)
	}
	if bsCreated {
		key, err := storeSnapshot([]byte(bs.DOM))
		if err != nil {
			setNotification(c, nError, "Failed to create snapshot: "+err.Error(), true)
			c.Redirect(http.StatusFound, URLFor("Create bookmark form"))
			return
		}

		s := &model.Snapshot{
			Key:        key,
			Text:       bs.Text,
			Title:      bs.Title,
			BookmarkID: b.ID,
			Size:       storage.GetSnapshotSize(key),
		}
		for _, r := range bs.Resources {
			if bytes.Equal(r.Content, []byte("")) {
				continue
			}
			key := storage.Hash(r.Content) + "." + r.Extension
			err = storage.SaveResource(key, r.Content)
			if err != nil {
				setNotification(c, nError, "Failed to create bookmark: "+err.Error(), true)
				c.Redirect(http.StatusFound, URLFor("Create bookmark form"))
				return
			}
			size := storage.GetResourceSize(key)
			s.Size += size
			// TODO check error in GetOrCreateResource
			s.Resources = append(s.Resources, model.GetOrCreateResource(key, r.Mimetype, r.Filename, size))
		}
		if err := model.DB.Save(s).Error; err != nil {
			setNotification(c, nError, "Failed to create bookmark: "+err.Error(), true)
			c.Redirect(http.StatusFound, URLFor("Create bookmark form"))
			return
		}
	}

	setNotification(c, nInfo, "Bookmark successfully created", true)
	c.Redirect(http.StatusFound, fmt.Sprintf("%s?id=%d", URLFor("Bookmark"), b.ID))
}

func createSnapshot(urlString string, to int) (*browserSnapshotResponse, error) {
	if snapshotJS == "" {
		b, err := static.FS.ReadFile("js/snapshot.js")
		if err != nil {
			log.Error().Err(err).Msg("Failed to read snapshot.js")
		}
		snapshotJS = string(b)
	}

	ctx, cancel := chromedp.NewContext(
		context.Background(),
		//chromedp.WithLogf(log.Printf),
		//chromedp.WithDebugf(log.Printf),
		//chromedp.WithErrorf(log.Printf),
	)
	defer cancel()
	if to > 0 {
		ctx, cancel = context.WithTimeout(ctx, time.Duration(to)*time.Second)
	}
	defer cancel()
	res := &browserSnapshotResponse{}
	err := chromedp.Run(ctx,
		chromedp.EmulateViewport(1200, 1000),
		chromedp.Navigate(urlString),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Evaluate(snapshotJS, nil, func(p *runtime.EvaluateParams) *runtime.EvaluateParams {
			return p.WithAwaitPromise(true)
		}),
		chromedp.Evaluate(`webapp_snapshot.createOmnomSnapshot();`, res, func(p *runtime.EvaluateParams) *runtime.EvaluateParams {
			return p.WithAwaitPromise(true)
		}),
	)
	return res, err
}

func addBookmark(c *gin.Context) {
	// TODO error handling
	tok := c.PostForm("token")
	u := model.GetUserBySubmissionToken(tok)
	if u == nil {
		setNotification(c, nError, "Invalid token", false)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid token",
		})
		return
	}
	b, new, err := model.GetOrCreateBookmark(
		u,
		c.PostForm("url"),
		c.PostForm("title"),
		c.PostForm("tags"),
		c.PostForm("notes"),
		c.PostForm("public"),
		c.PostForm("favicon"),
		c.PostForm("collection"),
		c.PostForm("unread"),
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create bookmark DB entry")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	if new {
		go apNotifyFollowers(c, b)
	}
	snapshotFile, _, err := c.Request.FormFile("snapshot")
	if err != nil {
		log.Debug().Msg("No snpashot found")
		c.JSON(http.StatusOK, map[string]any{
			"success":      true,
			"bookmark_url": baseURL(fmt.Sprintf("/bookmark?id=%d", b.ID)),
		})
		return
	}
	snapshot, err := io.ReadAll(snapshotFile)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read snapshot")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	var sSize uint
	var sKey = ""
	if !bytes.Equal(snapshot, []byte("")) {
		key, err := storeSnapshot(snapshot)
		if err != nil {
			log.Error().Err(err).Msg("Failed to validate snapshot HTML")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "HTML validation failed: " + err.Error(),
			})
			return
		}
		s := &model.Snapshot{
			Key:        key,
			Text:       c.PostForm("snapshot_text"),
			Title:      c.PostForm("snapshot_title"),
			BookmarkID: b.ID,
			Size:       storage.GetSnapshotSize(key),
		}
		if err := model.DB.Save(s).Error; err != nil {
			log.Error().Err(err).Msg("Failed to create snapshot DB entry")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		sSize = s.Size
		sKey = key
	}
	c.JSON(http.StatusOK, map[string]any{
		"success":       true,
		"bookmark_url":  baseURL(fmt.Sprintf("/bookmark?id=%d", b.ID)),
		"snapshot_url":  baseURL(fmt.Sprintf("/static/data/snapshots/%s/%s.gz", sKey[:2], sKey)),
		"snapshot_size": formatSize(sSize),
		"snapshot_key":  sKey,
	})
}

func checkBookmark(c *gin.Context) {
	resp := make(map[string]any)
	tok, ok := c.GetQuery("token")
	if !ok {
		resp["error"] = "missing token"
		c.JSON(401, resp)
		return
	}
	var URL string
	URL, ok = c.GetQuery("url")
	if !ok {
		resp["error"] = "missing URL"
		c.JSON(400, resp)
		return
	}
	var bc int64
	model.DB.
		Model(&model.Bookmark{}).
		Joins("join users on bookmarks.user_id = users.id").
		Joins("join tokens on tokens.user_id = users.id").
		Where("tokens.text = ? and bookmarks.url = ? and tokens.deleted_at IS NULL", tok, URL).
		Limit(1).
		Count(&bc)

	if bc == 0 {
		resp["found"] = false
	} else {
		resp["found"] = true
	}
	c.JSON(200, resp)
}

func viewBookmark(c *gin.Context) {
	u, _ := c.Get("user")
	bid, ok := c.GetQuery("id")
	if !ok {
		return
	}
	var b *model.Bookmark
	model.DB.Model(b).Where("id = ?", bid).Preload("Snapshots").Preload("Tags").First(&b)
	if b == nil {
		return
	}
	if !b.Public && (u == nil || u.(*model.User).ID != b.UserID) {
		return
	}
	render(c, http.StatusOK, "view-bookmark", map[string]any{
		"Bookmark": b,
	})
}

func editBookmark(c *gin.Context) {
	u, _ := c.Get("user")
	uid := u.(*model.User).ID
	bid, ok := c.GetQuery("id")
	if !ok {
		return
	}
	var b *model.Bookmark
	model.DB.Model(b).Where("id = ?", bid).Preload("Snapshots").Preload("Tags").Preload("Collection").First(&b)
	if b == nil {
		return
	}
	if uid != b.UserID {
		return
	}
	cols := model.GetCollectionTree(uid)
	col := ""
	if b.Collection != nil {
		col = b.Collection.Name
	}
	render(c, http.StatusOK, "edit-bookmark", map[string]any{
		"Bookmark":          b,
		"Collections":       cols,
		"CurrentCollection": col,
	})
}

func saveBookmark(c *gin.Context) {
	u, _ := c.Get("user")
	uid := u.(*model.User).ID
	bid := c.PostForm("id")
	if bid == "" {
		setNotification(c, nError, "Missing bookmark ID", true)
		c.Redirect(http.StatusFound, baseURL("/"))
		return
	}
	var b *model.Bookmark
	model.DB.Model(b).Where("id = ?", bid).First(&b)
	if b == nil {
		setNotification(c, nError, "Missing bookmark", true)
		c.Redirect(http.StatusFound, baseURL("/"))
		return
	}
	if uid != b.UserID {
		setNotification(c, nError, "Permission denied", true)
		c.Redirect(http.StatusFound, baseURL("/"))
		return
	}
	t := c.PostForm("title")
	if t != "" {
		b.Title = t
	}
	col := model.GetCollectionByName(uid, c.PostForm("collection"))
	if col != nil {
		b.CollectionID = col.ID
	}
	b.Public = c.PostForm("public") != ""
	b.Unread = c.PostForm("unread") != ""
	b.Notes = c.PostForm("notes")
	err := model.DB.Save(b).Error
	if err != nil {
		setNotification(c, nError, "Failed to save bookmark: "+err.Error(), true)
	} else {
		setNotification(c, nInfo, "Bookmark saved", true)
	}
	c.Redirect(http.StatusFound, baseURL("/edit_bookmark?id="+bid))
}

func deleteBookmark(c *gin.Context) {
	id := c.PostForm("id")
	if id == "" {
		return
	}
	u, _ := c.Get("user")
	var b *model.Bookmark
	err := model.DB.
		Model(&model.Bookmark{}).
		Where("bookmarks.id = ? and bookmarks.user_id", id, u.(*model.User).ID).First(&b).Error
	if err != nil {
		setNotification(c, nError, "Failed to delete bookmark: "+err.Error(), true)
	} else {
		setNotification(c, nInfo, "Bookmark deleted", true)
	}
	if b != nil {
		model.DB.Delete(&model.Snapshot{}, "bookmark_id = ?", id)
		model.DB.Delete(&model.Bookmark{}, "id = ?", id)
		model.DB.Delete("bookmark_tags", "bookmark_id = ?", id)
	}
	c.Redirect(http.StatusFound, baseURL("/"))
}

func addTag(c *gin.Context) {
	tag := c.PostForm("tag")
	bid := c.PostForm("bid")
	if tag == "" || bid == "" {
		return
	}
	var b *model.Bookmark
	err := model.DB.Where("id = ?", bid).Preload("Tags").First(&b).Error
	if err != nil {
		setNotification(c, nError, "Unknown bookmark", true)
		c.Redirect(http.StatusFound, baseURL("/"))
		return
	}
	u, _ := c.Get("user")
	if u.(*model.User).ID != b.UserID {
		setNotification(c, nError, "Permission denied", true)
		c.Redirect(http.StatusFound, baseURL("/"))
		return
	}
	b.Tags = append(b.Tags, model.GetOrCreateTag(tag))
	err = model.DB.Save(b).Error
	if err != nil {
		setNotification(c, nError, err.Error(), true)
		c.Redirect(http.StatusFound, baseURL("/edit_bookmark?id="+bid))
		return
	}
	setNotification(c, nInfo, "Tag added", true)
	c.Redirect(http.StatusFound, baseURL("/edit_bookmark?id="+bid))
}

func deleteTag(c *gin.Context) {
	tid := c.PostForm("tid")
	bid := c.PostForm("bid")
	if tid == "" || bid == "" {
		return
	}
	var b *model.Bookmark
	err := model.DB.Where("id = ?", bid).First(&b).Error
	if err != nil {
		setNotification(c, nError, "Unknown bookmark", true)
		c.Redirect(http.StatusFound, baseURL("/"))
		return
	}
	u, _ := c.Get("user")
	if u.(*model.User).ID != b.UserID {
		setNotification(c, nError, "Permission denied", true)
		c.Redirect(http.StatusFound, baseURL("/"))
		return
	}
	var t *model.Tag
	model.DB.Where("id = ?", tid).First(&t)
	err = model.DB.Model(b).Association("Tags").Delete(t)
	if err != nil {
		setNotification(c, nError, err.Error(), true)
		c.Redirect(http.StatusFound, baseURL("/edit_bookmark?id="+bid))
		return
	}
	setNotification(c, nInfo, "Tag deleted", true)
	c.Redirect(http.StatusFound, baseURL("/edit_bookmark?id="+bid))
}

func storeSnapshot(sb []byte) (string, error) {
	vr := validator.ValidateHTML(sb)
	if vr.Error != nil {
		return "", vr.Error
	}

	if vr.HasShadowDOM {
		sb = append(sb, shadowDOMScript...)
	}

	key := storage.Hash(sb)
	err := storage.SaveSnapshot(key, sb)
	if err != nil {
		return "", err
	}
	return key, nil
}

func pageInfo(c *gin.Context) {
	u := model.GetUserBySubmissionToken(c.PostForm("token"))
	if u == nil {
		c.IndentedJSON(http.StatusOK, nil)
		return
	}
	tags, err := model.GetUserTagsFromText(c.PostForm("text"), u.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query user tags")
	}
	ret := struct {
		Collections []*model.Collection `json:"collections"`
		Tags        []*model.Tag        `json:"tags"`
	}{
		Collections: model.GetCollections(u.ID),
		Tags:        tags,
	}
	c.IndentedJSON(http.StatusOK, ret)
}
