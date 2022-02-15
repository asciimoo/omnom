package model

import (
	"crypto/rand"
	"fmt"
	"log"

	"github.com/asciimoo/omnom/config"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	//"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
)

var DB *gorm.DB
var debug = false

func Init(c *config.Config) error {
	dbCfg := &gorm.Config{}
	if c.App.Debug {
		debug = true
		log.Println("Using database", c.DB.Connection)
		dbCfg.Logger = logger.Default.LogMode(logger.Info)
	}
	var err error
	switch c.DB.Type {
	case "sqlite":
		DB, err = gorm.Open(sqlite.Open(c.DB.Connection), dbCfg)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("Unknown database type")
		//dsn := "host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Europe/Budapest"
		//DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		//if err != nil {
		//	panic(err)
		//}
	}
	DB.AutoMigrate(
		&User{},
		&Bookmark{},
		&Snapshot{},
		&Tag{},
		&Token{},
		&Database{},
		&Resource{},
	)
	migrate()
	return nil
}

type User struct {
	gorm.Model
	ID               uint   `gorm:"primaryKey"`
	Username         string `gorm:"unique"`
	Email            string `gorm:"unique"`
	LoginToken       string
	SubmissionTokens []Token
	Bookmarks        []Bookmark
}

type Bookmark struct {
	gorm.Model
	ID        uint `gorm:"primaryKey"`
	URL       string
	Title     string
	Notes     string
	Domain    string
	Favicon   string
	Tags      []Tag `gorm:"many2many:bookmark_tags;"`
	Snapshots []Snapshot
	Public    bool
	UserID    uint
}

type Snapshot struct {
	gorm.Model
	ID         uint `gorm:"primaryKey"`
	Title      string
	Key        string
	Text       string
	BookmarkID uint
	Size       uint
	Resources  []Resource `gorm:"many2many:snapshot_resources;"`
}

type Token struct {
	gorm.Model
	ID     uint `gorm:"primaryKey"`
	UserID uint
	Text   string
}

type Tag struct {
	gorm.Model
	ID        uint       `gorm:"primaryKey"`
	Text      string     `gorm:"unique"`
	Bookmarks []Bookmark `gorm:"many2many:bookmark_tags;"`
}

type Resource struct {
	gorm.Model
	ID               uint   `gorm:"primaryKey"`
	Key              string `gorm:"unique"`
	MimeType         string
	OriginalFilename string
	Size             uint
	Snapshots        []Snapshot `gorm:"many2many:snapshot_resources;"`
}

type Database struct {
	ID      uint `gorm:"primaryKey"`
	Version uint
}

func GetUser(name string) *User {
	var u User
	err := DB.Where("LOWER(username) == LOWER(?) or LOWER(email) == LOWER(?)", name, name).First(&u).Error
	if err != nil {
		return nil
	}
	return &u
}

func GetUserByLoginToken(tok string) *User {
	var u User
	err := DB.Where(&User{LoginToken: tok}).First(&u).Error
	if err != nil {
		return nil
	}
	return &u
}

func GetUserBySubmissionToken(tok string) *User {
	var u User
	err := DB.Model(&User{}).Select("users.*").Joins("left join tokens on tokens.user_id = users.id").Where("tokens.text = ?", tok).First(&u).Error
	if err != nil {
		return nil
	}

	return &u
}

func CreateUser(username, email string) error {
	if GetUser(username) != nil {
		return fmt.Errorf("User already exists")
	}
	u := &User{
		Username:   username,
		Email:      email,
		LoginToken: GenerateToken(),
		SubmissionTokens: []Token{Token{
			Text: GenerateToken(),
		}},
	}
	return DB.Create(u).Error
}

func GenerateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	tok := fmt.Sprintf("%x", b)
	return tok
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

func GetOrCreateResource(key string, mimeType string, fname string, size uint) Resource {
	var r Resource
	if err := DB.Where("key = ?", key).First(&r).Error; err != nil {
		r = Resource{
			Key:              key,
			MimeType:         mimeType,
			OriginalFilename: fname,
			Size:             size,
		}
		DB.Create(&r)
	}
	return r
}
