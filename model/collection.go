// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

package model

// Collection represents a bookmark collection.
type Collection struct {
	CommonFields
	Name      string `gorm:"uniqueIndex:cuid" json:"name"`
	UserID    uint   `gorm:"uniqueIndex:cuid" json:"user_id"`
	ParentID  uint
	Children  []*Collection `gorm:"foreignKey:parent_id"`
	User      User          `json:"-"`
	Bookmarks []Bookmark    `json:"bookmarks"`
}

// GetCollectionByName retrieves a collection by its name.
func GetCollectionByName(uid uint, cname string) *Collection {
	if cname == "" {
		return nil
	}
	var c *Collection
	DB.Where("user_id = ?", uid).Where("name = ?", cname).First(&c)
	return c
}

// GetCollection retrieves a collection by its ID.
func GetCollection(uid uint, cid string) *Collection {
	if cid == "" {
		return nil
	}
	var c *Collection
	DB.Where("user_id = ?", uid).Where("id = ?", cid).First(&c)
	return c
}

// GetCollections retrieves all collections for a user.
func GetCollections(uid uint) []*Collection {
	var cols []*Collection
	DB.Model(&Collection{}).Where("user_id = ?", uid).Order("name asc").Find(&cols)
	return cols
}

// GetCollectionTree retrieves collections organized as a tree structure.
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
