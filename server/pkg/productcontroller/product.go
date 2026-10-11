package productcontroller

import (
	"context"

	apiproduct "helloworld/api/product"
	"helloworld/base/codeerror"
	"helloworld/base/logger"
	"helloworld/pkg/dao/productdao"
	productmodel "helloworld/pkg/model/product"
)

var (
	_ apiproduct.ProductHTTPServer      = (*Controller)(nil)
	_ apiproduct.AdminProductHTTPServer = (*Controller)(nil)
)

type Controller struct{}

func New() *Controller {
	return &Controller{}
}

func (c *Controller) ListProducts(ctx context.Context, req *apiproduct.ListProductsRequest) (*apiproduct.ListProductsReply, error) {
	list, total, err := productdao.NewProductDao(nil).List(productdao.ProductQuery{
		Status:   productmodel.StatusOn,
		Keyword:  req.Keyword,
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
	})
	if err != nil {
		logger.Errorf("ListProducts err=%v", err)
		return nil, err
	}
	imageMap, err := productdao.NewImageDao(nil).ListByProductIDs(productIDs(list))
	if err != nil {
		logger.Errorf("ListProducts ListByProductIDs err=%v", err)
		return nil, err
	}
	return &apiproduct.ListProductsReply{List: toProductInfos(list, imageMap), Total: total}, nil
}

func (c *Controller) ListProductsByCategory(ctx context.Context, req *apiproduct.ListProductsByCategoryRequest) (*apiproduct.ListProductsReply, error) {
	if req.CategoryId <= 0 {
		logger.Errorf("ListProductsByCategory invalid category_id=%d", req.CategoryId)
		return nil, codeerror.ErrorParam
	}
	list, total, err := productdao.NewProductDao(nil).List(productdao.ProductQuery{
		CategoryID: req.CategoryId,
		Status:     productmodel.StatusOn,
		Page:       int(req.Page),
		PageSize:   int(req.PageSize),
	})
	if err != nil {
		logger.Errorf("ListProductsByCategory categoryID=%d err=%v", req.CategoryId, err)
		return nil, err
	}
	imageMap, err := productdao.NewImageDao(nil).ListByProductIDs(productIDs(list))
	if err != nil {
		logger.Errorf("ListProductsByCategory ListByProductIDs err=%v", err)
		return nil, err
	}
	return &apiproduct.ListProductsReply{List: toProductInfos(list, imageMap), Total: total}, nil
}

func (c *Controller) GetProduct(ctx context.Context, req *apiproduct.GetProductRequest) (*apiproduct.GetProductReply, error) {
	if req.Id <= 0 {
		logger.Errorf("GetProduct invalid id=%d", req.Id)
		return nil, codeerror.ErrorParam
	}
	p, err := productdao.NewProductDao(nil).GetByID(req.Id)
	if err != nil {
		logger.Errorf("GetProduct id=%d err=%v", req.Id, err)
		return nil, err
	}
	if p.Status != productmodel.StatusOn {
		logger.Errorf("GetProduct not on shelf id=%d status=%d", req.Id, p.Status)
		return nil, codeerror.ErrorParam.Errorf("product not found")
	}
	images, err := productdao.NewImageDao(nil).ListByProductID(p.ID)
	if err != nil {
		logger.Errorf("GetProduct ListByProductID id=%d err=%v", p.ID, err)
		return nil, err
	}
	return &apiproduct.GetProductReply{Product: toProductInfo(p, images)}, nil
}

func (c *Controller) ListCategories(ctx context.Context, req *apiproduct.ListCategoriesRequest) (*apiproduct.ListCategoriesReply, error) {
	list, err := productdao.NewCategoryDao(nil).List(true)
	if err != nil {
		logger.Errorf("ListCategories err=%v", err)
		return nil, err
	}
	return &apiproduct.ListCategoriesReply{List: toCategoryInfos(list)}, nil
}
