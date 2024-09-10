// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

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
