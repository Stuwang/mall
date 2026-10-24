package orderbiz

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"helloworld/pkg/bizerror"
	"helloworld/pkg/dao/assetdao"
	"helloworld/pkg/dao/cartdao"
	"helloworld/pkg/dao/orderdao"
	"helloworld/pkg/dao/productdao"
	assetmodel "helloworld/pkg/model/asset"
	ordermodel "helloworld/pkg/model/order"
	productmodel "helloworld/pkg/model/product"
	"helloworld/pkg/pointrate"
	"mall/kit/codeerror"
	"helloworld/common/database"
	"mall/kit/logger"
)

var (
	payTimeoutSec int64 = 15 * 60
	pointsPerYuan int64 = 100 // 1 元 = 100 积分（可配置，下单时快照）
)

// Init 初始化订单业务配置
func Init(timeoutSec int, pointsPerYuanCfg int64) {
	if timeoutSec > 0 {
		payTimeoutSec = int64(timeoutSec)
	}
	if pointsPerYuanCfg > 0 {
		pointsPerYuan = pointsPerYuanCfg
	}
}

func currentPointsPerYuan() int64 {
	return pointsPerYuan
}

type ProductLine struct {
	ProductID int64
	Quantity  int32
}

type CreatedOrder struct {
	Order *ordermodel.Order
	Items []ordermodel.OrderItem
}

func genOrderNo(userID int64) string {
	return fmt.Sprintf("O%d%d", time.Now().UnixNano(), userID%10000)
}

func CreateProductOrder(userID int64, lines []ProductLine) (*CreatedOrder, error) {
	if userID <= 0 || len(lines) == 0 {
		logger.Errorf("CreateProductOrder invalid userID=%d lines=%d", userID, len(lines))
		return nil, codeerror.ErrorParam
	}
	return createProductOrderInternal(userID, lines, nil)
}

func CreateCartOrder(userID int64, cartItemIDs []int64) (*CreatedOrder, error) {
	if userID <= 0 || len(cartItemIDs) == 0 {
		logger.Errorf("CreateCartOrder invalid userID=%d cartItemIDs=%d", userID, len(cartItemIDs))
		return nil, codeerror.ErrorParam
	}
	cartItems, err := cartdao.NewCartDao(nil).ListByIDs(userID, cartItemIDs)
	if err != nil {
		logger.Errorf("CreateCartOrder ListByIDs userID=%d err=%v", userID, err)
		return nil, err
	}
	if len(cartItems) != len(cartItemIDs) {
		logger.Errorf("CreateCartOrder cart mismatch userID=%d want=%d got=%d", userID, len(cartItemIDs), len(cartItems))
		return nil, codeerror.ErrorParam.Errorf("cart items invalid")
	}
	lines := make([]ProductLine, 0, len(cartItems))
	ids := make([]int64, 0, len(cartItems))
	for i := range cartItems {
		lines = append(lines, ProductLine{ProductID: cartItems[i].ProductID, Quantity: cartItems[i].Quantity})
		ids = append(ids, cartItems[i].ID)
	}
	return createProductOrderInternal(userID, lines, ids)
}

