package model

import (
	"crypto/rand"
	"fmt"

	"gorm.io/gorm"
)

type Token struct {
	gorm.Model
	ID     uint `gorm:"primaryKey"`
	UserID uint
	Text   string
}

func GenerateToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	tok := fmt.Sprintf("%x", b)
	return tok
}
