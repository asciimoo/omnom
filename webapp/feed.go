// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

package webapp

import (
	"net/http"
	"sort"
	"time"

	"github.com/asciimoo/omnom/config"
	"github.com/asciimoo/omnom/model"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func feeds(c *gin.Context) {
	u, _ := c.Get("user")
	uid := u.(*model.User).ID
	res, err := model.GetUserFeeds(uid)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get feeds")
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	cfg, _ := c.Get("config")
	ipp := cfg.(*config.Config).Feed.ItemsPerPage
	fis := model.GetUnreadFeedItems(uid, ipp)
	bis := model.GetUnreadBookmarkItems(uid, ipp)
	render(c, http.StatusOK, "feeds", map[string]interface{}{
		"Feeds":           res,
		"UnreadItems":     mergeUnreadItems(fis, bis, ipp),
		"UnreadItemCount": model.GetUnreadFeedItemCount(uid),
	})
}

func addFeed(c *gin.Context) {
	url := c.PostForm("url")
	name := c.PostForm("name")
	u, _ := c.Get("user")
	uid := u.(*model.User).ID
	err := model.AddFeed(name, url, uid)
	if err != nil {
		setNotification(c, nError, "Failed to save feed: "+err.Error(), true)
	} else {
		setNotification(c, nInfo, "Feed added", true)
	}
	c.Redirect(http.StatusFound, URLFor("feeds"))
}

func mergeUnreadItems(fs []*model.UnreadFeedItem, bs []*model.Bookmark, maxNum uint) []any {
	fsl := len(fs)
	ret := make([]any, fsl+len(bs))
	for i, v := range fs {
		ret[i] = v
	}
	for i, v := range bs {
		ret[fsl+i] = v
	}
	sort.Slice(ret, func(i, j int) bool {
		var c1 time.Time
		switch m := ret[i].(type) {
		case *model.Bookmark:
			c1 = m.CreatedAt
		case *model.UnreadFeedItem:
			c1 = m.CreatedAt
		}
		var c2 time.Time
		switch m := ret[j].(type) {
		case *model.Bookmark:
			c2 = m.CreatedAt
		case *model.UnreadFeedItem:
			c2 = m.CreatedAt
		}
		return c1.After(c2)
	})
	if len(ret) >= int(maxNum) {
		return ret[:maxNum]
	}
	return ret
}
