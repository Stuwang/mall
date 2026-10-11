package cartdao

import (
	"gorm.io/gorm"
	"helloworld/base/codeerror"
	"helloworld/common/database"
	cartmodel "helloworld/pkg/model/cart"
)

func defaultDB(db *gorm.DB) *gorm.DB {
	if db != nil {
		return db
	}
	return database.GetDb()
}

func AutoMigrate() error {
	if err := database.GetDb().AutoMigrate(&cartmodel.CartItem{}); err != nil {
		return codeerror.OrSystem(err, "AutoMigrate cart tables")
	}
	return nil
}

func NewCartDao(db *gorm.DB) *CartDao {
	return &CartDao{db: defaultDB(db)}
}
