// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

package model

type Feed struct {
	CommonFields
	Name  string      `json:"name"`
	URL   string      `json:"url"`
	Type  string      `json:"type"`
	Items []*FeedItem `json:"items"`
	Users []*User     `gorm:"many2many:user_feeds;" json:"-"`
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
	URL     string  `json:"url"`
	Title   string  `json:"title"`
	Content string  `json:"content"`
	FeedID  uint    `json:"feed_id"`
	Feed    *Feed   `json:"feed"`
	UserID  uint    `json:"user_id"`
	Users   []*User `gorm:"many2many:user_feed_items;" json:"-"`
}

type UserFeedItem struct {
	CommonFields
	Unread     bool      `json:"unread"`
	FeedItemID uint      `json:"feed_item_id"`
	FeedItem   *FeedItem `json:"feed_item"`
	UserID     uint      `json:"user_id"`
	User       *User     `json:"-"`
}
