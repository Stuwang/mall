package order

// OrderType 订单类型
type OrderType int32

const (
	OrderTypeUnspecified OrderType = 0
	OrderTypeProduct     OrderType = 1 // 商品订单
	OrderTypePoint       OrderType = 2 // 积分充值订单
)

// OrderStatus 订单状态
type OrderStatus int32

const (
	OrderStatusUnspecified OrderStatus = 0
	OrderStatusPendingPay  OrderStatus = 1 // 待支付
	OrderStatusPaid        OrderStatus = 2 // 已支付
	OrderStatusCancelled   OrderStatus = 3 // 已取消（用户/超时）
)

// PayType 支付方式
type PayType int32

const (
	PayTypeUnspecified PayType = 0
	PayTypeWechat      PayType = 1 // 微信
	PayTypeAlipay      PayType = 2 // 支付宝
	PayTypePoint       PayType = 3 // 积分抵扣
	PayTypeMock        PayType = 4 // Mock 支付（测试）
)

// CancelReason 取消原因
type CancelReason int32

const (
	CancelReasonUnspecified CancelReason = 0
	CancelReasonUser        CancelReason = 1 // 用户取消
	CancelReasonTimeout     CancelReason = 2 // 支付超时
)
