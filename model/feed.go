// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

package model

import (
	"github.com/rs/zerolog/log"
	"gorm.io/gorm/clause"
)

const (
	RSSFeed = "rss"
)

type Feed struct {
	CommonFields
	Name    string      `json:"name"`
	URL     string      `json:"gorm:"unique" url"`
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
	URL     string  `gorm:"uniqueIndex:feeditemuidx" json:"url"`
	Title   string  `json:"title"`
	Content string  `json:"content"`
	FeedID  uint    `gorm:"uniqueIndex:feeditemuidx" json:"feed_id"`
	Feed    *Feed   `json:"feed"`
	Users   []*User `gorm:"many2many:user_feed_items;" json:"-"`
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
	FeedName       string
	Favicon        string
	UserFeedItemID uint
}

func GetFeeds() ([]*Feed, error) {
	var res []*Feed
	err := DB.
		Model(&Feed{}).
		Order("id").
		Find(&res).Error
	return res, err
}

func GetUserFeeds(uid uint) ([]*UserFeed, error) {
	var res []*UserFeed
	err := DB.
		Table("user_feeds").
		Joins("join feeds on feeds.id == user_feeds.feed_id").
		Where("user_feeds.user_id = ?", uid).
		Order("user_feeds.name").
		Find(&res).Error
	return res, err
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
		Select("feed_items.*, user_feeds.name as feed_name, feeds.favicon, user_feed_items.id as user_feed_item_id").
		Table("feed_items").
		Joins("join user_feed_items on feed_items.id == user_feed_items.feed_item_id").
		Joins("join user_feeds on user_feeds.feed_id == feed_items.feed_id and user_feeds.user_id = ?", uid).
		Joins("join feeds on feeds.id == user_feeds.feed_id").
		Where("user_feed_items.user_id = ?", uid).
		Where("user_feed_items.unread = ?", true).
		Order("feed_items.id asc").
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