func createProductOrderInternal(userID int64, lines []ProductLine, removeCartIDs []int64) (*CreatedOrder, error) {
	var result *CreatedOrder
	rate := currentPointsPerYuan()
	err := database.GetDb().Transaction(func(tx *gorm.DB) error {
		productDao := productdao.NewProductDao(tx)
		orderDao := orderdao.NewOrderDao(tx)
		itemDao := orderdao.NewItemDao(tx)
		cartDao := cartdao.NewCartDao(tx)

		now := time.Now().Unix()
		items := make([]ordermodel.OrderItem, 0, len(lines))
		var total int64
		for _, line := range lines {
			if line.ProductID <= 0 || line.Quantity <= 0 {
				return codeerror.ErrorParam
			}
			p, err := productDao.GetByID(line.ProductID)
			if err != nil {
				return err
			}
			if p.Status != productmodel.StatusOn {
				return codeerror.ErrorParam.Errorf("product not available")
			}
			if err := productDao.DecStock(p.ID, line.Quantity); err != nil {
				return err
			}
			amount := p.Price * int64(line.Quantity)
			total += amount
			items = append(items, ordermodel.OrderItem{
				ProductID:    p.ID,
				ProductName:  p.Name,
				ProductPrice: p.Price,
				Quantity:     line.Quantity,
				Amount:       amount,
				Ct:           now,
			})
		}

		o := &ordermodel.Order{
			OrderNo:       genOrderNo(userID),
			UserID:        userID,
			OrderType:     ordermodel.OrderTypeProduct,
			Status:        ordermodel.OrderStatusPendingPay,
			TotalAmount:   total,
			CashAmount:    total, // 支付前默认全额现金，Pay 时按积分重算
			PointAmount:   0,
			PointsPerYuan: rate,
			PayExpireAt:   now + payTimeoutSec,
			Ct:            now,
			Ut:            now,
		}
		if err := orderDao.Create(o); err != nil {
			return err
		}
		for i := range items {
			items[i].OrderID = o.ID
		}
		if err := itemDao.CreateBatch(items); err != nil {
			return err
		}
		if len(removeCartIDs) > 0 {
			if err := cartDao.SoftDeleteByIDs(userID, removeCartIDs); err != nil {
				return err
			}
		}
		result = &CreatedOrder{Order: o, Items: items}
		return nil
	})
	if err != nil {
		logger.Errorf("createProductOrderInternal userID=%d err=%v", userID, err)
		return nil, codeerror.OrSystem(err, "createProductOrderInternal userID=%d", userID)
	}
	return result, nil
}

// CreatePointOrder 购买积分；金额按当前汇率计算并快照
func CreatePointOrder(userID, points int64) (*CreatedOrder, error) {
	if userID <= 0 || points <= 0 {
		logger.Errorf("CreatePointOrder invalid userID=%d points=%d", userID, points)
		return nil, codeerror.ErrorParam
	}
	rate := currentPointsPerYuan()
	amount := pointrate.PointsToCents(points, rate)
	if amount <= 0 {
		logger.Errorf("CreatePointOrder amount<=0 userID=%d points=%d rate=%d", userID, points, rate)
		return nil, codeerror.ErrorParam
	}
	now := time.Now().Unix()
	o := &ordermodel.Order{
		OrderNo:       genOrderNo(userID),
		UserID:        userID,
		OrderType:     ordermodel.OrderTypePoint,
		Status:        ordermodel.OrderStatusPendingPay,
		TotalAmount:   amount,
		CashAmount:    amount,
		PointAmount:   points, // 待到账积分
		PointsPerYuan: rate,
		PayExpireAt:   now + payTimeoutSec,
		Ct:            now,
		Ut:            now,
	}
	if err := orderdao.NewOrderDao(nil).Create(o); err != nil {
		logger.Errorf("CreatePointOrder Create userID=%d err=%v", userID, err)
		return nil, err
	}
	return &CreatedOrder{Order: o, Items: nil}, nil
}

