package productcontroller

import (
	"context"
	"strings"

	apiproduct "helloworld/api/product"
	"helloworld/base/codeerror"
	"helloworld/base/logger"
	"helloworld/common/auth"
	"helloworld/pkg/dao/productdao"
	productmodel "helloworld/pkg/model/product"
)

func (c *Controller) AdminCreateCategory(ctx context.Context, req *apiproduct.AdminCreateCategoryRequest) (*apiproduct.AdminCategoryReply, error) {
	if _, err := auth.RequireAdmin(ctx); err != nil {
		logger.Errorf("AdminCreateCategory RequireAdmin err=%v", err)
		return nil, err
	}
	cat, err := productdao.CreateCategory(req.Name, req.Sort, int32(req.Status))
	if err != nil {
		logger.Errorf("AdminCreateCategory name=%s err=%v", req.Name, err)
		return nil, err
	}
	return &apiproduct.AdminCategoryReply{Category: toCategoryInfo(cat)}, nil
}

func (c *Controller) AdminUpdateCategory(ctx context.Context, req *apiproduct.AdminUpdateCategoryRequest) (*apiproduct.AdminCategoryReply, error) {
	if _, err := auth.RequireAdmin(ctx); err != nil {
		logger.Errorf("AdminUpdateCategory RequireAdmin err=%v", err)
		return nil, err
	}
	if req.Id <= 0 {
		logger.Errorf("AdminUpdateCategory invalid id=%d", req.Id)
		return nil, codeerror.ErrorParam
	}
	cat, err := productdao.UpdateCategory(req.Id, req.Name, req.Sort, int32(req.Status))
	if err != nil {
		logger.Errorf("AdminUpdateCategory id=%d err=%v", req.Id, err)
		return nil, err
	}
	return &apiproduct.AdminCategoryReply{Category: toCategoryInfo(cat)}, nil
}

func (c *Controller) AdminListCategories(ctx context.Context, req *apiproduct.AdminListCategoriesRequest) (*apiproduct.ListCategoriesReply, error) {
	if _, err := auth.RequireAdmin(ctx); err != nil {
		logger.Errorf("AdminListCategories RequireAdmin err=%v", err)
		return nil, err
	}
	list, err := productdao.ListCategories(false)
	if err != nil {
		logger.Errorf("AdminListCategories err=%v", err)
		return nil, err
	}
	return &apiproduct.ListCategoriesReply{List: toCategoryInfos(list)}, nil
}

func (c *Controller) AdminCreateProduct(ctx context.Context, req *apiproduct.AdminCreateProductRequest) (*apiproduct.AdminProductReply, error) {
	if _, err := auth.RequireAdmin(ctx); err != nil {
		logger.Errorf("AdminCreateProduct RequireAdmin err=%v", err)
		return nil, err
	}
	p, err := productdao.CreateProduct(&productmodel.Product{
		CategoryID:  req.CategoryId,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		Cover:       req.Cover,
		Status:      int32(req.Status),
	})
	if err != nil {
		logger.Errorf("AdminCreateProduct name=%s err=%v", req.Name, err)
		return nil, err
	}
	return &apiproduct.AdminProductReply{Product: toProductInfo(p)}, nil
}

func (c *Controller) AdminUpdateProduct(ctx context.Context, req *apiproduct.AdminUpdateProductRequest) (*apiproduct.AdminProductReply, error) {
	if _, err := auth.RequireAdmin(ctx); err != nil {
		logger.Errorf("AdminUpdateProduct RequireAdmin err=%v", err)
		return nil, err
	}
	if req.Id <= 0 {
		logger.Errorf("AdminUpdateProduct invalid id=%d", req.Id)
		return nil, codeerror.ErrorParam
	}
	p, err := productdao.UpdateProduct(req.Id, req.Name, req.Description, req.Cover, req.Price)
	if err != nil {
		logger.Errorf("AdminUpdateProduct id=%d err=%v", req.Id, err)
		return nil, err
	}
	return &apiproduct.AdminProductReply{Product: toProductInfo(p)}, nil
}

