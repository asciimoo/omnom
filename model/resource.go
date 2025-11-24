// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

package model

// Resource represents a webpage resource like images or stylesheets.
type Resource struct {
	CommonFields
	Key              string     `gorm:"unique" json:"key"`
	MimeType         string     `json:"mimeType"`
	OriginalFilename string     `json:"originalFilename"`
	Size             uint       `json:"size"`
	Snapshots        []Snapshot `gorm:"many2many:snapshot_resources;" json:"snapshots"`
}

// GetOrCreateResource retrieves an existing resource or creates a new one.
func GetOrCreateResource(key string, mimeType string, fname string, size uint) Resource {
	var r Resource
	if err := DB.Where("key = ?", key).First(&r).Error; err != nil {
		r = Resource{
			Key:              key,
			MimeType:         mimeType,
			OriginalFilename: fname,
			Size:             size,
		}
		DB.Create(&r)
	}
	return r
}
