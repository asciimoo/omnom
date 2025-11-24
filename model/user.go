// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

package model

import (
	"fmt"
)

// User represents a user account.
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

// GetUser retrieves a user by username or email.
func GetUser(name string) *User {
	var u User
	err := DB.Where("LOWER(username) == LOWER(?) or LOWER(email) == LOWER(?)", name, name).First(&u).Error
	if err != nil {
		return nil
	}
	return &u
}

// GetUserByLoginToken retrieves a user by their login token.
func GetUserByLoginToken(tok string) *User {
	var u User
	err := DB.Where(&User{LoginToken: tok}).First(&u).Error
	if err != nil {
		return nil
	}
	return &u
}

// GetUserByOAuthID retrieves a user by their OAuth ID.
func GetUserByOAuthID(id string) *User {
	var u User
	err := DB.Where(&User{OAuthID: &id}).First(&u).Error
	if err != nil {
		return nil
	}
	return &u
}

// GetUserBySubmissionToken retrieves a user by their submission token.
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

// CreateUser creates a new user with the specified username and email.
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
