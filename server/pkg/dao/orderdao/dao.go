package orderdao

import (
	"gorm.io/gorm"
	ordermodel "helloworld/pkg/model/order"
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
	if err := database.GetDb().AutoMigrate(&ordermodel.Order{}, &ordermodel.OrderItem{}); err != nil {
		return codeerror.OrSystem(err, "AutoMigrate order tables")
	}
	return nil
}

func NewOrderDao(db *gorm.DB) *OrderDao {
	return &OrderDao{db: defaultDB(db)}
}

func NewItemDao(db *gorm.DB) *ItemDao {
	return &ItemDao{db: defaultDB(db)}
}
