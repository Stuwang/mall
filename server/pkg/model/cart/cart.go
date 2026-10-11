package cart

// CartItem 购物车项
type CartItem struct {
	ID        int64 `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	UserID    int64 `json:"user_id" gorm:"column:user_id;uniqueIndex:uk_user_product"`
	ProductID int64 `json:"product_id" gorm:"column:product_id;uniqueIndex:uk_user_product"`
	Quantity  int32 `json:"quantity" gorm:"column:quantity"`
	Ct        int64 `json:"ct" gorm:"column:ct"`
	Ut        int64 `json:"ut" gorm:"column:ut"`
}

func (CartItem) TableName() string {
	return "cart_item"
}
