// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

package model

type APFollower struct {
	CommonFields
	Name        string `gorm:"uniqueIndex:uidx" json:"name"`
	FollowedURL string `gorm:"uniqueIndex:uidx" json:"followed_url"`
}

func CreateAPFollower(name, url string) error {
	f := APFollower{
		Name:        name,
		FollowedURL: url,
	}
	return DB.Create(&f).Error
}
