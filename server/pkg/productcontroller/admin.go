package productcontroller

import (
	"context"
	"strings"

	"gorm.io/gorm"
	apiproduct "helloworld/api/product"
	"helloworld/base/codeerror"
	"helloworld/base/logger"
	"helloworld/common/auth"
	"helloworld/common/database"
	"helloworld/pkg/dao/productdao"
	productmodel "helloworld/pkg/model/product"
)

func (c *Controller) AdminCreateCategory(ctx context.Context, req *apiproduct.AdminCreateCategoryRequest) (*apiproduct.AdminCategoryReply, error) {
	if _, err := auth.RequireAdmin(ctx); err != nil {
		logger.Errorf("AdminCreateCategory RequireAdmin err=%v", err)
		return nil, err
	}
	cat, err := productdao.NewCategoryDao(nil).Create(req.Name, req.Sort, int32(req.Status))
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
	cat, err := productdao.NewCategoryDao(nil).Update(req.Id, req.Name, req.Sort, int32(req.Status))
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
	list, err := productdao.NewCategoryDao(nil).List(false)
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

	var (
		created *productmodel.Product
		images  []productmodel.ProductImage
	)
	err := database.GetDb().Transaction(func(tx *gorm.DB) error {
		categoryDao := productdao.NewCategoryDao(tx)
		productDao := productdao.NewProductDao(tx)
		imageDao := productdao.NewImageDao(tx)

		if _, err := categoryDao.GetByID(req.CategoryId); err != nil {
			return err
		}
		p, err := productDao.Create(&productmodel.Product{
			CategoryID:  req.CategoryId,
			Name:        req.Name,
			Description: req.Description,
			Price:       req.Price,
			Stock:       req.Stock,
			Status:      int32(req.Status),
		})
		if err != nil {
			return err
		}
		if err := imageDao.Replace(p.ID, normalizeImageURLs(req.Images)); err != nil {
			return err
		}
		imgs, err := imageDao.ListByProductID(p.ID)
		if err != nil {
			return err
		}
		created = p
		images = imgs
		return nil
	})
	if err != nil {
		logger.Errorf("AdminCreateProduct name=%s err=%v", req.Name, err)
		return nil, codeerror.OrSystem(err, "AdminCreateProduct name=%s", req.Name)
	}
	return &apiproduct.AdminProductReply{Product: toProductInfo(created, images)}, nil
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

	var (
		updated *productmodel.Product
		images  []productmodel.ProductImage
	)
	err := database.GetDb().Transaction(func(tx *gorm.DB) error {
		productDao := productdao.NewProductDao(tx)
		imageDao := productdao.NewImageDao(tx)

		p, err := productDao.Update(req.Id, req.Name, req.Description, req.Price)
		if err != nil {
			return err
		}
		if err := imageDao.Replace(p.ID, normalizeImageURLs(req.Images)); err != nil {
			return err
		}
		imgs, err := imageDao.ListByProductID(p.ID)
		if err != nil {
			return err
		}
		updated = p
		images = imgs
		return nil
	})
	if err != nil {
		logger.Errorf("AdminUpdateProduct id=%d err=%v", req.Id, err)
		return nil, codeerror.OrSystem(err, "AdminUpdateProduct id=%d", req.Id)
	}
	return &apiproduct.AdminProductReply{Product: toProductInfo(updated, images)}, nil
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
	p, err := productdao.NewProductDao(nil).SetStatus(req.Id, int32(req.Status))
	if err != nil {
		logger.Errorf("AdminSetProductStatus id=%d status=%v err=%v", req.Id, req.Status, err)
		return nil, err
	}
	images, err := productdao.NewImageDao(nil).ListByProductID(p.ID)
	if err != nil {
		logger.Errorf("AdminSetProductStatus ListByProductID id=%d err=%v", p.ID, err)
		return nil, err
	}
	return &apiproduct.AdminProductReply{Product: toProductInfo(p, images)}, nil
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
	p, err := productdao.NewProductDao(nil).UpdateStock(req.Id, req.Stock)
	if err != nil {
		logger.Errorf("AdminUpdateProductStock id=%d stock=%d err=%v", req.Id, req.Stock, err)
		return nil, err
	}
	images, err := productdao.NewImageDao(nil).ListByProductID(p.ID)
	if err != nil {
		logger.Errorf("AdminUpdateProductStock ListByProductID id=%d err=%v", p.ID, err)
		return nil, err
	}
	return &apiproduct.AdminProductReply{Product: toProductInfo(p, images)}, nil
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

	var (
		updated *productmodel.Product
		images  []productmodel.ProductImage
	)
	err := database.GetDb().Transaction(func(tx *gorm.DB) error {
		categoryDao := productdao.NewCategoryDao(tx)
		productDao := productdao.NewProductDao(tx)
		imageDao := productdao.NewImageDao(tx)

		if _, err := categoryDao.GetByID(req.CategoryId); err != nil {
			return err
		}
		p, err := productDao.UpdateCategory(req.Id, req.CategoryId)
		if err != nil {
			return err
		}
		imgs, err := imageDao.ListByProductID(p.ID)
		if err != nil {
			return err
		}
		updated = p
		images = imgs
		return nil
	})
	if err != nil {
		logger.Errorf("AdminUpdateProductCategory id=%d categoryID=%d err=%v", req.Id, req.CategoryId, err)
		return nil, codeerror.OrSystem(err, "AdminUpdateProductCategory id=%d", req.Id)
	}
	return &apiproduct.AdminProductReply{Product: toProductInfo(updated, images)}, nil
}

