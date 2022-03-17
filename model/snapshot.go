package model

import (
	"gorm.io/gorm"
)

type Snapshot struct {
	gorm.Model
	ID         uint `gorm:"primaryKey"`
	Title      string
	Key        string
	Text       string
	BookmarkID uint
	Size       uint
	Resources  []Resource `gorm:"many2many:snapshot_resources;"`
}
