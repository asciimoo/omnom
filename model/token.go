// SPDX-FileCopyrightText: 2021-2022 Adam Tauber, <asciimoo@gmail.com> et al.
//
// SPDX-License-Identifier: AGPL-3.0-only

package model

import (
	"crypto/rand"
	"fmt"
)

type Token struct {
	CommonFields
	UserID uint   `json:"user_id"`
	Text   string `json:"text"`
}

func GenerateToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	tok := fmt.Sprintf("%x", b)
	return tok
}
