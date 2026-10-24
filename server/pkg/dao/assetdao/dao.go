package assetdao

import (
	"gorm.io/gorm"
	assetmodel "helloworld/pkg/model/asset"
	"mall/kit/codeerror"
	"helloworld/common/database"
)

func defaultDB(db *gorm.DB) *gorm.DB {
	if db != nil {
		return db
	}
	return database.GetDb()
}

func AutoMigrate() error {
	if err := database.GetDb().AutoMigrate(&assetmodel.UserAsset{}, &assetmodel.UserAssetHistory{}); err != nil {
		return codeerror.OrSystem(err, "AutoMigrate asset tables")
	}
	return nil
}

func NewAssetDao(db *gorm.DB) *AssetDao {
	return &AssetDao{db: defaultDB(db)}
}

func NewHistoryDao(db *gorm.DB) *HistoryDao {
	return &HistoryDao{db: defaultDB(db)}
}
