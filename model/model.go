// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

// Package model provides database models and data access layer for Omnom.
//
// This package defines all database entities and their relationships, including:
//   - Users: User accounts with authentication tokens
//   - Bookmarks: Saved web pages with metadata, tags, and collections
//   - Snapshots: Archived copies of bookmarked web pages
//   - Tags: User-defined labels for organizing bookmarks
//   - Collections: Hierarchical bookmark organization
//   - Feeds: RSS/Atom and ActivityPub feed subscriptions
//   - FeedItems: Individual posts from subscribed feeds
//   - Resources: Embedded media and assets from web pages
//   - ActivityPub: Federation followers and interactions
//
// The package uses GORM as the ORM layer and supports both SQLite and PostgreSQL
// databases. All models embed CommonFields which provide ID, timestamps, and
// soft-delete functionality.
//
// Database migrations are handled automatically through the Init function which
// sets up the database connection and runs necessary schema updates.
//
// Example usage:
//
//	err := model.Init(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	user := model.GetUser("username")
//	bookmarks, _, err := model.SearchBookmarks(user.ID, 50, "golang")
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

// DBTypedef represents the type of database being used.
type DBTypedef int

const (
	// Sqlite represents SQLite database type.
	Sqlite DBTypedef = iota
	// Psql represents PostgreSQL database type.
	Psql
)

// ErrDBType is returned when an unknown database type is encountered.
const ErrDBType = utils.StringError("Unknown database type")

// DB is the global database instance.
var DB *gorm.DB

// DBType holds the type of the database being used.
var DBType = Sqlite

// Init initializes the database connection and runs migrations.
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
		return ErrDBType
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

// Database represents the database version tracking table.
type Database struct {
	ID      uint `gorm:"primaryKey"`
	Version uint
}

// CommonFields contains fields common to all models.
type CommonFields struct {
	ID        uint       `gorm:"primary_key" json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}
