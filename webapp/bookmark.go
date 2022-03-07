package webapp

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/asciimoo/omnom/model"
	"github.com/asciimoo/omnom/storage"
	"github.com/asciimoo/omnom/validator"

	"github.com/gin-gonic/gin"
)

const (
	dateAsc  = "date_asc"
	dateDesc = "date_desc"
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
		setNotification(c, nError, err.Error(), false)
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	} else {
		if !reflect.DeepEqual(*sp, searchParams{}) {
			hasSearch = true
			filterText(sp.Q, sp.SearchInNote, sp.SearchInSnapshot, q, cq)
			filterOwner(sp.Owner, q, cq)
			_ = filterFromDate(sp.FromDate, q, cq)
			_ = filterToDate(sp.ToDate, q, cq)
			filterDomain(sp.Domain, q, cq)
			filterTag(sp.Tag, q, cq)
		}
	}
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
	renderHTML(c, http.StatusOK, "bookmarks", map[string]interface{}{
		"Bookmarks":     bs,
		"Pageno":        pageno,
		"BookmarkCount": bookmarkCount,
		"HasNextPage":   offset+bookmarksPerPage < bookmarkCount,
		"SearchParams":  sp,
		"HasSearch":     hasSearch,
		"OrderBy":       orderBy,
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
			if sp.IsPublic {
				filterPublic(q, cq)
			}
			if sp.IsPrivate {
				filterPublic(q, cq)
			}
		}
	}
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
	renderHTML(c, http.StatusOK, "my-bookmarks", map[string]interface{}{
		"Bookmarks":     bs,
		"Pageno":        pageno,
		"BookmarkCount": bookmarkCount,
		"HasNextPage":   offset+bookmarksPerPage < bookmarkCount,
		"SearchParams":  sp,
		"HasSearch":     hasSearch,
		"OrderBy":       orderBy,
	})
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
	url, err := url.Parse(c.PostForm("url"))
	if err != nil || url.Hostname() == "" || url.Scheme == "" {
		setNotification(c, nError, "Invalid URL", false)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Invalid URL",
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
			setNotification(c, nError, "Missing title", false)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "Missing title",
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
				b.Tags = append(b.Tags, model.GetOrCreateTag(t))
			}
		}
		if err := model.DB.Save(b).Error; err != nil {
			setNotification(c, nError, err.Error(), false)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
	}
	snapshotFile, _, err := c.Request.FormFile("snapshot")
	if err != nil {
		setNotification(c, nError, err.Error(), false)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	snapshot, err := ioutil.ReadAll(snapshotFile)
	if err != nil {
		setNotification(c, nError, err.Error(), false)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	var sSize uint
	var sKey = ""
	if !bytes.Equal(snapshot, []byte("")) {
		// TODO don't save identical snapshots twice
		if err := validator.ValidateHTML(snapshot); err != nil {
			setNotification(c, nError, "HTML validation failed: "+err.Error(), false)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "HTML validation failed: " + err.Error(),
			})
			return
		}
		key := storage.Hash(snapshot)
		_ = storage.SaveSnapshot(key, snapshot)
		s := &model.Snapshot{
			Key:        key,
			Text:       c.PostForm("snapshot_text"),
			Title:      c.PostForm("snapshot_title"),
			BookmarkID: b.ID,
			Size:       storage.GetSnapshotSize(key),
		}
		if err := model.DB.Save(s).Error; err != nil {
			setNotification(c, nError, err.Error(), false)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		sSize = s.Size
		sKey = key
	}
	c.JSON(200, map[string]interface{}{
		"success":       true,
		"bookmark_url":  baseURL("/bookmark?id=" + strconv.Itoa(int(b.ID))),
		"snapshot_url":  baseURL(fmt.Sprintf("/static/data/snapshots/%s/%s.gz", sKey[:2], sKey)),
		"snapshot_size": formatSize(sSize),
		"snapshot_key":  sKey,
	})
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
		"Bookmark": b,
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
	if u.(*model.User).ID != b.UserID {
		setNotification(c, nError, "Permission denied", true)
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
