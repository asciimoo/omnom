package webapp

import (
	"net/http"

	"github.com/asciimoo/omnom/model"
	"github.com/asciimoo/omnom/storage"

	"github.com/gin-gonic/gin"
)

func snapshotWrapper(c *gin.Context) {
	sid, ok := c.GetQuery("sid")
	if !ok {
		return
	}
	bid, ok := c.GetQuery("bid")
	if !ok {
		return
	}
	var s *model.Snapshot
	err := model.DB.Where("key = ?", sid).First(&s).Error
	if err != nil {
		return
	}
	var b *model.Bookmark
	err = model.DB.Where("id = ?", bid).First(&b).Error
	if err != nil {
		return
	}
	if s.BookmarkID != b.ID {
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
	sBody, err := storage.GetSnapshot(id)
	if err != nil {
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", sBody)
}
