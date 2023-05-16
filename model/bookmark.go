// SPDX-FileCopyrightText: 2021-2022 Adam Tauber, <asciimoo@gmail.com> et al.
//
// SPDX-License-Identifier: AGPL-3.0-only

package model

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
