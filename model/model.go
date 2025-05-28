// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

package model

import (
	"fmt"
	"time"

	"github.com/asciimoo/omnom/config"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	//"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
)

var DB *gorm.DB

func Init(c *config.Config) error {
	dbCfg := &gorm.Config{}
	if c.App.DebugSQL {
		dbCfg.Logger = logger.Default.LogMode(logger.Info)
	} else {
		dbCfg.Logger = logger.Default.LogMode(logger.Silent)
	}
	var err error
	switch c.DB.Type {
	case "sqlite":
		DB, err = gorm.Open(sqlite.Open(c.DB.Connection), dbCfg)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown database type")
		//dsn := "host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Europe/Budapest"
		//DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		//if err != nil {
		//	panic(err)
		//}
	}
	err = migrate()
	if err != nil {
		return fmt.Errorf("custom migration of database '%s' has failed: %w", c.DB.Connection, err)
	}
	err = DB.AutoMigrate(
		&User{},
		&Bookmark{},
		&Snapshot{},
		&Tag{},
		&Token{},
		&Database{},
		&Resource{},
		&APFollower{},
	)
	if err != nil {
		return fmt.Errorf("auto migration of database '%s' has failed: %w", c.DB.Connection, err)
	}
	return nil
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
