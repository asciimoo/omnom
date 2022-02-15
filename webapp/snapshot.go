package webapp

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"

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
		setNotification(c, nError, err.Error(), false)
		return
	}
	if s.BookmarkID != b.ID {
		setNotification(c, nError, "Invalid bookmark ID", false)
		return
	}
	err = model.DB.Where("key = ? and bookmark_id = ?", sid, bid).First(&s).Error
	if err != nil {
		setNotification(c, nError, err.Error(), false)
		return
	}
	if s.Size == 0 {
		s.Size = storage.GetSnapshotSize(s.Key)
		err = model.DB.Save(s).Error
		if err != nil {
			setNotification(c, nError, err.Error(), false)
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
		setNotification(c, nError, err.Error(), false)
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
		setNotification(c, nError, "Failed to delete snapshot: "+err.Error(), true)
	} else {
		setNotification(c, nInfo, "Snapshot deleted", true)
	}
	if s != nil {
		err = model.DB.Delete(&model.Snapshot{}, "id = ? and bookmark_id = ?", sid, bid).Error
		if err != nil {
			setNotification(c, nError, "Failed to delete snapshot: "+err.Error(), true)
		}
	}
	c.Redirect(http.StatusFound, baseURL("/edit_bookmark?id="+bid))
}

func addResource(c *gin.Context) {
	tok := c.PostForm("token")
	u := model.GetUserBySubmissionToken(tok)
	if u == nil {
		setNotification(c, nError, "Invalid token", false)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid token",
		})
		return
	}
	sid := c.PostForm("sid")
	var s *model.Snapshot
	model.DB.
		Model(&model.Snapshot{}).
		Joins("join bookmarks on bookmarks.id = snapshots.bookmark_id").
		Where("snapshots.key = ? and bookmarks.user_id = ?", sid, u.ID).
		Order("snapshots.updated_at desc").
		First(&s)
	if s == nil {
		setNotification(c, nError, "Snapshot not found", false)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Snapshot not found",
		})
		return
	}
	fname := c.PostForm("filename")
	resourceFile, _, err := c.Request.FormFile("resource")
	if err != nil {
		setNotification(c, nError, err.Error(), false)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	resource, err := ioutil.ReadAll(resourceFile)
	if err != nil {
		setNotification(c, nError, err.Error(), false)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	if !bytes.Equal(resource, []byte("")) {
		fparts := strings.Split(fname, ".")
		key := storage.Hash(resource) + "." + fparts[len(fparts)-1]
		err = storage.SaveResource(key, resource)
		if err != nil {
			setNotification(c, nError, err.Error(), false)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		mt := c.PostForm("mimetype")
		size := storage.GetResourceSize(key)
		s.Size += size
		s.Resources = append(s.Resources, model.GetOrCreateResource(key, fname, mt, size))
		model.DB.Save(s)
	}
	c.JSON(200, map[string]interface{}{
		"success": true,
	})
}
