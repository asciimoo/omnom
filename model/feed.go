// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

package model

import (
	"database/sql"
	"errors"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type FeedType string

const (
	RSSFeed         FeedType = "rss"
	ActivityPubFeed FeedType = "ap"
)

type Feed struct {
	CommonFields
	Name    string      `json:"name"`
	URL     string      `json:"gorm:"unique" url"`
	Author  string      `json:"author"`
	Type    string      `json:"type"`
	Favicon string      `json:"favicon"`
	Items   []*FeedItem `json:"items"`
	Users   []*User     `gorm:"many2many:user_feeds;" json:"-"`
}

type UserFeed struct {
	CommonFields
	Name   string `json:"name"`
	Public bool   `json:"public"`
	FeedID uint   `json:"feed_id"`
	Feed   *Feed  `json:"feed"`
	UserID uint   `json:"user_id"`
	User   *User  `json:"-"`
}

type FeedItem struct {
	CommonFields
	URL                string  `gorm:"uniqueIndex:feeditemuidx" json:"url"`
	Title              string  `json:"title"`
	Content            string  `json:"content"`
	OriginalAuthorID   string  `json:"original_author_id"`
	OriginalAuthorName string  `json:"original_author_name"`
	Favicon            string  `json:"favicon"`
	FeedID             uint    `gorm:"uniqueIndex:feeditemuidx" json:"feed_id"`
	Feed               *Feed   `json:"feed"`
	Users              []*User `gorm:"many2many:user_feed_items;" json:"-"`
}

type UserFeedItem struct {
	CommonFields
	Unread     bool      `json:"unread"`
	FeedItemID uint      `gorm:"uniqueIndex:userfeeditemuidx" json:"feed_item_id"`
	FeedItem   *FeedItem `json:"feed_item"`
	UserID     uint      `gorm:"uniqueIndex:userfeeditemuidx" json:"user_id"`
	User       *User     `json:"-"`
}

type UnreadFeedItem struct {
	FeedItem
	FeedName       string `json:"feed_name"`
	FeedAuthor     string `json:"feed_author"`
	FeedURL        string `json:"feed_url"`
	FeedType       string `json:"feed_type"`
	FeedFavicon    string `json:"feed_favicon"`
	UserFeedItemID uint
	Unread         bool
}

type UserFeedSummary struct {
	UserFeed
	Count uint
}

func GetFeeds() ([]*Feed, error) {
	var res []*Feed
	err := DB.
		Model(&Feed{}).
		Order("id").
		Find(&res).Error
	return res, err
}

func GetFeedItem(fid uint, u string) (*FeedItem, error) {
	var i *FeedItem
	err := DB.
		Model(&FeedItem{}).
		Where("feed_id = ? AND url = ?", fid, u).
		First(&i).Error
	return i, err
}

func GetUserFeeds(uid uint, unread bool) ([]*UserFeedSummary, error) {
	var res []*UserFeedSummary
	q := DB.
		Table("user_feeds")
	if unread {
		q = q.Select("user_feeds.*, sum(user_feed_items.unread) as count")
	} else {
		q = q.Select("user_feeds.*, count(user_feed_items.id) as count")
	}
	err := q.Joins("join feeds on feeds.id == user_feeds.feed_id").
		Joins("left join feed_items on feed_items.feed_id == feeds.id").
		Joins("left join user_feed_items on user_feed_items.feed_item_id == feed_items.id").
		Where("user_feeds.user_id = ?", uid).
		Group("feeds.id").
		Order("count desc, user_feeds.name").
		Find(&res).Error
	return res, err
}

func GetUserFeed(uid uint, fid string) (*UserFeed, error) {
	var f *UserFeed
	res := DB.
		Table("user_feeds").
		Select("*").
		Where("user_id = ? and id = ?", uid, fid).
		Find(&f)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, errors.New("no feed found")
	}
	return f, nil
}

func DeleteUserFeed(f *UserFeed) error {
	if err := DB.Delete(
		&UserFeedItem{},
		"id in (?)",
		DB.Table("user_feed_items").
			Select("user_feed_items.id").
			Joins("join feed_items on user_feed_items.feed_item_id = feed_items.id").
			Joins("join feeds on feeds.id == feed_items.feed_id").
			Joins("join user_feeds on user_feeds.feed_id == feeds.id").
			Where("user_feed_items.user_id = ? and user_feeds.id = ?", f.UserID, f.ID),
	).Error; err != nil {
		return err
	}
	res := DB.Delete(f, "id = ?", f.ID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return nil
	}
	var ufCount int64
	if err := DB.Table("user_feeds").Where("feed_id = ?", f.FeedID).Count(&ufCount).Error; err != nil {
		return err
	}
	if ufCount == 0 {
		return DB.Delete(&Feed{}, "id = ?", f.FeedID).Error
	}
	return nil
}

