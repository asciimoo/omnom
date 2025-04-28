// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

package model

type APFollower struct {
	CommonFields
	Name   string `gorm:"uniqueIndex:uidx" json:"name"`
	Filter string `gorm:"uniqueIndex:uidx" json:"filter"`
}

func CreateAPFollower(name, filter string) error {
	f := APFollower{
		Name:   name,
		Filter: filter,
	}
	return DB.Create(&f).Error
}
