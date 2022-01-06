package webapp

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/asciimoo/omnom/model"
	"github.com/asciimoo/omnom/storage"

	"github.com/gin-gonic/gin"

	"github.com/gin-gonic/contrib/sessions"
)

func bookmarks(c *gin.Context) {
	var bs []*model.Bookmark
	pageno := getPageno(c)
	offset := (pageno - 1) * bookmarksPerPage
	var bookmarkCount int64
	cq := model.DB.Model(&model.Bookmark{}).Where("bookmarks.public = 1")
	q := model.DB.Limit(int(bookmarksPerPage)).Offset(int(offset)).Where("bookmarks.public = 1").Preload("Snapshots").Preload("Tags")
	sp := &searchParams{}
	hasSearch := false
	if err := c.ShouldBind(sp); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	} else {
		if !reflect.DeepEqual(*sp, searchParams{}) {
			hasSearch = true
			filterText(sp.Q, sp.SearchInNote, sp.SearchInSnapshot, q, cq)
			filterOwner(sp.Owner, q, cq)
			filterFromDate(sp.FromDate, q, cq)
			filterToDate(sp.ToDate, q, cq)
			filterDomain(sp.Domain, q, cq)
			filterTag(sp.Tag, q, cq)
		}
	}
	cq.Count(&bookmarkCount)
	q.Order("bookmarks.updated_at desc").Find(&bs)
	renderHTML(c, http.StatusOK, "bookmarks", map[string]interface{}{
		"Bookmarks":     bs,
		"Pageno":        pageno,
		"BookmarkCount": bookmarkCount,
		"HasNextPage":   offset+bookmarksPerPage < bookmarkCount,
		"SearchParams":  sp,
		"HasSearch":     hasSearch,
	})
}

func myBookmarks(c *gin.Context) {
	u, _ := c.Get("user")
	var bs []*model.Bookmark
	pageno := getPageno(c)
	offset := (pageno - 1) * bookmarksPerPage
	var bookmarkCount int64
	cq := model.DB.Model(&model.Bookmark{}).Where("bookmarks.user_id = ?", u.(*model.User).ID)
	q := model.DB.Limit(int(bookmarksPerPage)).Offset(int(offset)).Model(&model.Bookmark{}).Where("bookmarks.user_id = ?", u.(*model.User).ID).Preload("Snapshots").Preload("Tags")
	sp := &searchParams{}
	hasSearch := false
	if err := c.ShouldBind(sp); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	} else {
		if !reflect.DeepEqual(*sp, searchParams{}) {
			hasSearch = true
			filterText(sp.Q, sp.SearchInNote, sp.SearchInSnapshot, q, cq)
			filterFromDate(sp.FromDate, q, cq)
			filterToDate(sp.ToDate, q, cq)
			filterDomain(sp.Domain, q, cq)
			filterTag(sp.Tag, q, cq)
			if sp.IsPublic {
				filterPublic(q, cq)
			}
			if sp.IsPrivate {
				filterPublic(q, cq)
			}
		}
	}
	cq.Count(&bookmarkCount)
	q.Order("bookmarks.updated_at desc").Find(&bs)
	renderHTML(c, http.StatusOK, "my-bookmarks", map[string]interface{}{
		"Bookmarks":     bs,
		"Pageno":        pageno,
		"BookmarkCount": bookmarkCount,
		"HasNextPage":   offset+bookmarksPerPage < bookmarkCount,
		"SearchParams":  sp,
		"HasSearch":     hasSearch,
	})
}

func addBookmark(c *gin.Context) {
	// TODO error handling
	tok := c.PostForm("token")
	u := model.GetUserBySubmissionToken(tok)
	if u == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "invalid token",
		})
		return
	}
	url, err := url.Parse(c.PostForm("url"))
	if err != nil || url.Hostname() == "" || url.Scheme == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "invalid url",
		})
		return
	}
	var b *model.Bookmark = nil
	newBookmark := false
	r := model.DB.
		Model(&model.Bookmark{}).
		Preload("Snapshots").
		Where("url = ? and user_id = ?", url.String(), u.ID).
		First(&b)
	if r.RowsAffected < 1 {
		newBookmark = true
	}
	if newBookmark {
		b = &model.Bookmark{
			Title:     c.PostForm("title"),
			URL:       url.String(),
			Domain:    url.Hostname(),
			Notes:     c.PostForm("notes"),
			Favicon:   c.PostForm("favicon"),
			UserID:    u.ID,
			Snapshots: make([]model.Snapshot, 0, 8),
		}
		if b.Title == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "missing title",
			})
			return
		}
		if !strings.HasPrefix(b.Favicon, "data:image") {
			b.Favicon = ""
		}
		if c.PostForm("public") != "" {
			b.Public = true
		}
		tags := c.PostForm("tags")
		if tags != "" {
			b.Tags = make([]model.Tag, 0, 8)
			for _, t := range strings.Split(tags, ",") {
				t = strings.TrimSpace(t)
				b.Tags = append(b.Tags, model.Tag{
					Text: t,
				})
			}
		}
		if err := model.DB.Save(b).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
	}
	snapshotFile, _, err := c.Request.FormFile("snapshot")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	snapshot, err := ioutil.ReadAll(snapshotFile)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	if !bytes.Equal(snapshot, []byte("")) {
		key := storage.Hash(snapshot)
		_ = storage.SaveSnapshot(key, snapshot)
		s := &model.Snapshot{
			Key:        key,
			Text:       c.PostForm("snapshot_text"),
			Title:      c.PostForm("snapshot_title"),
			BookmarkID: b.ID,
		}
		if err := model.DB.Save(s).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
	}
	c.JSON(200, map[string]bool{"success": true})
}

func checkBookmark(c *gin.Context) {
	resp := make(map[string]interface{})
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
	renderHTML(c, http.StatusOK, "view-bookmark", map[string]interface{}{
		"Bookmark": b
	})
}

func editBookmark(c *gin.Context) {
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
	if u.(*model.User).ID != b.UserID {
		return
	}
	renderHTML(c, http.StatusOK, "edit-bookmark", map[string]interface{}{
		"Bookmark": b,
	})
}

func saveBookmark(c *gin.Context) {
	u, _ := c.Get("user")
	session := sessions.Default(c)
	defer session.Save()
	bid := c.PostForm("id")
	if bid == "" {
		session.Set("Error", "Missing bookmark ID")
		c.Redirect(http.StatusFound, baseURL("/"))
		return
	}
	var b *model.Bookmark
	model.DB.Model(b).Where("id = ?", bid).First(&b)
	if b == nil {
		session.Set("Error", "Missing bookmark")
		c.Redirect(http.StatusFound, baseURL("/"))
		return
	}
	if u.(*model.User).ID != b.UserID {
		session.Set("Error", "Insufficient permission")
		c.Redirect(http.StatusFound, baseURL("/"))
		return
	}
	t := c.PostForm("title")
	if t != "" {
		b.Title = t
	}
	b.Public = c.PostForm("public") != ""
	b.Notes = c.PostForm("notes")
	err := model.DB.Save(b).Error
	if err != nil {
		session.Set("Error", "Failed to save bookmark: "+err.Error())
	} else {
		session.Set("Info", "Bookmark saved")
	}
	c.Redirect(http.StatusFound, baseURL("/edit_bookmark?id="+bid))
}
