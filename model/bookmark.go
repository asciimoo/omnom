// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

package model

import (
	"database/sql"
	"errors"
	"net/url"
	"strings"

	"gorm.io/gorm"
)

type Bookmark struct {
	CommonFields
	URL          string      `json:"url"`
	Title        string      `json:"title"`
	Notes        string      `json:"notes"`
	Domain       string      `json:"domain"`
	Favicon      string      `json:"favicon"`
	Tags         []Tag       `gorm:"many2many:bookmark_tags;" json:"tags"`
	Snapshots    []Snapshot  `json:"snapshots"`
	CollectionID uint        `json:"-"`
	Collection   *Collection `json:"collection"`
	Public       bool        `json:"public"`
	Unread       bool        `json:"unread"`
	UserID       uint        `json:"user_id"`
	User         User        `json:"-"`
}

// TODO use Bookmark as parameter instead of strings
func GetOrCreateBookmark(u *User, urlString, title, tags, notes, public, favicon, collection, unread string) (*Bookmark, bool, error) {
	url, err := url.Parse(urlString)
	new := false
	if err != nil || url.Hostname() == "" || url.Scheme == "" {
		return nil, new, errors.New("invalid URL")
	}
	var b *Bookmark = nil
	r := DB.
		Model(&Bookmark{}).
		Preload("Snapshots").
		Preload("Tags").
		Preload("User").
		Where("url = ? and user_id = ?", url.String(), u.ID).
		First(&b)
	if r.RowsAffected >= 1 {
		return b, new, nil
	} else {
		new = true
	}
	if title == "" {
		return nil, new, errors.New("missing title")
	}
	b = &Bookmark{
		Title:     title,
		URL:       url.String(),
		Domain:    url.Hostname(),
		Notes:     notes,
		Favicon:   favicon,
		UserID:    u.ID,
		User:      *u,
		Snapshots: make([]Snapshot, 0, 8),
	}
	if !strings.HasPrefix(b.Favicon, "data:image") {
		b.Favicon = ""
	}
	if public != "" && public != "0" {
		b.Public = true
	}
	if unread != "" && unread != "0" {
		b.Unread = true
	}
	if tags != "" {
		b.Tags = make([]Tag, 0, 8)
		for t := range strings.SplitSeq(tags, ",") {
			t = strings.TrimSpace(t)
			b.Tags = append(b.Tags, GetOrCreateTag(t))
		}
	}
	col := GetCollection(u.ID, collection)
	if col != nil {
		b.CollectionID = col.ID
	}
	if err := DB.Save(b).Error; err != nil {
		return nil, new, err
	}
	return b, new, nil
}

func GetUnreadBookmarkItems(uid, limit uint) []*Bookmark {
	var res []*Bookmark
	DB.
		Model(&Bookmark{}).
		Joins("join users on bookmarks.user_id = users.id").
		Where("users.id = ?", uid).
		Where("bookmarks.unread = ?", true).
		Order("bookmarks.id asc").
		Find(&res)
	return res
}

func GetUnreadBookmarkCount(uid uint) int64 {
	var res int64
	DB.
		Table("bookmarks").
		Where("bookmarks.user_id = ?", uid).
		Where("bookmarks.unread = ?", true).
		Count(&res)
	return res
}

func SearchBookmarks(uid, limit uint, query string) ([]*Bookmark, int64, error) {
	var res []*Bookmark
	var resCount int64
	q := DB.Select("*").Table("bookmarks")
	if uid == 0 {
		q = q.Where("bookmarks.public = 1")
	} else {
		q = q.Where("bookmarks.public = 1 or bookmarks.user_id = ?", uid)
	}
	if query != "" {
		q = q.Where("bookmarks.title LIKE LOWER(@query) OR bookmarks.notes LIKE LOWER(@query)", sql.Named("query", CreateGlob(query)))
	}
	q = q.Session(&gorm.Session{})
	err := q.Preload("Snapshots").Preload("Tags").Preload("User").Preload("Collection").Order("bookmarks.id asc").Limit(int(limit)).Find(&res).Error //nolint:gosec // TODO
	if err != nil {
		return nil, 0, err
	}
	err = q.Count(&resCount).Error
	if err != nil {
		return nil, 0, err
	}
	return res, resCount, nil
}
