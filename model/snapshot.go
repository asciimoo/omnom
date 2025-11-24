// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

package model

// Snapshot represents a saved webpage snapshot.
type Snapshot struct {
	CommonFields
	Title      string     `json:"title"`
	Key        string     `json:"key"`
	Text       string     `json:"text"`
	BookmarkID uint       `json:"bookmark_id"`
	Bookmark   Bookmark   `json:"bookmark"`
	Size       uint       `json:"size"`
	Resources  []Resource `gorm:"many2many:snapshot_resources;" json:"resources"`
}

// GetSnapshotWithResources retrieves a snapshot with its associated resources.
func GetSnapshotWithResources(key string) (*Snapshot, error) {
	var s *Snapshot
	if err := DB.Where("key = ?", key).Preload("Resources").First(&s).Error; err != nil {
		return nil, err
	}
	return s, nil
}
