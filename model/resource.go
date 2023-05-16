// SPDX-FileCopyrightText: 2021-2022 Adam Tauber, <asciimoo@gmail.com> et al.
//
// SPDX-License-Identifier: AGPL-3.0-only

package model

type Resource struct {
	CommonFields
	Key              string     `gorm:"unique" json:"key"`
	MimeType         string     `json:"mimeType"`
	OriginalFilename string     `json:"originalFilename"`
	Size             uint       `json:"size"`
	Snapshots        []Snapshot `gorm:"many2many:snapshot_resources;" json:"snapshots"`
}

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