func (c *Controller) AdminSetProductStatus(ctx context.Context, req *apiproduct.AdminSetProductStatusRequest) (*apiproduct.AdminProductReply, error) {
	if _, err := auth.RequireAdmin(ctx); err != nil {
		logger.Errorf("AdminSetProductStatus RequireAdmin err=%v", err)
		return nil, err
	}
	if req.Id <= 0 {
		logger.Errorf("AdminSetProductStatus invalid id=%d", req.Id)
		return nil, codeerror.ErrorParam
	}
	p, err := productdao.SetProductStatus(req.Id, int32(req.Status))
	if err != nil {
		logger.Errorf("AdminSetProductStatus id=%d status=%v err=%v", req.Id, req.Status, err)
		return nil, err
	}
	return &apiproduct.AdminProductReply{Product: toProductInfo(p)}, nil
}

func (c *Controller) AdminUpdateProductStock(ctx context.Context, req *apiproduct.AdminUpdateProductStockRequest) (*apiproduct.AdminProductReply, error) {
	if _, err := auth.RequireAdmin(ctx); err != nil {
		logger.Errorf("AdminUpdateProductStock RequireAdmin err=%v", err)
		return nil, err
	}
	if req.Id <= 0 {
		logger.Errorf("AdminUpdateProductStock invalid id=%d", req.Id)
		return nil, codeerror.ErrorParam
	}
	p, err := productdao.UpdateProductStock(req.Id, req.Stock)
	if err != nil {
		logger.Errorf("AdminUpdateProductStock id=%d stock=%d err=%v", req.Id, req.Stock, err)
		return nil, err
	}
	return &apiproduct.AdminProductReply{Product: toProductInfo(p)}, nil
}

func (c *Controller) AdminUpdateProductCategory(ctx context.Context, req *apiproduct.AdminUpdateProductCategoryRequest) (*apiproduct.AdminProductReply, error) {
	if _, err := auth.RequireAdmin(ctx); err != nil {
		logger.Errorf("AdminUpdateProductCategory RequireAdmin err=%v", err)
		return nil, err
	}
	if req.Id <= 0 || req.CategoryId <= 0 {
		logger.Errorf("AdminUpdateProductCategory invalid id=%d category_id=%d", req.Id, req.CategoryId)
		return nil, codeerror.ErrorParam
	}
	p, err := productdao.UpdateProductCategory(req.Id, req.CategoryId)
	if err != nil {
		logger.Errorf("AdminUpdateProductCategory id=%d categoryID=%d err=%v", req.Id, req.CategoryId, err)
		return nil, err
	}
	return &apiproduct.AdminProductReply{Product: toProductInfo(p)}, nil
}

func (c *Controller) AdminListProducts(ctx context.Context, req *apiproduct.AdminListProductsRequest) (*apiproduct.ListProductsReply, error) {
	if _, err := auth.RequireAdmin(ctx); err != nil {
		logger.Errorf("AdminListProducts RequireAdmin err=%v", err)
		return nil, err
	}
	list, total, err := productdao.ListProducts(productdao.ProductQuery{
		CategoryID: req.CategoryId,
		Status:     int32(req.Status),
		Keyword:    strings.TrimSpace(req.Keyword),
		Page:       int(req.Page),
		PageSize:   int(req.PageSize),
	})
	if err != nil {
		logger.Errorf("AdminListProducts err=%v", err)
		return nil, err
	}
	return &apiproduct.ListProductsReply{List: toProductInfos(list), Total: total}, nil
}

func (c *Controller) AdminGetProduct(ctx context.Context, req *apiproduct.AdminGetProductRequest) (*apiproduct.GetProductReply, error) {
	if _, err := auth.RequireAdmin(ctx); err != nil {
		logger.Errorf("AdminGetProduct RequireAdmin err=%v", err)
		return nil, err
	}
	if req.Id <= 0 {
		logger.Errorf("AdminGetProduct invalid id=%d", req.Id)
		return nil, codeerror.ErrorParam
	}
	p, err := productdao.GetProductByID(req.Id)
	if err != nil {
		logger.Errorf("AdminGetProduct id=%d err=%v", req.Id, err)
		return nil, err
	}
	return &apiproduct.GetProductReply{Product: toProductInfo(p)}, nil
}
