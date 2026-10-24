package miscdao

import (
	"gorm.io/gorm"
	miscmodel "helloworld/pkg/model/misc"
	"mall/kit/codeerror"
	"mall/kit/database"
)

func defaultDB(db *gorm.DB) *gorm.DB {
	if db != nil {
		return db
	}
	return database.GetDb()
}

func AutoMigrate() error {
	if err := database.GetDb().AutoMigrate(
		&miscmodel.Banner{},
		&miscmodel.Announcement{},
	); err != nil {
		return codeerror.OrSystem(err, "AutoMigrate misc tables")
	}
	return nil
}

func NewBannerDao(db *gorm.DB) *BannerDao {
	return &BannerDao{db: defaultDB(db)}
}

func NewAnnouncementDao(db *gorm.DB) *AnnouncementDao {
	return &AnnouncementDao{db: defaultDB(db)}
}
