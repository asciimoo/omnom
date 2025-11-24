// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

package model

import (
	"crypto/rand"
	"fmt"
)

// Token represents an API or submission token.
type Token struct {
	CommonFields
	UserID uint   `json:"user_id"`
	Text   string `json:"text"`
}

// GenerateToken generates a new random token string.
func GenerateToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	tok := fmt.Sprintf("%x", b)
	return tok
}

// CreateAddonToken creates a new addon token for a user.
func CreateAddonToken(uid uint) (*Token, error) {
	tok := &Token{
		Text:   GenerateToken(),
		UserID: uid,
	}
	return tok, DB.Create(tok).Error
}
