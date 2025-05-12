// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

package model

import (
	"github.com/asciimoo/omnom/storage"

	"github.com/rs/zerolog/log"
)

var migrationFunctions = []func() error{
	addSnapshotSizes,             // db version 1
	removeUnusedAPFollowerFields, // db version 2
}

func migrate() error {
	var dbVer int64
	err := DB.Model(&Database{}).
		Select("version").
		First(&dbVer).Error
	if err != nil {
		DB.Save(&Database{Version: 0})
	}
	migCount := 0
	for i, m := range migrationFunctions {
		if int64(i) >= dbVer {
			log.Info().Msgf("Migrating DB to version %d", i+1)
			err := m()
			if err != nil {
				return err
			}
			dbVer = int64(i) + 1
			DB.Model(&Database{}).Where("id = 1").Update("version", dbVer)
			migCount += 1
		}
	}
	log.Debug().Int("Migrations performed", migCount).Msg("DB migrations completed")
	return nil
}

func addSnapshotSizes() error {
	log.Debug().Msg("Updating snapshot sizes")
	var snapshotsWithoutSize []string
	DB.Model(&Snapshot{}).Distinct().
		Select("key").
		Where("size is null and key is not null").
		Find(&snapshotsWithoutSize)
	for _, s := range snapshotsWithoutSize {
		if s == "" {
			continue
		}
		size := storage.GetSnapshotSize(s)
		DB.Model(&Snapshot{}).Where("key = ?", s).Update("size", size)
	}
	return nil
}

func removeUnusedAPFollowerFields() error {
	log.Debug().Msg("Removing unused ActivityPub follower fields")
	for _, col := range []string{"filter", "name"} {
		err := DB.Migrator().DropColumn(&APFollower{}, col)
		if err != nil {
			log.Debug().Str("column", col).Msg("Failed to delete")
		}
	}
	return nil
}