func GetFeedByURL(u string) (*Feed, error) {
	var f *Feed
	err := DB.Table("feeds").
		Preload("Users").
		Where("feeds.url = ?", u).First(&f).Error
	return f, err
}

func GetFeedByID(id uint) (*Feed, error) {
	var f *Feed
	err := DB.Table("feeds").
		Preload("Users").
		Where("feeds.id = ?", id).First(&f).Error
	return f, err
}

func AddFeedItem(i *FeedItem) int64 {
	f, err := GetFeedByID(i.FeedID)
	if f == nil || err != nil {
		log.Error().Uint("ID", i.FeedID).Msg("Feed not found")
		return 0
	}
	err = DB.Create(i).Error
	// TODO accept only UNIQUE constraint failed
	// According to docs it is type of  gorm.ErrDuplicatedKey, but it does not work
	if err != nil {
		DB.Where("feed_id = ? and url = ?", f.ID, i.URL).First(&i)
	}
	uis := make([]*UserFeedItem, len(f.Users))
	for n, u := range f.Users {
		uis[n] = &UserFeedItem{
			UserID:     u.ID,
			FeedItemID: i.ID,
			Unread:     true,
		}
	}
	return DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&uis).RowsAffected
}

func GetUnreadFeedItems(uid, limit uint) []*UnreadFeedItem {
	var res []*UnreadFeedItem
	DB.
		Select("feed_items.*, user_feeds.name as feed_name, feeds.feed_author as feed_author, feeds.url as feed_url, feeds.type as feed_type, feeds.favicon as feed_favicon, user_feed_items.id as user_feed_item_id, user_feed_items.unread as unread").
		Table("feed_items").
		Joins("join user_feed_items on feed_items.id == user_feed_items.feed_item_id").
		Joins("join user_feeds on user_feeds.feed_id == feed_items.feed_id and user_feeds.user_id = ?", uid).
		Joins("join feeds on feeds.id == user_feeds.feed_id").
		Where("user_feed_items.user_id = ?", uid).
		Where("user_feed_items.unread = ?", true).
		Order("feed_items.id asc").
		Limit(int(limit)). //nolint:gosec // TODO
		Find(&res)
	return res
}

func GetUnreadFeedItemCount(uid uint) int64 {
	var res int64
	DB.
		Table("feed_items").
		Joins("join user_feed_items on feed_items.id == user_feed_items.feed_item_id").
		Where("user_feed_items.user_id = ?", uid).
		Where("user_feed_items.unread = ?", true).
		Count(&res)
	return res
}

func SearchFeedItems(uid, limit uint, query string, feedID uint, includeRead bool) ([]*UnreadFeedItem, int64, error) {
	var res []*UnreadFeedItem
	var resCount int64
	q := DB.
		Select("feed_items.*, user_feeds.name as feed_name, feeds.feed_author as feed_author, feeds.url as feed_url, feeds.favicon as feed_favicon, user_feed_items.id as user_feed_item_id, user_feed_items.unread as unread").
		Table("feed_items").
		Joins("join user_feed_items on feed_items.id == user_feed_items.feed_item_id").
		Joins("join user_feeds on user_feeds.feed_id == feed_items.feed_id and user_feeds.user_id = ?", uid).
		Joins("join feeds on feeds.id == user_feeds.feed_id").
		Where("user_feed_items.user_id = ?", uid)
	if feedID != 0 {
		q = q.Where("user_feeds.id == ?", feedID)
	}
	if query != "" {
		q = q.Where("feed_items.title LIKE LOWER(@query) OR feed_items.content LIKE LOWER(@query)", sql.Named("query", CreateGlob(query)))
	}
	if !includeRead {
		q = q.Where("user_feed_items.unread = ?", true)
	}
	q = q.Session(&gorm.Session{})
	q.Order("feed_items.id asc").Limit(int(limit)).Find(&res) //nolint:gosec // TODO
	q.Count(&resCount)
	return res, resCount, nil
}
