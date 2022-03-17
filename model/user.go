package model

import (
	"fmt"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID               uint   `gorm:"primaryKey"`
	Username         string `gorm:"unique"`
	Email            string `gorm:"unique"`
	LoginToken       string
	SubmissionTokens []Token
	Bookmarks        []Bookmark
}

func GetUser(name string) *User {
	var u User
	err := DB.Where("LOWER(username) == LOWER(?) or LOWER(email) == LOWER(?)", name, name).First(&u).Error
	if err != nil {
		return nil
	}
	return &u
}

func GetUserByLoginToken(tok string) *User {
	var u User
	err := DB.Where(&User{LoginToken: tok}).First(&u).Error
	if err != nil {
		return nil
	}
	return &u
}

func GetUserBySubmissionToken(tok string) *User {
	var u User
	err := DB.Model(&User{}).Select("users.*").Joins("left join tokens on tokens.user_id = users.id").Where("tokens.text = ?", tok).First(&u).Error
	if err != nil {
		return nil
	}

	return &u
}

func CreateUser(username, email string) error {
	if GetUser(username) != nil {
		return fmt.Errorf("User already exists")
	}
	u := &User{
		Username:   username,
		Email:      email,
		LoginToken: GenerateToken(),
		SubmissionTokens: []Token{Token{
			Text: GenerateToken(),
		}},
	}
	return DB.Create(u).Error
}
