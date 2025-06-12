// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

package webapp

import (
	"net/http"

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
	render(c, http.StatusOK, "feeds", map[string]interface{}{
		"Feeds":           res,
		"UnreadItems":     model.GetUnreadFeedItems(uid, 30),
		"UnreadItemCount": model.GetUnreadFeedItemCount(uid),
	})
}

func addFeed(c *gin.Context) {
	url := c.PostForm("url")
	name := c.PostForm("name")
	u, _ := c.Get("user")
	uid := u.(*model.User).ID
	err := model.AddFeed(name, url, model.RSSFeed, uid)
	if err != nil {
		setNotification(c, nError, "Failed to save feed: "+err.Error(), true)
	} else {
		setNotification(c, nInfo, "Feed added", true)
	}
	c.Redirect(http.StatusFound, URLFor("feeds"))
}
