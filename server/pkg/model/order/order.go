package order

// Order 订单主表
type Order struct {
	ID            int64        `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	OrderNo       string       `json:"order_no" gorm:"column:order_no;uniqueIndex"`
	UserID        int64        `json:"user_id" gorm:"column:user_id;index"`
	OrderType     OrderType    `json:"order_type" gorm:"column:order_type"`
	Status        OrderStatus  `json:"status" gorm:"column:status;index"`
	PayType       PayType      `json:"pay_type" gorm:"column:pay_type"`
	TotalAmount   int64        `json:"total_amount" gorm:"column:total_amount"`     // 订单总价（分）
	CashAmount    int64        `json:"cash_amount" gorm:"column:cash_amount"`       // 现金应付/已付（分）
	PointAmount   int64        `json:"point_amount" gorm:"column:point_amount"`     // 使用积分数 / 充值积分数
	PointsPerYuan int64        `json:"points_per_yuan" gorm:"column:points_per_yuan"` // 汇率快照：1 元 = N 积分
	PayExpireAt   int64        `json:"pay_expire_at" gorm:"column:pay_expire_at"`
	PaidAt        int64        `json:"paid_at" gorm:"column:paid_at"`
	CancelReason  CancelReason `json:"cancel_reason" gorm:"column:cancel_reason"`
	Ct            int64        `json:"ct" gorm:"column:ct"`
	Ut            int64        `json:"ut" gorm:"column:ut"`
}

func (Order) TableName() string {
	return "order"
}

// OrderItem 订单明细
type OrderItem struct {
	ID           int64  `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	OrderID      int64  `json:"order_id" gorm:"column:order_id;index"`
	ProductID    int64  `json:"product_id" gorm:"column:product_id"`
	ProductName  string `json:"product_name" gorm:"column:product_name"`
	ProductPrice int64  `json:"product_price" gorm:"column:product_price"` // 下单时单价快照（分）
	Quantity     int32  `json:"quantity" gorm:"column:quantity"`
	Amount       int64  `json:"amount" gorm:"column:amount"` // 小计（分）
	Ct           int64  `json:"ct" gorm:"column:ct"`
}

func (OrderItem) TableName() string {
	return "order_item"
}
