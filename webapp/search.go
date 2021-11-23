package webapp

import (
	"fmt"
	"net/http"

	"github.com/asciimoo/omnom/model"

	"github.com/gin-gonic/gin"
)

type searchParams struct {
	Q                string `form:"query"`
	Owner            string `form:"owner"`
	FromDate         string `form:"from"`
	ToDate           string `form:"to"`
	Tag              string `form:"tag"`
	Domain           string `form:"domain"`
	IsPublic         bool   `form:"public"`
	SearchInSnapshot bool   `form:"search_in_snapshot"`
	SearchInNote     bool   `form:"search_in_note"`
}

func search(c *gin.Context) {
	sp := &searchParams{}
	if err := c.ShouldBind(sp); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	u, _ := c.Get("user")
	var bs []*model.Bookmark
	var bookmarkCount int64
	pageno := getPageno(c)
	offset := (pageno - 1) * bookmarksPerPage
	q := model.DB.Model(&model.Bookmark{}).Limit(int(bookmarksPerPage)).Offset(int(offset)).Preload("Snapshots")
	cq := model.DB.Model(&model.Bookmark{})
	if sp.Q != "" {
		q = q.Where("LOWER(title) LIKE LOWER(?)", fmt.Sprintf("%%%s%%", sp.Q))
		cq = cq.Where("LOWER(title) LIKE LOWER(?)", fmt.Sprintf("%%%s%%", sp.Q))
		if sp.SearchInNote {
			q = q.Or("LOWER(notes) LIKE LOWER(?)", fmt.Sprintf("%%%s%%", sp.Q))
			cq = cq.Or("LOWER(notes) LIKE LOWER(?)", fmt.Sprintf("%%%s%%", sp.Q))
		}
	}
	if sp.IsPublic || u == nil {
		q = q.Where("public == true")
		cq = cq.Where("public == true")
	}
	if u != nil {
		q = q.Where("user_id == ? or public == true", u.(*model.User).ID)
		cq = cq.Where("user_id == ? or public == true", u.(*model.User).ID)
	}
	if sp.Domain != "" {
		q = q.Where("domain LIKE ?", fmt.Sprintf("%%%s%%", sp.Domain))
		cq = cq.Where("domain LIKE ?", fmt.Sprintf("%%%s%%", sp.Domain))
	}
	cq.Count(&bookmarkCount)
	q.Order("created_at desc").Find(&bs)
	renderHTML(c, http.StatusOK, "search", map[string]interface{}{
		"BookmarkCount": bookmarkCount,
		"Bookmarks":     bs,
		"Pageno":        pageno,
		"SearchParams":  sp,
		"HasNextPage":   offset+bookmarksPerPage < bookmarkCount,
	})
}
