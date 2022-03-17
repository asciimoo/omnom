package model

import (
	"gorm.io/gorm"
)

type Bookmark struct {
	gorm.Model
	ID        uint `gorm:"primaryKey"`
	URL       string
	Title     string
	Notes     string
	Domain    string
	Favicon   string
	Tags      []Tag `gorm:"many2many:bookmark_tags;"`
	Snapshots []Snapshot
	Public    bool
	UserID    uint
}