func (c *Controller) AdminSetProductHot(ctx context.Context, req *apiproduct.AdminSetProductHotRequest) (*apiproduct.AdminProductReply, error) {
	if _, err := auth.RequireAdmin(ctx); err != nil {
		logger.Errorf("AdminSetProductHot RequireAdmin err=%v", err)
		return nil, err
	}
	if req.Id <= 0 {
		logger.Errorf("AdminSetProductHot invalid id=%d", req.Id)
		return nil, codeerror.ErrorParam
	}
	p, err := productdao.NewProductDao(nil).SetHot(req.Id, req.IsHot)
	if err != nil {
		logger.Errorf("AdminSetProductHot id=%d isHot=%v err=%v", req.Id, req.IsHot, err)
		return nil, err
	}
	images, err := productdao.NewImageDao(nil).ListByProductID(p.ID)
	if err != nil {
		logger.Errorf("AdminSetProductHot ListByProductID id=%d err=%v", p.ID, err)
		return nil, err
	}
	return &apiproduct.AdminProductReply{Product: toProductInfo(p, images)}, nil
}

func (c *Controller) AdminListProducts(ctx context.Context, req *apiproduct.AdminListProductsRequest) (*apiproduct.ListProductsReply, error) {
	if _, err := auth.RequireAdmin(ctx); err != nil {
		logger.Errorf("AdminListProducts RequireAdmin err=%v", err)
		return nil, err
	}
	list, total, err := productdao.NewProductDao(nil).List(productdao.ProductQuery{
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
	imageMap, err := productdao.NewImageDao(nil).ListByProductIDs(productIDs(list))
	if err != nil {
		logger.Errorf("AdminListProducts ListByProductIDs err=%v", err)
		return nil, err
	}
	return &apiproduct.ListProductsReply{List: toProductInfos(list, imageMap), Total: total}, nil
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
	p, err := productdao.NewProductDao(nil).GetByID(req.Id)
	if err != nil {
		logger.Errorf("AdminGetProduct id=%d err=%v", req.Id, err)
		return nil, err
	}
	images, err := productdao.NewImageDao(nil).ListByProductID(p.ID)
	if err != nil {
		logger.Errorf("AdminGetProduct ListByProductID id=%d err=%v", p.ID, err)
		return nil, err
	}
	return &apiproduct.GetProductReply{Product: toProductInfo(p, images)}, nil
}
