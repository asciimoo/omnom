package webapp

import (
	"net/http"
	"net/url"

	"github.com/asciimoo/omnom/model"

	"github.com/gin-gonic/gin"
)

func bookmarks(c *gin.Context) {
	var bs []*model.Bookmark
	pageno := getPageno(c)
	offset := (pageno - 1) * bookmarksPerPage
	var bookmarkCount int64
	model.DB.Model(&model.Bookmark{}).Where("bookmarks.public = 1").Count(&bookmarkCount)
	model.DB.Limit(int(bookmarksPerPage)).Offset(int(offset)).Where("bookmarks.public = 1").Preload("Snapshots").Order("created_at desc").Find(&bs)
	renderHTML(c, http.StatusOK, "bookmarks", map[string]interface{}{
		"Bookmarks":     bs,
		"Pageno":        pageno,
		"BookmarkCount": bookmarkCount,
		"HasNextPage":   offset+bookmarksPerPage < bookmarkCount,
	})
}

func myBookmarks(c *gin.Context) {
	u, _ := c.Get("user")
	var bs []*model.Bookmark
	pageno := getPageno(c)
	offset := (pageno - 1) * bookmarksPerPage
	var bookmarkCount int64
	model.DB.Model(&model.Bookmark{}).Where("bookmarks.user_id = ?", u.(*model.User).ID).Count(&bookmarkCount)
	model.DB.Limit(int(bookmarksPerPage)).Offset(int(offset)).Model(u).Preload("Snapshots").Order("created_at desc").Association("Bookmarks").Find(&bs)
	renderHTML(c, http.StatusOK, "my-bookmarks", map[string]interface{}{
		"Bookmarks":     bs,
		"Pageno":        pageno,
		"BookmarkCount": bookmarkCount,
		"HasNextPage":   offset+bookmarksPerPage < bookmarkCount,
	})
}

func addBookmark(c *gin.Context) {
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
	if c.PostForm("public") != "" {
		b.Public = true
	}
	snapshot := c.PostForm("snapshot")
	if snapshot != "" {
		b.Snapshots = []model.Snapshot{
			model.Snapshot{
				Site:  snapshot,
				Title: c.PostForm("snapshot_title"),
			},
		}
	}
	model.DB.Save(b)
	c.Redirect(http.StatusFound, "/")
}
