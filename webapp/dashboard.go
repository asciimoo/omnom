// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

package webapp

import (
	"net/http"
	"time"

	"github.com/asciimoo/omnom/model"

	"github.com/gin-gonic/gin"
)

func dashboard(c *gin.Context, u *model.User) {
	var weeklyBookmarkCount int64
	var monthlyBookmarkCount int64
	var yearlyBookmarkCount int64
	var bs []*model.Bookmark
	now := time.Now()
	var tags []*model.TagCount
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	model.DB.Model(&model.Bookmark{}).Where("bookmarks.user_id = ? and bookmarks.updated_at > ? and bookmarks.updated_at < ?", u.ID, today.Truncate(time.Hour*7*24), now).Count(&weeklyBookmarkCount)
	model.DB.Model(&model.Bookmark{}).Where("bookmarks.user_id = ? and bookmarks.updated_at > ? and bookmarks.updated_at < ?", u.ID, today.AddDate(0, -1, 0), now).Count(&monthlyBookmarkCount)
	model.DB.Model(&model.Bookmark{}).Where("bookmarks.user_id = ? and bookmarks.updated_at > ? and bookmarks.updated_at < ?", u.ID, today.AddDate(-1, 0, 0), now).Count(&yearlyBookmarkCount)
	_ = model.DB.Limit(10).Model(u).Preload("Snapshots").Preload("Tags").Preload("User").Order("updated_at desc").Association("Bookmarks").Find(&bs)
	model.DB.Limit(20).Table("tags").Select("tags.text as tag, count(bookmarks.user_id) as `count`").Joins("join bookmark_tags on bookmark_tags.tag_id == tags.id").Joins("join bookmarks on bookmarks.id == bookmark_tags.bookmark_id").Where("bookmarks.user_id = ?", u.ID).Group("tags.text").Order("`count` desc, tag asc").Find(&tags)
	render(c, http.StatusOK, "dashboard", map[string]interface{}{
		"WeeklyBookmarkCount":  weeklyBookmarkCount,
		"MonthlyBookmarkCount": monthlyBookmarkCount,
		"YearlyBookmarkCount":  yearlyBookmarkCount,
		"Bookmarks":            bs,
		"Tags":                 tags,
	})
}
