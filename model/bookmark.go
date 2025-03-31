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
	URL       string     `json:"url"`
	Title     string     `json:"title"`
	Notes     string     `json:"notes"`
	Domain    string     `json:"domain"`
	Favicon   string     `json:"favicon"`
	Tags      []Tag      `gorm:"many2many:bookmark_tags;" json:"tags"`
	Snapshots []Snapshot `json:"snapshots"`
	Public    bool       `json:"public"`
	UserID    uint       `json:"user_id"`
	User      User       `json:"-"`
}

func GetOrCreateBookmark(u *User, urlString, title, tags, notes, public, favicon string) (*Bookmark, error) {
	url, err := url.Parse(urlString)
	if err != nil || url.Hostname() == "" || url.Scheme == "" {
		return nil, errors.New("invalid URL")
	}
	var b *Bookmark = nil
	r := DB.
		Model(&Bookmark{}).
		Preload("Snapshots").
		Where("url = ? and user_id = ?", url.String(), u.ID).
		First(&b)
	if r.RowsAffected >= 1 {
		return b, nil
	}
	if title == "" {
		return nil, errors.New("missing title")
	}
	b = &Bookmark{
		Title:     title,
		URL:       url.String(),
		Domain:    url.Hostname(),
		Notes:     notes,
		Favicon:   favicon,
		UserID:    u.ID,
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
	if err := DB.Save(b).Error; err != nil {
		return nil, err
	}
	return b, nil
}
