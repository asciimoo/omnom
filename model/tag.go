// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

package model

type Tag struct {
	CommonFields
	Text      string     `gorm:"unique" json:"text"`
	Bookmarks []Bookmark `gorm:"many2many:bookmark_tags;" json:"bookmarks"`
}

type TagCount struct {
	Tag   string
	Count int64
}

func GetFrequentPublicTags(count int) []*TagCount {
	var tags []*TagCount
	DB.Limit(20).Table("tags").Select("tags.text as tag, count(tags.text) as `count`").Joins("join bookmark_tags on bookmark_tags.tag_id == tags.id").Joins("join bookmarks on bookmarks.id == bookmark_tags.bookmark_id").Where("bookmarks.public = true").Group("tags.text").Order("`count` desc, tag asc").Find(&tags)
	return tags
}
func GetOrCreateTag(tag string) Tag {
	var t Tag
	if err := DB.Where("text = ?", tag).First(&t).Error; err != nil {
		t = Tag{
			Text: tag,
		}
		DB.Create(&t)
	}
	return t
}

func GetTagsFromText(s string) ([]Tag, error) {
	var res []Tag
	var err error
	switch DBType {
	case Sqlite:
		err = DB.Raw("WITH cte AS (SELECT ? AS namevar) SELECT tags.* FROM cte, tags WHERE instr(lower(cte.namevar), lower(tags.text)) > 0;", s).Scan(&res).Error
	case Psql:
		err = DB.Raw("WITH cte AS (SELECT ? AS namevar) SELECT tags.* FROM cte, tags WHERE position(lower(tags.text) IN lower(cte.namevar)) > 0;", s).Scan(&res).Error
	default:
		return nil, DBTypeErr
	}
	return res, err
}
