// SPDX-FileCopyrightText: 2021-2022 Adam Tauber, <asciimoo@gmail.com> et al.
//
// SPDX-License-Identifier: AGPL-3.0-only

package model

type Snapshot struct {
	CommonFields
	Title      string     `json:"title"`
	Key        string     `json:"key"`
	Text       string     `json:"text"`
	BookmarkID uint       `json:"bookmark_id"`
	Size       uint       `json:"size"`
	Resources  []Resource `gorm:"many2many:snapshot_resources;" json:"resources"`
}
