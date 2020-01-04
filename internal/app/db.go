package app

import (
	"path/filepath"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

type GormID = uint

func openDB(fid string) *gorm.DB {
	dbf := filepath.Join(cmdrDir(fid), cmdrDb)
	res, err := gorm.Open("sqlite3", dbf)
	if err != nil {
		log.Panice(err)
	}
	return res.AutoMigrate(&Story{}, &Hint{})
}

type Story struct {
	ID     GormID
	Master string    `gorm:"not null"`
	Title  string    `gorm:"not null"`
	Joined time.Time `gorm:"not null"`
	Over   *time.Time
}

type Hint struct {
	ID      GormID
	StoryID GormID
	Story   *Story
	Found   time.Time `gorm:"not null"`
	Written time.Time
	Text    string
}
