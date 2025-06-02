package webapp

import (
	"net/http"

	"github.com/asciimoo/omnom/model"

	"github.com/gin-gonic/gin"
	//"github.com/rs/zerolog/log"
)

func editCollection(c *gin.Context) {
	cid, _ := c.GetQuery("cid")
	u, _ := c.Get("user")
	col := model.GetCollection(u.(*model.User).ID, cid)
	var parents []*model.Collection
	model.DB.Where("parent_id is NULL or parent_id == 0").Where("user_id = ?", u.(*model.User).ID).Order("name desc").Find(&parents)
	render(c, http.StatusOK, "edit-collection", gin.H{
		"Collection": col,
		"Parents":    parents,
	})
}

func saveCollection(c *gin.Context) {
	cid := c.PostForm("cid")
	pcid := c.PostForm("parent_cid")
	u, _ := c.Get("user")
	col := model.GetCollection(u.(*model.User).ID, cid)
	if col == nil {
		col = &model.Collection{}
	}
	col.Name = c.PostForm("name")
	if col.Name == "" {
		setNotification(c, nError, "Invalid collection name", true)
		c.Redirect(http.StatusFound, URLFor("edit collection form")+"?cid="+cid)
		return
	}
	parent := model.GetCollection(u.(*model.User).ID, pcid)
	if parent != nil && col.ID != parent.ID {
		col.ParentID = parent.ID
	} else {
		col.ParentID = 0
	}
	col.UserID = u.(*model.User).ID
	err := model.DB.Save(col).Error
	if err != nil {
		setNotification(c, nError, "Failed to save collection: "+err.Error(), true)
		c.Redirect(http.StatusFound, URLFor("edit collection form")+"?cid="+cid)
		return
	}
	setNotification(c, nInfo, "Save success", true)
	c.Redirect(http.StatusFound, URLFor("my bookmarks"))
}

func showCollections(c *gin.Context) {
	u := model.GetUserBySubmissionToken(c.PostForm("token"))
	if u == nil {
		c.IndentedJSON(http.StatusOK, make([]interface{}, 0))
		return
	}
	c.IndentedJSON(http.StatusOK, model.GetCollections(u.ID))
}
