package product

const (
	StatusOn  int32 = 1 // 上架 / 启用
	StatusOff int32 = 2 // 下架 / 停用
)

// Category 商品分类
type Category struct {
	ID     int64  `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Name   string `json:"name" gorm:"column:name"`
	Sort   int32  `json:"sort" gorm:"column:sort"`
	Status int32  `json:"status" gorm:"column:status;index"`
	Ct     int64  `json:"ct" gorm:"column:ct"`
	Ut     int64  `json:"ut" gorm:"column:ut"`
}

func (Category) TableName() string {
	return "product_category"
}

// Product 商品
type Product struct {
	ID          int64  `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	CategoryID  int64  `json:"category_id" gorm:"column:category_id;index"`
	Name        string `json:"name" gorm:"column:name"`
	Description string `json:"description" gorm:"column:description"`
	Price       int64  `json:"price" gorm:"column:price"` // 分
	Stock       int32  `json:"stock" gorm:"column:stock"`
	Status      int32  `json:"status" gorm:"column:status;index"`
	Cover       string `json:"cover" gorm:"column:cover"`
	Ct          int64  `json:"ct" gorm:"column:ct"`
	Ut          int64  `json:"ut" gorm:"column:ut"`
}

func (Product) TableName() string {
	return "product"
}
