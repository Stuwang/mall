package database

import (
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"mall/kit/logger"
)

var db *gorm.DB

func GetDb() *gorm.DB {
	return db
}

func InitDb(filename string) {
	if filename == "" {
		filename = "im.db"
	}
	var err error
	db, err = gorm.Open(sqlite.Open(filename), &gorm.Config{})
	if err != nil {
		logger.Panicf("error open db %v", err)
		return
	}
}

func Close() {
	if db == nil {
		return
	}
	sqlDb, err := db.DB()
	if err != nil {
		logger.Errorf("error %v", err)
		return
	}
	err = sqlDb.Close()
	if err != nil {
		logger.Errorf("error %v", err)
	}
	db = nil
}