// PayOrder 选择支付：usePoints 为抵扣积分；payType 为现金渠道（全额积分时传 POINT）
func PayOrder(userID, orderID int64, payType ordermodel.PayType, usePoints int64) (*CreatedOrder, error) {
	if userID <= 0 || orderID <= 0 {
		logger.Errorf("PayOrder invalid userID=%d orderID=%d", userID, orderID)
		return nil, codeerror.ErrorParam
	}
	if usePoints < 0 {
		logger.Errorf("PayOrder invalid usePoints=%d orderID=%d", usePoints, orderID)
		return nil, codeerror.ErrorParam
	}

	var result *CreatedOrder
	err := database.GetDb().Transaction(func(tx *gorm.DB) error {
		orderDao := orderdao.NewOrderDao(tx)
		itemDao := orderdao.NewItemDao(tx)

		o, err := orderDao.GetByID(orderID, userID)
		if err != nil {
			return err
		}
		if err := ensurePendingPayable(o); err != nil {
			return err
		}

		items, err := itemDao.ListByOrderID(o.ID)
		if err != nil {
			return err
		}

		switch o.OrderType {
		case ordermodel.OrderTypePoint:
			// 积分充值只能货币支付
			if usePoints > 0 || payType == ordermodel.PayTypePoint {
				return codeerror.ErrorParam.Errorf("point order cannot use point pay")
			}
			if payType != ordermodel.PayTypeWechat && payType != ordermodel.PayTypeAlipay && payType != ordermodel.PayTypeMock {
				return codeerror.ErrorParam
			}
			o.PayType = payType
			o.CashAmount = o.TotalAmount
			if err := orderDao.Update(o, "pay_type", "cash_amount"); err != nil {
				return err
			}
		case ordermodel.OrderTypeProduct:
			rate := o.PointsPerYuan
			if rate <= 0 {
				rate = currentPointsPerYuan()
				o.PointsPerYuan = rate
			}
			if usePoints > 0 {
				a, err := assetdao.NewAssetDao(tx).FindByUserAndType(userID, assetmodel.TypePoint)
				if err != nil {
					return err
				}
				balance := int64(0)
				if a != nil {
					balance = a.Balance
				}
				if usePoints > balance {
					logger.Errorf("PayOrder points insufficient orderID=%d usePoints=%d balance=%d", orderID, usePoints, balance)
					return bizerror.ErrorAssetInsufficient
				}
			}
			pointCents := pointrate.PointsToCents(usePoints, rate)
			if pointCents > o.TotalAmount {
				return codeerror.ErrorParam.Errorf("points exceed order amount")
			}
			cashCents := o.TotalAmount - pointCents
			o.PointAmount = usePoints
			o.CashAmount = cashCents

			if cashCents == 0 {
				// 全额积分
				o.PayType = ordermodel.PayTypePoint
				if err := orderDao.Update(o, "pay_type", "point_amount", "cash_amount", "points_per_yuan"); err != nil {
					return err
				}
				if err := markPaidTx(tx, o); err != nil {
					return err
				}
			} else {
				if payType != ordermodel.PayTypeWechat && payType != ordermodel.PayTypeAlipay && payType != ordermodel.PayTypeMock {
					return codeerror.ErrorParam.Errorf("cash pay type required")
				}
				o.PayType = payType
				if err := orderDao.Update(o, "pay_type", "point_amount", "cash_amount", "points_per_yuan"); err != nil {
					return err
				}
			}
		default:
			return codeerror.ErrorParam
		}

		result = &CreatedOrder{Order: o, Items: items}
		return nil
	})
	if err != nil {
		logger.Errorf("PayOrder userID=%d orderID=%d payType=%d usePoints=%d err=%v", userID, orderID, payType, usePoints, err)
		return nil, codeerror.OrSystem(err, "PayOrder orderID=%d", orderID)
	}
	return result, nil
}

func MockPaySuccess(userID, orderID int64) (*CreatedOrder, error) {
	if userID <= 0 || orderID <= 0 {
		logger.Errorf("MockPaySuccess invalid userID=%d orderID=%d", userID, orderID)
		return nil, codeerror.ErrorParam
	}
	var result *CreatedOrder
	err := database.GetDb().Transaction(func(tx *gorm.DB) error {
		orderDao := orderdao.NewOrderDao(tx)
		itemDao := orderdao.NewItemDao(tx)
		o, err := orderDao.GetByID(orderID, userID)
		if err != nil {
			return err
		}
		if err := ensurePendingPayable(o); err != nil {
			return err
		}
		if o.PayType != ordermodel.PayTypeMock {
			return codeerror.ErrorParam.Errorf("pay type is not mock")
		}
		items, err := itemDao.ListByOrderID(o.ID)
		if err != nil {
			return err
		}
		if err := markPaidTx(tx, o); err != nil {
			return err
		}
		result = &CreatedOrder{Order: o, Items: items}
		return nil
	})
	if err != nil {
		logger.Errorf("MockPaySuccess userID=%d orderID=%d err=%v", userID, orderID, err)
		return nil, codeerror.OrSystem(err, "MockPaySuccess orderID=%d", orderID)
	}
	return result, nil
}

