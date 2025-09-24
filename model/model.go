// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

package model

import (
	"fmt"
	"time"

	"github.com/asciimoo/omnom/config"
	"github.com/asciimoo/omnom/utils"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
)

type DBTypedef int

const (
	Sqlite DBTypedef = iota
	Psql
)

const DBTypeErr = utils.ErrString("Unknown database type")

var DB *gorm.DB
var DBType = Sqlite

func Init(c *config.Config) error {
	dbCfg := &gorm.Config{}
	if c.App.DebugSQL {
		dbCfg.Logger = logger.Default.LogMode(logger.Info)
	} else {
		dbCfg.Logger = logger.Default.LogMode(logger.Silent)
	}
	var err error
	switch c.DB.Type {
	case "sqlite", "sqlite3":
		DB, err = gorm.Open(sqlite.Open(c.DB.Connection), dbCfg)
		if err != nil {
			return err
		}
		DBType = Sqlite
	case "postgresql", "postgres", "psql":
		DB, err = gorm.Open(postgres.Open(c.DB.Connection), dbCfg)
		if err != nil {
			return err
		}
		DBType = Psql
	default:
		return DBTypeErr
	}
	err = migrate()
	if err != nil {
		return fmt.Errorf("custom migration of database '%s' has failed: %w", c.DB.Connection, err)
	}
	err = DB.SetupJoinTable(Feed{}, "Users", &UserFeed{})
	if err != nil {
		return fmt.Errorf("failed to setup join table for users and feeds: %w", err)
	}
	err = DB.SetupJoinTable(FeedItem{}, "Users", &UserFeedItem{})
	if err != nil {
		return fmt.Errorf("failed to setup join table for users and feed items: %w", err)
	}
	err = automigrate()
	if err != nil {
		return fmt.Errorf("auto migration of database '%s' has failed: %w", c.DB.Connection, err)
	}
	return nil
}

func automigrate() error {
	return DB.AutoMigrate(
		&User{},
		&Bookmark{},
		&Snapshot{},
		&Tag{},
		&Token{},
		&Database{},
		&Resource{},
		&APFollower{},
		&Collection{},
		&Feed{},
		&FeedItem{},
		&UserFeed{},
		&UserFeedItem{},
	)
}

type Database struct {
	ID      uint `gorm:"primaryKey"`
	Version uint
}

type CommonFields struct {
	ID        uint       `gorm:"primary_key" json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}
