package model

import (
	"crypto/rand"
	"fmt"
	"log"
	"time"

	"github.com/asciimoo/omnom/config"

	"gorm.io/gorm"

	//"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
)

var DB *gorm.DB
var debug = false

func Init(c *config.Config) error {
	if c.App.Debug {
		debug = true
		log.Println("Using database", c.DB.Connection)
	}
	var err error
	switch c.DB.Type {
	case "sqlite":
		DB, err = gorm.Open(sqlite.Open(c.DB.Connection), &gorm.Config{})
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
	)
	return nil
}

type User struct {
	gorm.Model
	ID               uint   `gorm:"primaryKey"`
	Username         string `gorm:"unique"`
	Email            string
	LoginToken       string
	SubmissionTokens []Token
	Bookmarks        []Bookmark
	CreatedAt        time.Time
}

type Bookmark struct {
	gorm.Model
	ID        uint `gorm:"primaryKey"`
	URL       string
	Title     string
	Notes     string
	Tags      []Tag `gorm:"foreignKey:ID"`
	Snapshots []Snapshot
	Public    bool
	UserID    uint
	CreatedAt time.Time
}

type Snapshot struct {
	gorm.Model
	ID         uint `gorm:"primaryKey"`
	Site       string
	BookmarkID uint
	CreatedAt  time.Time
}

type Token struct {
	gorm.Model
	ID     uint `gorm:"primaryKey"`
	UserID uint
	Text   string
}

type Tag struct {
	gorm.Model
	ID   uint `gorm:"primaryKey"`
	Text string
}

func GetUser(name string) *User {
	var u User
	err := DB.Where(&User{Username: name}).First(&u).Error
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
	err := DB.Model(&User{}).Select("*").Joins("left join tokens on tokens.user_id = users.id").Where("tokens.text = ?", tok).First(&u).Error
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
