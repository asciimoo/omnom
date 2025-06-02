// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

package model

type Collection struct {
	CommonFields
	Name      string `gorm:"uniqueIndex:cuid" json:"name"`
	UserID    uint   `gorm:"uniqueIndex:cuid" json:"user_id"`
	ParentID  uint
	Children  []*Collection `gorm:"foreignKey:parent_id"`
	User      User          `json:"-"`
	Bookmarks []Bookmark    `json:"bookmarks"`
}

func GetCollectionByName(uid uint, cname string) *Collection {
	if cname == "" {
		return nil
	}
	var c *Collection
	DB.Where("user_id = ?", uid).Where("name = ?", cname).First(&c)
	return c
}

func GetCollection(uid uint, cid string) *Collection {
	if cid == "" {
		return nil
	}
	var c *Collection
	DB.Where("user_id = ?", uid).Where("id = ?", cid).First(&c)
	return c
}

func GetCollections(uid uint) []*Collection {
	var cols []*Collection
	DB.Model(&Collection{}).Where("user_id = ?", uid).Order("name asc").Find(&cols)
	return cols
}

func GetCollectionTree(uid uint) []*Collection {
	cols := GetCollections(uid)
	res := make([]*Collection, 0, len(cols))
	for _, c := range cols {
		if c.ParentID < 1 {
			res = append(res, c)
		} else {
			for _, cc := range cols {
				if cc.ID == c.ParentID {
					cc.Children = append(cc.Children, c)
					break
				}
			}
		}
	}
	return res
}
