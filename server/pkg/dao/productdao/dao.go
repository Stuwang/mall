package productdao

import (
	"gorm.io/gorm"
	"helloworld/base/codeerror"
	"helloworld/common/database"
	productmodel "helloworld/pkg/model/product"
)

func defaultDB(db *gorm.DB) *gorm.DB {
	if db != nil {
		return db
	}
	return database.GetDb()
}

func AutoMigrate() error {
	if err := database.GetDb().AutoMigrate(
		&productmodel.Category{},
		&productmodel.Product{},
		&productmodel.ProductImage{},
	); err != nil {
		return codeerror.OrSystem(err, "AutoMigrate product tables")
	}
	return nil
}

func normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

// NewCategoryDao 分类 Dao；db 为 nil 时使用全局连接，跨表事务请传入 tx
func NewCategoryDao(db *gorm.DB) *CategoryDao {
	return &CategoryDao{db: defaultDB(db)}
}

// NewProductDao 商品 Dao；db 为 nil 时使用全局连接，跨表事务请传入 tx
func NewProductDao(db *gorm.DB) *ProductDao {
	return &ProductDao{db: defaultDB(db)}
}

// NewImageDao 商品图片 Dao；db 为 nil 时使用全局连接，跨表事务请传入 tx
func NewImageDao(db *gorm.DB) *ImageDao {
	return &ImageDao{db: defaultDB(db)}
}
