package cartcontroller

import (
	"context"

	apicart "helloworld/api/cart"
	"helloworld/base/codeerror"
	"helloworld/base/logger"
	"helloworld/common/auth"
	"helloworld/pkg/dao/cartdao"
	"helloworld/pkg/dao/productdao"
	cartmodel "helloworld/pkg/model/cart"
	"helloworld/pkg/model/common"
	productmodel "helloworld/pkg/model/product"
)

var _ apicart.CartHTTPServer = (*Controller)(nil)

type Controller struct{}

func New() *Controller {
	return &Controller{}
}

func (c *Controller) AddToCart(ctx context.Context, req *apicart.AddToCartRequest) (*apicart.CartItemReply, error) {
	info := auth.MustFromContext(ctx)
	qty := req.Quantity
	if qty <= 0 {
		qty = 1
	}
	if req.ProductId <= 0 {
		logger.Errorf("AddToCart invalid product_id=%d userID=%d", req.ProductId, info.UserID)
		return nil, codeerror.ErrorParam
	}

	product, err := productdao.NewProductDao(nil).GetByID(req.ProductId)
	if err != nil {
		logger.Errorf("AddToCart GetProduct productID=%d err=%v", req.ProductId, err)
		return nil, err
	}
	if product.Status != productmodel.StatusOn {
		logger.Errorf("AddToCart product not on shelf productID=%d status=%d", req.ProductId, product.Status)
		return nil, codeerror.ErrorParam.Errorf("product not available")
	}

	cartDao := cartdao.NewCartDao(nil)
	item, err := cartDao.FindByUserProduct(info.UserID, req.ProductId)
	if err != nil {
		logger.Errorf("AddToCart FindByUserProduct userID=%d productID=%d err=%v", info.UserID, req.ProductId, err)
		return nil, err
	}

	var targetQty int32
	if item == nil {
		anyItem, err := cartDao.FindAnyByUserProduct(info.UserID, req.ProductId)
		if err != nil {
			logger.Errorf("AddToCart FindAnyByUserProduct userID=%d productID=%d err=%v", info.UserID, req.ProductId, err)
			return nil, err
		}
		targetQty = qty
		if targetQty > product.Stock {
			logger.Errorf("AddToCart stock insufficient productID=%d stock=%d qty=%d", req.ProductId, product.Stock, targetQty)
			return nil, codeerror.ErrorStockInsufficient
		}
		if anyItem != nil && anyItem.Deleted == common.DeletedYes {
			if err := cartDao.Revive(anyItem, targetQty); err != nil {
				logger.Errorf("AddToCart Revive id=%d err=%v", anyItem.ID, err)
				return nil, err
			}
			item = anyItem
		} else {
			item = &cartmodel.CartItem{
				UserID:    info.UserID,
				ProductID: req.ProductId,
				Quantity:  targetQty,
			}
			if err := cartDao.Create(item); err != nil {
				logger.Errorf("AddToCart Create userID=%d productID=%d err=%v", info.UserID, req.ProductId, err)
				return nil, err
			}
		}
	} else {
		targetQty = item.Quantity + qty
		if targetQty > product.Stock {
			logger.Errorf("AddToCart stock insufficient productID=%d stock=%d qty=%d", req.ProductId, product.Stock, targetQty)
			return nil, codeerror.ErrorStockInsufficient
		}
		itemID := item.ID
		item, err = cartDao.UpdateQuantity(itemID, info.UserID, targetQty)
		if err != nil {
			logger.Errorf("AddToCart UpdateQuantity id=%d qty=%d err=%v", itemID, targetQty, err)
			return nil, err
		}
	}

	cover, err := firstImage(product.ID)
	if err != nil {
		logger.Errorf("AddToCart firstImage productID=%d err=%v", product.ID, err)
		return nil, err
	}
	return &apicart.CartItemReply{Item: toCartItem(item, product, cover)}, nil
}

func (c *Controller) RemoveFromCart(ctx context.Context, req *apicart.RemoveFromCartRequest) (*apicart.RemoveFromCartReply, error) {
	info := auth.MustFromContext(ctx)
	cartDao := cartdao.NewCartDao(nil)

	switch {
	case req.Id > 0:
		if err := cartDao.SoftDeleteByID(req.Id, info.UserID); err != nil {
			logger.Errorf("RemoveFromCart SoftDeleteByID id=%d userID=%d err=%v", req.Id, info.UserID, err)
			return nil, err
		}
	case req.ProductId > 0:
		if err := cartDao.SoftDeleteByUserProduct(info.UserID, req.ProductId); err != nil {
			logger.Errorf("RemoveFromCart SoftDeleteByUserProduct userID=%d productID=%d err=%v", info.UserID, req.ProductId, err)
			return nil, err
		}
	default:
		logger.Errorf("RemoveFromCart missing id/product_id userID=%d", info.UserID)
		return nil, codeerror.ErrorParam
	}
	return &apicart.RemoveFromCartReply{}, nil
}

