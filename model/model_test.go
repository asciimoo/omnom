// SPDX-FileCopyrightText: 2021-2022 Adam Tauber, <asciimoo@gmail.com> et al.
//
// SPDX-License-Identifier: AGPL-3.0-only

package model

import (
	"os"
	"testing"

	"github.com/asciimoo/omnom/config"
)

func TestSQLiteInit(t *testing.T) {
	db := "./test_db.sqlite3"
	defer os.Remove(db)
	cfg := config.Config{
		DB: config.DB{
			Connection: db,
			Type:       "sqlite",
		},
	}
	err := Init(&cfg)
	if err != nil {
		t.Errorf("Failed to initialize SQLite DB: %s", err)
	}
}
