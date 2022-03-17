package model

import (
	"gorm.io/gorm"
)

type Resource struct {
	gorm.Model
	ID               uint   `gorm:"primaryKey"`
	Key              string `gorm:"unique"`
	MimeType         string
	OriginalFilename string
	Size             uint
	Snapshots        []Snapshot `gorm:"many2many:snapshot_resources;"`
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
