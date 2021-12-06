package webapp

import (
	"bytes"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/asciimoo/omnom/model"
	"github.com/asciimoo/omnom/storage"

	"github.com/gin-gonic/gin"
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
	q.Order("bookmarks.created_at desc").Find(&bs)
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
	q.Order("bookmarks.created_at desc").Find(&bs)
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
	b := &model.Bookmark{
		Title:   c.PostForm("title"),
		URL:     url.String(),
		Domain:  url.Hostname(),
		Notes:   c.PostForm("notes"),
		Favicon: c.PostForm("favicon"),
		UserID:  u.ID,
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
	snapshot := []byte(c.PostForm("snapshot"))
	if !bytes.Equal(snapshot, []byte("")) {
		key := storage.Hash(snapshot)
		_ = storage.SaveSnapshot(key, snapshot)
		b.Snapshots = []model.Snapshot{
			model.Snapshot{
				Key:   key,
				Text:  c.PostForm("snapshot_text"),
				Title: c.PostForm("snapshot_title"),
			},
		}
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
	model.DB.Save(b)
	c.Redirect(http.StatusFound, "/")
}
