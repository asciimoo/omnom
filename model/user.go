// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

package model

import (
	"fmt"
)

type User struct {
	CommonFields
	Username         string      `gorm:"unique" json:"username"`
	Email            *string     `gorm:"unique" json:"email"`
	OAuthID          *string     `gorm:"unique" json:"-"`
	LoginToken       string      `json:"-"`
	SubmissionTokens []Token     `json:"-"`
	Bookmarks        []Bookmark  `json:"bookmarks"`
	Feeds            []*Feed     `gorm:"many2many:user_feeds;" json:"feeds"`
	FeedsItems       []*FeedItem `gorm:"many2many:user_feed_items;" json:"feed_items"`
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

func GetUserByOAuthID(id string) *User {
	var u User
	err := DB.Where(&User{OAuthID: &id}).First(&u).Error
	if err != nil {
		return nil
	}
	return &u
}

func GetUserBySubmissionToken(tok string) *User {
	if tok == "" {
		return nil
	}
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
	var dbemail *string
	if email != "" {
		dbemail = &email
	}
	u := &User{
		Username:   username,
		Email:      dbemail,
		OAuthID:    dbemail,
		LoginToken: GenerateToken(),
		SubmissionTokens: []Token{Token{
			Text: GenerateToken(),
		}},
	}
	return DB.Create(u).Error
}
