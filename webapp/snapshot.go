package webapp

import (
	"net/http"

	"github.com/asciimoo/omnom/model"
	"github.com/asciimoo/omnom/storage"

	"github.com/gin-gonic/gin"

	"github.com/gin-gonic/contrib/sessions"
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
	err := model.DB.Where("key = ? and bookmark_id = ?", sid, bid).First(&s).Error
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
	err = model.DB.Where("key = ? and bookmark_id = ?", sid, bid).First(&s).Error
	if err != nil {
		return
	}
	if s.Size == 0 {
		s.Size = storage.GetSnapshotSize(s.Key)
		err = model.DB.Save(s).Error
		if err != nil {
			return
		}
	}
	var otherSnapshots []struct {
		Title string
		Bid   int64
		Sid   string
	}
	err = model.DB.
		Model(&model.Snapshot{}).
		Select("bookmarks.id as bid, snapshots.key as sid, snapshots.title as title").
		Joins("join bookmarks on bookmarks.id = snapshots.bookmark_id").
		Where("bookmarks.url = ? and snapshots.key != ?", b.URL, s.Key).Find(&otherSnapshots).Error
	if err != nil {
		return
	}
	renderHTML(c, http.StatusOK, "snapshotWrapper", map[string]interface{}{
		"Bookmark":       b,
		"Snapshot":       s,
		"hideFooter":     true,
		"OtherSnapshots": otherSnapshots,
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
	c.Header("Content-Encoding", "gzip")
	c.Data(http.StatusOK, "text/html; charset=utf-8", sBody)
}

func deleteSnapshot(c *gin.Context) {
	u, _ := c.Get("user")
	session := sessions.Default(c)
	defer session.Save()
	bid := c.PostForm("bid")
	sid := c.PostForm("sid")
	if bid == "" || sid == "" {
		return
	}
	var s *model.Snapshot
	err := model.DB.
		Model(&model.Snapshot{}).
		Joins("join bookmarks on bookmarks.id = snapshots.bookmark_id").
		Where("snapshots.id = ? and snapshots.bookmark_id = ? and bookmarks.user_id", sid, bid, u.(*model.User).ID).First(&s).Error
	if err != nil {
		session.Set("Error", "Failed to delete snapshot: "+err.Error())
	} else {
		session.Set("Info", "Snapshot deleted")
	}
	if s != nil {
		model.DB.Delete(&model.Snapshot{}, "id = ? and bookmark_id = ?", sid, bid)
	}
	c.Redirect(http.StatusFound, baseURL("/edit_bookmark?id="+bid))
}
