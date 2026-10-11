package ordercontroller

import (
	"context"

	apiorder "helloworld/api/order"
	"helloworld/base/codeerror"
	"helloworld/base/logger"
	"helloworld/common/auth"
	ordermodel "helloworld/pkg/model/order"
	"helloworld/pkg/orderbiz"
)

var _ apiorder.OrderHTTPServer = (*Controller)(nil)

type Controller struct{}

func New() *Controller {
	return &Controller{}
}

func (c *Controller) CreateProductOrder(ctx context.Context, req *apiorder.CreateProductOrderRequest) (*apiorder.OrderReply, error) {
	info := auth.MustFromContext(ctx)
	if len(req.Items) == 0 {
		logger.Errorf("CreateProductOrder empty items userID=%d", info.UserID)
		return nil, codeerror.ErrorParam
	}
	lines := make([]orderbiz.ProductLine, 0, len(req.Items))
	for _, it := range req.Items {
		lines = append(lines, orderbiz.ProductLine{ProductID: it.ProductId, Quantity: it.Quantity})
	}
	created, err := orderbiz.CreateProductOrder(info.UserID, lines)
	if err != nil {
		logger.Errorf("CreateProductOrder userID=%d err=%v", info.UserID, err)
		return nil, err
	}
	return &apiorder.OrderReply{Order: toOrderInfo(created)}, nil
}

func (c *Controller) CreateCartOrder(ctx context.Context, req *apiorder.CreateCartOrderRequest) (*apiorder.OrderReply, error) {
	info := auth.MustFromContext(ctx)
	if len(req.CartItemIds) == 0 {
		logger.Errorf("CreateCartOrder empty cart_item_ids userID=%d", info.UserID)
		return nil, codeerror.ErrorParam
	}
	created, err := orderbiz.CreateCartOrder(info.UserID, req.CartItemIds)
	if err != nil {
		logger.Errorf("CreateCartOrder userID=%d err=%v", info.UserID, err)
		return nil, err
	}
	return &apiorder.OrderReply{Order: toOrderInfo(created)}, nil
}

func (c *Controller) CreatePointOrder(ctx context.Context, req *apiorder.CreatePointOrderRequest) (*apiorder.OrderReply, error) {
	info := auth.MustFromContext(ctx)
	created, err := orderbiz.CreatePointOrder(info.UserID, req.Points)
	if err != nil {
		logger.Errorf("CreatePointOrder userID=%d points=%d err=%v", info.UserID, req.Points, err)
		return nil, err
	}
	return &apiorder.OrderReply{Order: toOrderInfo(created)}, nil
}

func (c *Controller) PayOrder(ctx context.Context, req *apiorder.PayOrderRequest) (*apiorder.OrderReply, error) {
	info := auth.MustFromContext(ctx)
	if req.OrderId <= 0 {
		logger.Errorf("PayOrder invalid order_id userID=%d", info.UserID)
		return nil, codeerror.ErrorParam
	}
	created, err := orderbiz.PayOrder(info.UserID, req.OrderId, ordermodel.PayType(req.PayType), req.UsePoints)
	if err != nil {
		logger.Errorf("PayOrder userID=%d orderID=%d payType=%v usePoints=%d err=%v", info.UserID, req.OrderId, req.PayType, req.UsePoints, err)
		return nil, err
	}
	return &apiorder.OrderReply{Order: toOrderInfo(created)}, nil
}

func (c *Controller) MockPaySuccess(ctx context.Context, req *apiorder.MockPaySuccessRequest) (*apiorder.OrderReply, error) {
	info := auth.MustFromContext(ctx)
	if req.OrderId <= 0 {
		logger.Errorf("MockPaySuccess invalid order_id userID=%d", info.UserID)
		return nil, codeerror.ErrorParam
	}
	created, err := orderbiz.MockPaySuccess(info.UserID, req.OrderId)
	if err != nil {
		logger.Errorf("MockPaySuccess userID=%d orderID=%d err=%v", info.UserID, req.OrderId, err)
		return nil, err
	}
	return &apiorder.OrderReply{Order: toOrderInfo(created)}, nil
}

func (c *Controller) CancelOrder(ctx context.Context, req *apiorder.CancelOrderRequest) (*apiorder.OrderReply, error) {
	info := auth.MustFromContext(ctx)
	if req.OrderId <= 0 {
		logger.Errorf("CancelOrder invalid order_id userID=%d", info.UserID)
		return nil, codeerror.ErrorParam
	}
	created, err := orderbiz.CancelOrder(info.UserID, req.OrderId, ordermodel.CancelReasonUser)
	if err != nil {
		logger.Errorf("CancelOrder userID=%d orderID=%d err=%v", info.UserID, req.OrderId, err)
		return nil, err
	}
	return &apiorder.OrderReply{Order: toOrderInfo(created)}, nil
}

func (c *Controller) GetOrder(ctx context.Context, req *apiorder.GetOrderRequest) (*apiorder.OrderReply, error) {
	info := auth.MustFromContext(ctx)
	if req.OrderId <= 0 {
		logger.Errorf("GetOrder invalid order_id userID=%d", info.UserID)
		return nil, codeerror.ErrorParam
	}
	created, err := orderbiz.GetOrder(info.UserID, req.OrderId)
	if err != nil {
		logger.Errorf("GetOrder userID=%d orderID=%d err=%v", info.UserID, req.OrderId, err)
		return nil, err
	}
	return &apiorder.OrderReply{Order: toOrderInfo(created)}, nil
}

func (c *Controller) ListOrders(ctx context.Context, req *apiorder.ListOrdersRequest) (*apiorder.ListOrdersReply, error) {
	info := auth.MustFromContext(ctx)
	list, total, err := orderbiz.ListOrders(info.UserID, ordermodel.OrderStatus(req.Status), int(req.Page), int(req.PageSize))
	if err != nil {
		logger.Errorf("ListOrders userID=%d err=%v", info.UserID, err)
		return nil, err
	}
	out := make([]*apiorder.OrderInfo, 0, len(list))
	for i := range list {
		out = append(out, toOrderInfo(&list[i]))
	}
	return &apiorder.ListOrdersReply{List: out, Total: total}, nil
}

func toOrderInfo(created *orderbiz.CreatedOrder) *apiorder.OrderInfo {
	if created == nil || created.Order == nil {
		return nil
	}
	o := created.Order
	items := make([]*apiorder.OrderItemInfo, 0, len(created.Items))
	for i := range created.Items {
		it := &created.Items[i]
		items = append(items, &apiorder.OrderItemInfo{
			Id:           it.ID,
			ProductId:    it.ProductID,
			ProductName:  it.ProductName,
			ProductPrice: it.ProductPrice,
			Quantity:     it.Quantity,
			Amount:       it.Amount,
		})
	}
	return &apiorder.OrderInfo{
		Id:            o.ID,
		OrderNo:       o.OrderNo,
		UserId:        o.UserID,
		OrderType:     apiorder.OrderType(o.OrderType),
		Status:        apiorder.OrderStatus(o.Status),
		PayType:       apiorder.PayType(o.PayType),
		TotalAmount:   o.TotalAmount,
		CashAmount:    o.CashAmount,
		PointAmount:   o.PointAmount,
		PointsPerYuan: o.PointsPerYuan,
		PayExpireAt:   o.PayExpireAt,
		PaidAt:        o.PaidAt,
		CancelReason:  apiorder.CancelReason(o.CancelReason),
		Items:         items,
		Ct:            o.Ct,
		Ut:            o.Ut,
	}
}
