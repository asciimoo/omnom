package model

type Tag struct {
	CommonFields
	Text      string     `gorm:"unique" json:"text"`
	Bookmarks []Bookmark `gorm:"many2many:bookmark_tags;" json:"bookmarks"`
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
