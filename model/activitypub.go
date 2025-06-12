// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

package model

type APFollower struct {
	CommonFields
	UserID   uint   `gorm:"uniqueIndex:apuidx" json:"uid"`
	Follower string `gorm:"uniqueIndex:apuidx" json:"follower"`
}

func CreateAPFollower(uid uint, follower string) error {
	f := APFollower{
		UserID:   uid,
		Follower: follower,
	}
	return DB.Create(&f).Error
}
