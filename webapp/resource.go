package webapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/asciimoo/omnom/model"
	"github.com/asciimoo/omnom/storage"

	"github.com/gin-gonic/gin"
)

type ResourceMeta struct {
	Filename  string `json:"filename"`
	Mimetype  string `json:"mimetype"`
	Extension string `json:"extension"`
}

type ResourceMetas []ResourceMeta

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
	var meta ResourceMetas
	err := json.Unmarshal([]byte(c.PostForm("meta")), &meta)
	if err != nil {
		setNotification(c, nError, err.Error(), false)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	for i, m := range meta {
		resourceFile, _, err := c.Request.FormFile(fmt.Sprintf("resource%d", i))
		if err != nil {
			setNotification(c, nError, err.Error(), false)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		resource, err := io.ReadAll(resourceFile)
		if err != nil {
			setNotification(c, nError, err.Error(), false)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if !bytes.Equal(resource, []byte("")) {
			key := storage.Hash(resource) + "." + m.Extension
			err = storage.SaveResource(key, resource)
			if err != nil {
				setNotification(c, nError, err.Error(), false)
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})
				return
			}
			size := storage.GetResourceSize(key)
			s.Size += size
			s.Resources = append(s.Resources, model.GetOrCreateResource(key, m.Filename, m.Mimetype, size))
		}
	}
	model.DB.Save(s)
	c.JSON(200, map[string]interface{}{
		"success": true,
	})
}
