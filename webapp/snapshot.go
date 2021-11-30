package webapp

import (
	"net/http"

	"github.com/asciimoo/omnom/model"

	"github.com/gin-gonic/gin"
)

func snapshotWrapper(c *gin.Context) {
	id, ok := c.GetQuery("id")
	if !ok {
		return
	}
	var s *model.Snapshot
	err := model.DB.Where("id = ?", id).First(&s).Error
	if err != nil {
		return
	}
	var b *model.Bookmark
	err = model.DB.Where("id = ?", s.BookmarkID).First(&b).Error
	if err != nil {
		return
	}
	renderHTML(c, http.StatusOK, "snapshotWrapper", map[string]interface{}{
		"Bookmark":   b,
		"Snapshot":   s,
		"hideFooter": true,
	})
}

func snapshot(c *gin.Context) {
	id, ok := c.GetQuery("id")
	if !ok {
		return
	}
	var s *model.Snapshot
	err := model.DB.Where("id = ?", id).First(&s).Error
	if err != nil {
		return
	}
	var b *model.Bookmark
	err = model.DB.Where("id = ?", s.BookmarkID).First(&b).Error
	if err != nil {
		return
	}
	u, _ := c.Get("user")
	if !b.Public && (u == nil || u.(*model.User).ID != b.UserID) {
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(s.Site))
}