func CancelOrder(userID, orderID int64, reason ordermodel.CancelReason) (*CreatedOrder, error) {
	if userID <= 0 || orderID <= 0 {
		logger.Errorf("CancelOrder invalid userID=%d orderID=%d", userID, orderID)
		return nil, codeerror.ErrorParam
	}
	if reason == ordermodel.CancelReasonUnspecified {
		reason = ordermodel.CancelReasonUser
	}
	var result *CreatedOrder
	err := database.GetDb().Transaction(func(tx *gorm.DB) error {
		created, err := cancelOrderTx(tx, userID, orderID, reason)
		if err != nil {
			return err
		}
		result = created
		return nil
	})
	if err != nil {
		logger.Errorf("CancelOrder userID=%d orderID=%d err=%v", userID, orderID, err)
		return nil, codeerror.OrSystem(err, "CancelOrder orderID=%d", orderID)
	}
	return result, nil
}

func CancelExpiredOrders() {
	now := time.Now().Unix()
	list, err := orderdao.NewOrderDao(nil).ListExpiredPending(100, now)
	if err != nil {
		logger.Errorf("CancelExpiredOrders ListExpiredPending err=%v", err)
		return
	}
	for i := range list {
		o := list[i]
		_, err := CancelOrder(o.UserID, o.ID, ordermodel.CancelReasonTimeout)
		if err != nil {
			logger.Errorf("CancelExpiredOrders cancel orderID=%d err=%v", o.ID, err)
		}
	}
}

func GetOrder(userID, orderID int64) (*CreatedOrder, error) {
	o, err := orderdao.NewOrderDao(nil).GetByID(orderID, userID)
	if err != nil {
		logger.Errorf("GetOrder orderID=%d userID=%d err=%v", orderID, userID, err)
		return nil, err
	}
	_ = maybeAutoCancel(o)
	o, err = orderdao.NewOrderDao(nil).GetByID(orderID, userID)
	if err != nil {
		logger.Errorf("GetOrder reload orderID=%d err=%v", orderID, err)
		return nil, err
	}
	items, err := orderdao.NewItemDao(nil).ListByOrderID(orderID)
	if err != nil {
		logger.Errorf("GetOrder ListByOrderID orderID=%d err=%v", orderID, err)
		return nil, err
	}
	return &CreatedOrder{Order: o, Items: items}, nil
}

func ListOrders(userID int64, status ordermodel.OrderStatus, page, pageSize int) ([]CreatedOrder, int64, error) {
	return listOrders(orderdao.ListQuery{
		UserID:      userID,
		Status:      status,
		Page:        page,
		PageSize:    pageSize,
		RequireUser: true,
	})
}

// AdminListOrders 管理侧多条件订单查询
func AdminListOrders(q orderdao.ListQuery) ([]CreatedOrder, int64, error) {
	q.RequireUser = false
	return listOrders(q)
}

func listOrders(q orderdao.ListQuery) ([]CreatedOrder, int64, error) {
	list, total, err := orderdao.NewOrderDao(nil).List(q)
	if err != nil {
		logger.Errorf("listOrders err=%v", err)
		return nil, 0, err
	}
	for i := range list {
		_ = maybeAutoCancel(&list[i])
	}
	list, total, err = orderdao.NewOrderDao(nil).List(q)
	if err != nil {
		logger.Errorf("listOrders reload err=%v", err)
		return nil, 0, err
	}
	ids := make([]int64, 0, len(list))
	for i := range list {
		ids = append(ids, list[i].ID)
	}
	itemMap, err := orderdao.NewItemDao(nil).ListByOrderIDs(ids)
	if err != nil {
		logger.Errorf("listOrders ListByOrderIDs err=%v", err)
		return nil, 0, err
	}
	out := make([]CreatedOrder, 0, len(list))
	for i := range list {
		out = append(out, CreatedOrder{Order: &list[i], Items: itemMap[list[i].ID]})
	}
	return out, total, nil
}

