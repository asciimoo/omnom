// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

package webapp

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/asciimoo/omnom/config"
	"github.com/asciimoo/omnom/feed"
	"github.com/asciimoo/omnom/model"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func feeds(c *gin.Context) {
	u, _ := c.Get("user")
	uid := u.(*model.User).ID
	fs, err := model.GetUserFeeds(uid, true)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get feeds")
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	cfg, _ := c.Get("config")
	ipp := cfg.(*config.Config).Feed.ItemsPerPage
	fis := model.GetUnreadFeedItems(uid, ipp)
	bis := model.GetUnreadBookmarkItems(uid, ipp)
	is := mergeUnreadItems(fis, bis, ipp)
	render(c, http.StatusOK, "feeds", map[string]any{
		"Feeds":           fs,
		"UnreadItems":     is,
		"UnreadItemCount": model.GetUnreadFeedItemCount(uid) + model.GetUnreadBookmarkCount(uid),
		"FeedItemIDs":     concatFeedItemIDs(is),
		"BookmarkIDs":     concatBookmarkIDs(is),
	})
}

func searchFeedItems(c *gin.Context) {
	u, _ := c.Get("user")
	uid := u.(*model.User).ID
	fs, err := model.GetUserFeeds(uid, false)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get feeds")
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	cfg, _ := c.Get("config")
	ipp := cfg.(*config.Config).Feed.ItemsPerPage
	q := c.Query("query")
	includeRead := false
	var fid uint
	if c.Query("feed_id") != "" {
		if i, err := strconv.ParseUint(c.Query("feed_id"), 10, 64); err == nil {
			fid = uint(i)
		}
	}
	if c.Query("include_read_items") != "" && c.Query("include_read_items") != "0" && c.Query("include_read_items") != "false" {
		includeRead = true
	}
	res, resCount, err := model.SearchFeedItems(uid, ipp, q, fid, includeRead)
	if err != nil {
		log.Error().Err(err).Msg("Failed to filter feed items")
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	fn := ""
	if fid > 0 {
		for _, f := range fs {
			if f.ID == fid {
				fn = f.Name
			}
		}
	}
	render(c, http.StatusOK, "feed-search", map[string]any{
		"Feeds":       fs,
		"Items":       res,
		"ItemCount":   resCount,
		"IncludeRead": includeRead,
		"Query":       q,
		"FeedID":      fid,
		"FeedName":    fn,
	})
}

func archiveItems(c *gin.Context) {
	u, _ := c.Get("user")
	uid := u.(*model.User).ID
	fids := c.PostForm("fids")
	var rows int64
	if fids != "" {
		rows += model.DB.Table("user_feed_items").Where("user_id = ? AND id IN ?", uid, sliceAtoi(strings.Split(fids, ","))).Update("unread", false).RowsAffected
	}
	bids := c.PostForm("bids")
	if bids != "" {
		rows += model.DB.Model(&model.Bookmark{}).Where("user_id = ? AND id IN ?", uid, sliceAtoi(strings.Split(bids, ","))).Update("unread", false).RowsAffected
	}
	if rows > 0 {
		setNotification(c, nInfo, "Items archived", true)
	}
	c.Redirect(http.StatusFound, URLFor("feeds"))
}

func editFeedForm(c *gin.Context) {
	f, err := getUserFeedOrAbort(c)
	if err != nil || f == nil {
		return
	}
	render(c, http.StatusOK, "edit-feed", map[string]any{
		"Feed": f,
	})
}

func editFeed(c *gin.Context) {
	f, err := getUserFeedOrAbort(c)
	if err != nil {
		return
	}
	if c.PostForm("name") != "" {
		f.Name = c.PostForm("name")
	}
	// TODO resolve activitypub feed changes
	err = model.DB.Save(f).Error
	if err == nil {
		setNotification(c, nInfo, "Feed saved", true)
	} else {
		setNotification(c, nError, "Failed to save feed", true)
	}
	render(c, http.StatusOK, "edit-feed", map[string]any{
		"Feed": f,
	})
}

func deleteFeed(c *gin.Context) {
	f, err := getUserFeedOrAbort(c)
	if err != nil {
		return
	}
	cfg, _ := c.Get("config")
	err = feed.DeleteFeed(cfg.(*config.Config), f)
	if err == nil {
		setNotification(c, nInfo, "Feed deleted", true)
	} else {
		log.Error().Err(err).Str("Feed", f.Name).Msg("Failed to delete feed")
		setNotification(c, nError, "Failed to delete feed", true)
	}
	c.Redirect(http.StatusFound, URLFor("feeds"))
}

func addFeed(c *gin.Context) {
	url := c.PostForm("url")
	name := c.PostForm("name")
	u, _ := c.Get("user")
	uid := u.(*model.User).ID
	cfg, _ := c.Get("config")
	err := feed.AddFeed(cfg.(*config.Config), name, url, uid)
	if err != nil {
		setNotification(c, nError, "Failed to save feed: "+err.Error(), true)
	} else {
		setNotification(c, nInfo, "Feed added", true)
	}
	c.Redirect(http.StatusFound, URLFor("feeds"))
}

func getUserFeedOrAbort(c *gin.Context) (*model.UserFeed, error) {
	u, _ := c.Get("user")
	uid := u.(*model.User).ID
	fid := c.Query("id")
	if fid == "" {
		fid = c.PostForm("id")
	}
	f, err := model.GetUserFeed(uid, fid)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get feed")
		c.Redirect(http.StatusFound, URLFor("feeds"))
		return nil, err
	}
	return f, nil
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
		return c1.Before(c2)
	})
	if len(ret) >= int(maxNum) { //nolint: gosec // safe
		return ret[:maxNum]
	}
	return ret
}

func concatFeedItemIDs(is []any) string {
	var ids = []string{}
	for _, i := range is {
		switch v := i.(type) {
		case *model.UnreadFeedItem:
			ids = append(ids, fmt.Sprintf("%d", v.UserFeedItemID))
		}
	}
	return strings.Join(ids, ",")
}

func concatBookmarkIDs(is []any) string {
	var ids = []string{}
	for _, i := range is {
		switch v := i.(type) {
		case *model.Bookmark:
			ids = append(ids, fmt.Sprintf("%d", v.ID))
		}
	}
	return strings.Join(ids, ",")
}

func sliceAtoi(s []string) []int {
	var l = []int{}
	for _, i := range s {
		j, err := strconv.Atoi(i)
		if err != nil {
			continue
		}
		l = append(l, j)
	}
	return l
}
