// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

package model

import (
	"errors"
	"net/url"
	"strings"
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
func GetOrCreateBookmark(u *User, urlString, title, tags, notes, public, favicon, collection string) (*Bookmark, bool, error) {
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
	if public != "" {
		b.Public = true
	}
	if tags != "" {
		b.Tags = make([]Tag, 0, 8)
		for _, t := range strings.Split(tags, ",") {
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
		Table("bookmarks").
		Joins("join users on bookmark.user_id == users.id").
		Where("users.id = ?", uid).
		Where("bookmarks.unread = ?", true).
		Order("bookmarks.id asc").
		Find(&res)
	return res
}