func maybeAutoCancel(o *ordermodel.Order) error {
	if o == nil || o.Status != ordermodel.OrderStatusPendingPay {
		return nil
	}
	if o.PayExpireAt > 0 && o.PayExpireAt < time.Now().Unix() {
		_, err := CancelOrder(o.UserID, o.ID, ordermodel.CancelReasonTimeout)
		return err
	}
	return nil
}

func ensurePendingPayable(o *ordermodel.Order) error {
	if o.Status != ordermodel.OrderStatusPendingPay {
		return bizerror.ErrorOrderStatus
	}
	if o.PayExpireAt > 0 && o.PayExpireAt < time.Now().Unix() {
		return bizerror.ErrorPayExpired
	}
	return nil
}

func markPaidTx(tx *gorm.DB, o *ordermodel.Order) error {
	orderDao := orderdao.NewOrderDao(tx)
	now := time.Now().Unix()

	switch o.OrderType {
	case ordermodel.OrderTypeProduct:
		// 混合/全额积分：支付完成时扣积分
		if o.PointAmount > 0 {
			if _, err := assetdao.ChangeTx(tx, assetdao.ChangeReq{
				UserID:       o.UserID,
				Type:         assetmodel.TypePoint,
				Change:       -o.PointAmount,
				IdempotentID: "order_pay_point_" + o.OrderNo,
				BizType:      "order_pay_point",
				BizID:        o.OrderNo,
				Remark:       "订单积分抵扣",
			}); err != nil {
				return err
			}
		}
	case ordermodel.OrderTypePoint:
		if _, err := assetdao.ChangeTx(tx, assetdao.ChangeReq{
			UserID:       o.UserID,
			Type:         assetmodel.TypePoint,
			Change:       o.PointAmount,
			IdempotentID: "order_point_" + o.OrderNo,
			BizType:      "order_point_recharge",
			BizID:        o.OrderNo,
			Remark:       "购买积分到账",
		}); err != nil {
			return err
		}
	default:
		return codeerror.ErrorParam
	}

	o.Status = ordermodel.OrderStatusPaid
	o.PaidAt = now
	if err := orderDao.Update(o, "status", "paid_at", "point_amount", "cash_amount", "pay_type", "points_per_yuan"); err != nil {
		return err
	}
	return nil
}

func cancelOrderTx(tx *gorm.DB, userID, orderID int64, reason ordermodel.CancelReason) (*CreatedOrder, error) {
	orderDao := orderdao.NewOrderDao(tx)
	itemDao := orderdao.NewItemDao(tx)
	productDao := productdao.NewProductDao(tx)

	o, err := orderDao.GetByID(orderID, userID)
	if err != nil {
		return nil, err
	}
	if o.Status != ordermodel.OrderStatusPendingPay {
		return nil, bizerror.ErrorOrderStatus
	}
	items, err := itemDao.ListByOrderID(o.ID)
	if err != nil {
		return nil, err
	}
	if o.OrderType == ordermodel.OrderTypeProduct {
		for i := range items {
			if err := productDao.IncStock(items[i].ProductID, items[i].Quantity); err != nil {
				return nil, err
			}
		}
	}
	o.Status = ordermodel.OrderStatusCancelled
	o.CancelReason = reason
	if err := orderDao.Update(o, "status", "cancel_reason"); err != nil {
		return nil, err
	}
	return &CreatedOrder{Order: o, Items: items}, nil
}
