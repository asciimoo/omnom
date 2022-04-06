package model

type Bookmark struct {
	CommonFields
	URL       string     `json:"url"`
	Title     string     `json:"title"`
	Notes     string     `json:"notes"`
	Domain    string     `json:"domain"`
	Favicon   string     `json:"favicon"`
	Tags      []Tag      `gorm:"many2many:bookmark_tags;" json:"tags"`
	Snapshots []Snapshot `json:"snapshots"`
	Public    bool       `json:"public"`
	UserID    uint       `json:"user_id"`
}