func (c *Controller) UpdateCartQuantity(ctx context.Context, req *apicart.UpdateCartQuantityRequest) (*apicart.CartItemReply, error) {
	info := auth.MustFromContext(ctx)
	if req.Id <= 0 {
		logger.Errorf("UpdateCartQuantity invalid id=%d userID=%d", req.Id, info.UserID)
		return nil, codeerror.ErrorParam
	}

	cartDao := cartdao.NewCartDao(nil)
	if req.Quantity <= 0 {
		if err := cartDao.SoftDeleteByID(req.Id, info.UserID); err != nil {
			logger.Errorf("UpdateCartQuantity SoftDeleteByID id=%d userID=%d err=%v", req.Id, info.UserID, err)
			return nil, err
		}
		return &apicart.CartItemReply{}, nil
	}

	item, err := cartDao.GetByID(req.Id, info.UserID)
	if err != nil {
		logger.Errorf("UpdateCartQuantity GetByID id=%d userID=%d err=%v", req.Id, info.UserID, err)
		return nil, err
	}
	product, err := productdao.NewProductDao(nil).GetByID(item.ProductID)
	if err != nil {
		logger.Errorf("UpdateCartQuantity GetProduct productID=%d err=%v", item.ProductID, err)
		return nil, err
	}
	if product.Status != productmodel.StatusOn {
		logger.Errorf("UpdateCartQuantity product not on shelf productID=%d status=%d", item.ProductID, product.Status)
		return nil, codeerror.ErrorParam.Errorf("product not available")
	}
	if req.Quantity > product.Stock {
		logger.Errorf("UpdateCartQuantity stock insufficient productID=%d stock=%d qty=%d", item.ProductID, product.Stock, req.Quantity)
		return nil, codeerror.ErrorStockInsufficient
	}

	item, err = cartDao.UpdateQuantity(req.Id, info.UserID, req.Quantity)
	if err != nil {
		logger.Errorf("UpdateCartQuantity UpdateQuantity id=%d qty=%d err=%v", req.Id, req.Quantity, err)
		return nil, err
	}
	cover, err := firstImage(product.ID)
	if err != nil {
		logger.Errorf("UpdateCartQuantity firstImage productID=%d err=%v", product.ID, err)
		return nil, err
	}
	return &apicart.CartItemReply{Item: toCartItem(item, product, cover)}, nil
}

func (c *Controller) ListCart(ctx context.Context, req *apicart.ListCartRequest) (*apicart.ListCartReply, error) {
	info := auth.MustFromContext(ctx)
	list, err := cartdao.NewCartDao(nil).ListByUserID(info.UserID)
	if err != nil {
		logger.Errorf("ListCart ListByUserID userID=%d err=%v", info.UserID, err)
		return nil, err
	}

	out := make([]*apicart.CartItem, 0, len(list))
	var totalQty int64
	for i := range list {
		item := &list[i]
		totalQty += int64(item.Quantity)
		product, err := productdao.NewProductDao(nil).GetByID(item.ProductID)
		if err != nil {
			logger.Errorf("ListCart GetProduct productID=%d err=%v", item.ProductID, err)
			if ce, ok := codeerror.As(err); ok && ce.Code == codeerror.ErrorParam.Code {
				out = append(out, &apicart.CartItem{
					Id:        item.ID,
					UserId:    item.UserID,
					ProductId: item.ProductID,
					Quantity:  item.Quantity,
					Ct:        item.Ct,
					Ut:        item.Ut,
				})
				continue
			}
			return nil, err
		}
		cover, err := firstImage(product.ID)
		if err != nil {
			logger.Errorf("ListCart firstImage productID=%d err=%v", product.ID, err)
			return nil, err
		}
		out = append(out, toCartItem(item, product, cover))
	}
	return &apicart.ListCartReply{
		List:          out,
		TotalQuantity: totalQty,
	}, nil
}

func firstImage(productID int64) (string, error) {
	images, err := productdao.NewImageDao(nil).ListByProductID(productID)
	if err != nil {
		return "", err
	}
	if len(images) == 0 {
		return "", nil
	}
	return images[0].URL, nil
}

func toCartItem(item *cartmodel.CartItem, product *productmodel.Product, cover string) *apicart.CartItem {
	if item == nil {
		return nil
	}
	out := &apicart.CartItem{
		Id:        item.ID,
		UserId:    item.UserID,
		ProductId: item.ProductID,
		Quantity:  item.Quantity,
		Ct:        item.Ct,
		Ut:        item.Ut,
	}
	if product != nil {
		out.ProductName = product.Name
		out.ProductPrice = product.Price
		out.ProductStock = product.Stock
		out.ProductStatus = product.Status
		out.ProductCover = cover
	}
	return out
}
