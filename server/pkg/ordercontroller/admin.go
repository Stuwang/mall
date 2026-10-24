package ordercontroller

import (
	"context"

	apiorder "helloworld/api/order"
	"helloworld/pkg/dao/orderdao"
	ordermodel "helloworld/pkg/model/order"
	"helloworld/pkg/orderbiz"
	"mall/kit/auth"
	"mall/kit/codeerror"
	"mall/kit/logger"
)

func (c *Controller) AdminListOrders(ctx context.Context, req *apiorder.AdminListOrdersRequest) (*apiorder.ListOrdersReply, error) {
	if _, err := auth.RequireAdmin(ctx); err != nil {
		logger.Errorf("AdminListOrders RequireAdmin err=%v", err)
		return nil, err
	}
	if req.StartCt > 0 && req.EndCt > 0 && req.StartCt > req.EndCt {
		logger.Errorf("AdminListOrders invalid date range start=%d end=%d", req.StartCt, req.EndCt)
		return nil, codeerror.ErrorParam
	}
	list, total, err := orderbiz.AdminListOrders(orderdao.ListQuery{
		UserID:    req.UserId,
		Status:    ordermodel.OrderStatus(req.Status),
		StartCt:   req.StartCt,
		EndCt:     req.EndCt,
		ProductID: req.ProductId,
		Page:      int(req.Page),
		PageSize:  int(req.PageSize),
	})
	if err != nil {
		logger.Errorf("AdminListOrders err=%v", err)
		return nil, err
	}
	out := make([]*apiorder.OrderInfo, 0, len(list))
	for i := range list {
		out = append(out, toOrderInfo(&list[i]))
	}
	return &apiorder.ListOrdersReply{List: out, Total: total}, nil
}

func (c *Controller) AdminGetOrder(ctx context.Context, req *apiorder.AdminGetOrderRequest) (*apiorder.OrderReply, error) {
	if _, err := auth.RequireAdmin(ctx); err != nil {
		logger.Errorf("AdminGetOrder RequireAdmin err=%v", err)
		return nil, err
	}
	if req.OrderId <= 0 {
		logger.Errorf("AdminGetOrder invalid order_id=%d", req.OrderId)
		return nil, codeerror.ErrorParam
	}
	// userID=0：不按归属用户过滤
	created, err := orderbiz.GetOrder(0, req.OrderId)
	if err != nil {
		logger.Errorf("AdminGetOrder orderID=%d err=%v", req.OrderId, err)
		return nil, err
	}
	return &apiorder.OrderReply{Order: toOrderInfo(created)}, nil
}
