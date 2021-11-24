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
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	model.DB.Model(&model.Bookmark{}).Where("bookmarks.user_id = ? and bookmarks.created_at > ? and bookmarks.created_at < ?", u.ID, today.Truncate(time.Hour*7*24), now).Count(&weeklyBookmarkCount)
	model.DB.Model(&model.Bookmark{}).Where("bookmarks.user_id = ? and bookmarks.created_at > ? and bookmarks.created_at < ?", u.ID, today.AddDate(0, -1, 0), now).Count(&monthlyBookmarkCount)
	model.DB.Model(&model.Bookmark{}).Where("bookmarks.user_id = ? and bookmarks.created_at > ? and bookmarks.created_at < ?", u.ID, today.AddDate(-1, 0, 0), now).Count(&yearlyBookmarkCount)
	model.DB.Limit(10).Model(u).Preload("Snapshots").Preload("Tags").Order("created_at desc").Association("Bookmarks").Find(&bs)
	renderHTML(c, http.StatusOK, "dashboard", map[string]interface{}{
		"WeeklyBookmarkCount":  weeklyBookmarkCount,
		"MonthlyBookmarkCount": monthlyBookmarkCount,
		"YearlyBookmarkCount":  yearlyBookmarkCount,
		"Bookmarks":            bs,
	})
}
