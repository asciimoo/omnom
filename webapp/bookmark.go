package webapp

import (
	"net/http"
	"strconv"

	"github.com/asciimoo/omnom/model"

	"github.com/gin-gonic/gin"
)

func bookmarks(c *gin.Context) {
	var bs []*model.Bookmark
	var pageno int64 = 1
	if pagenoStr, ok := c.GetQuery("pageno"); ok {
		if userPageno, err := strconv.Atoi(pagenoStr); err == nil && userPageno > 0 {
			pageno = int64(userPageno)
		}
	}
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
	var pageno int64 = 1
	if pagenoStr, ok := c.GetQuery("pageno"); ok {
		if userPageno, err := strconv.Atoi(pagenoStr); err == nil && userPageno > 0 {
			pageno = int64(userPageno)
		}
	}
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
	b := &model.Bookmark{
		Title:   c.PostForm("title"),
		URL:     c.PostForm("url"),
		Notes:   c.PostForm("notes"),
		Favicon: c.PostForm("favicon"),
		UserID:  u.ID,
	}
	if c.PostForm("public") != "" {
		b.Public = true
	}
	snapshot := c.PostForm("snapshot")
	if snapshot != "" {
		b.Snapshots = []model.Snapshot{
			model.Snapshot{
				Site: snapshot,
			},
		}
	}
	model.DB.Save(b)
	c.Redirect(http.StatusFound, "/")
}
