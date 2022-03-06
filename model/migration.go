package model

import (
	"log"

	"github.com/asciimoo/omnom/storage"
)

var migrationFunctions = []func() error{
	addSnapshotSizes, // db version 1
}

func migrate() {
	log.Println("Checking DB migrations")
	var dbVer int64
	err := DB.Model(&Database{}).
		Select("version").
		First(&dbVer).Error
	if err != nil {
		DB.Save(&Database{Version: 0})
	}
	for i, m := range migrationFunctions {
		if int64(i) >= dbVer {
			log.Println("Migrating DB to version ", i+1)
			err := m()
			if err != nil {
				panic(err)
			}
			dbVer = int64(i) + 1
			DB.Model(&Database{}).Where("id = 1").Update("version", dbVer)
		}
	}
}

func addSnapshotSizes() error {
	log.Println("updating snapshot sizes")
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
