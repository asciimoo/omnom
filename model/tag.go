package model

import (
	"gorm.io/gorm"
)

type Tag struct {
	gorm.Model
	ID        uint       `gorm:"primaryKey"`
	Text      string     `gorm:"unique"`
	Bookmarks []Bookmark `gorm:"many2many:bookmark_tags;"`
}

func GetOrCreateTag(tag string) Tag {
	var t Tag
	if err := DB.Where("text = ?", tag).First(&t).Error; err != nil {
		t = Tag{
			Text: tag,
		}
		DB.Create(&t)
	}
	return t
}
