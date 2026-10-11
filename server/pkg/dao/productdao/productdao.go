package productdao

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"helloworld/base/codeerror"
	"helloworld/common/database"
	productmodel "helloworld/pkg/model/product"
)

func db() *gorm.DB {
	return database.GetDb()
}

func AutoMigrate() error {
	if err := db().AutoMigrate(&productmodel.Category{}, &productmodel.Product{}); err != nil {
		return codeerror.OrSystem(err, "AutoMigrate product tables")
	}
	return nil
}

func normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

// ----- category -----

func CreateCategory(name string, sort int32, status int32) (*productmodel.Category, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, codeerror.OrSystem(codeerror.ErrorParam, "CreateCategory empty name")
	}
	if status == 0 {
		status = productmodel.StatusOn
	}
	now := time.Now().Unix()
	c := &productmodel.Category{
		Name:   name,
		Sort:   sort,
		Status: status,
		Ct:     now,
		Ut:     now,
	}
	if err := db().Create(c).Error; err != nil {
		return nil, codeerror.OrSystem(err, "CreateCategory name=%s", name)
	}
	return c, nil
}

func UpdateCategory(id int64, name string, sort int32, status int32) (*productmodel.Category, error) {
	c, err := GetCategoryByID(id)
	if err != nil {
		return nil, codeerror.OrSystem(err, "UpdateCategory GetCategoryByID id=%d", id)
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, codeerror.OrSystem(codeerror.ErrorParam, "UpdateCategory empty name id=%d", id)
	}
	if status != productmodel.StatusOn && status != productmodel.StatusOff {
		return nil, codeerror.OrSystem(codeerror.ErrorParam, "UpdateCategory invalid status=%d id=%d", status, id)
	}
	c.Name = name
	c.Sort = sort
	c.Status = status
	c.Ut = time.Now().Unix()
	if err := db().Model(c).Select("name", "sort", "status", "ut").Updates(c).Error; err != nil {
		return nil, codeerror.OrSystem(err, "UpdateCategory id=%d", id)
	}
	return c, nil
}

func GetCategoryByID(id int64) (*productmodel.Category, error) {
	var c productmodel.Category
	err := db().Where("id = ?", id).First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, codeerror.OrSystem(codeerror.ErrorParam.Errorf("category not found"), "GetCategoryByID not found id=%d", id)
	}
	if err != nil {
		return nil, codeerror.OrSystem(err, "GetCategoryByID id=%d", id)
	}
	return &c, nil
}

func ListCategories(onlyOn bool) ([]productmodel.Category, error) {
	q := db().Model(&productmodel.Category{})
	if onlyOn {
		q = q.Where("status = ?", productmodel.StatusOn)
	}
	var list []productmodel.Category
	if err := q.Order("sort desc, id asc").Find(&list).Error; err != nil {
		return nil, codeerror.OrSystem(err, "ListCategories onlyOn=%v", onlyOn)
	}
	return list, nil
}

// ----- product -----

type ProductQuery struct {
	CategoryID int64
	Status     int32 // 0=全部
	Keyword    string
	Page       int
	PageSize   int
}

func CreateProduct(p *productmodel.Product) (*productmodel.Product, error) {
	if p == nil || strings.TrimSpace(p.Name) == "" || p.CategoryID <= 0 || p.Price < 0 || p.Stock < 0 {
		return nil, codeerror.OrSystem(codeerror.ErrorParam, "CreateProduct invalid param")
	}
	if _, err := GetCategoryByID(p.CategoryID); err != nil {
		return nil, codeerror.OrSystem(err, "CreateProduct category id=%d", p.CategoryID)
	}
	if p.Status == 0 {
		p.Status = productmodel.StatusOff
	}
	if p.Status != productmodel.StatusOn && p.Status != productmodel.StatusOff {
		return nil, codeerror.OrSystem(codeerror.ErrorParam, "CreateProduct invalid status=%d", p.Status)
	}
	now := time.Now().Unix()
	p.Name = strings.TrimSpace(p.Name)
	p.Ct = now
	p.Ut = now
	if err := db().Create(p).Error; err != nil {
		return nil, codeerror.OrSystem(err, "CreateProduct name=%s", p.Name)
	}
	return p, nil
}

func GetProductByID(id int64) (*productmodel.Product, error) {
	var p productmodel.Product
	err := db().Where("id = ?", id).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, codeerror.OrSystem(codeerror.ErrorParam.Errorf("product not found"), "GetProductByID not found id=%d", id)
	}
	if err != nil {
		return nil, codeerror.OrSystem(err, "GetProductByID id=%d", id)
	}
	return &p, nil
}

func ListProducts(q ProductQuery) ([]productmodel.Product, int64, error) {
	page, pageSize := normalizePage(q.Page, q.PageSize)
	tx := db().Model(&productmodel.Product{})
	if q.CategoryID > 0 {
		tx = tx.Where("category_id = ?", q.CategoryID)
	}
	if q.Status > 0 {
		tx = tx.Where("status = ?", q.Status)
	}
	if kw := strings.TrimSpace(q.Keyword); kw != "" {
		tx = tx.Where("name LIKE ?", "%"+kw+"%")
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, codeerror.OrSystem(err, "ListProducts count")
	}
	var list []productmodel.Product
	if err := tx.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, codeerror.OrSystem(err, "ListProducts page=%d", page)
	}
	return list, total, nil
}

func UpdateProduct(id int64, name, description, cover string, price int64) (*productmodel.Product, error) {
	p, err := GetProductByID(id)
	if err != nil {
		return nil, codeerror.OrSystem(err, "UpdateProduct GetProductByID id=%d", id)
	}
	name = strings.TrimSpace(name)
	if name == "" || price < 0 {
		return nil, codeerror.OrSystem(codeerror.ErrorParam, "UpdateProduct invalid param id=%d", id)
	}
	p.Name = name
	p.Description = description
	p.Cover = cover
	p.Price = price
	p.Ut = time.Now().Unix()
	if err := db().Model(p).Select("name", "description", "cover", "price", "ut").Updates(p).Error; err != nil {
		return nil, codeerror.OrSystem(err, "UpdateProduct id=%d", id)
	}
	return p, nil
}

func SetProductStatus(id int64, status int32) (*productmodel.Product, error) {
	if status != productmodel.StatusOn && status != productmodel.StatusOff {
		return nil, codeerror.OrSystem(codeerror.ErrorParam, "SetProductStatus invalid status=%d id=%d", status, id)
	}
	p, err := GetProductByID(id)
	if err != nil {
		return nil, codeerror.OrSystem(err, "SetProductStatus GetProductByID id=%d", id)
	}
	p.Status = status
	p.Ut = time.Now().Unix()
	if err := db().Model(p).Select("status", "ut").Updates(p).Error; err != nil {
		return nil, codeerror.OrSystem(err, "SetProductStatus id=%d status=%d", id, status)
	}
	return p, nil
}

func UpdateProductStock(id int64, stock int32) (*productmodel.Product, error) {
	if stock < 0 {
		return nil, codeerror.OrSystem(codeerror.ErrorParam, "UpdateProductStock invalid stock=%d id=%d", stock, id)
	}
	p, err := GetProductByID(id)
	if err != nil {
		return nil, codeerror.OrSystem(err, "UpdateProductStock GetProductByID id=%d", id)
	}
	p.Stock = stock
	p.Ut = time.Now().Unix()
	if err := db().Model(p).Select("stock", "ut").Updates(p).Error; err != nil {
		return nil, codeerror.OrSystem(err, "UpdateProductStock id=%d stock=%d", id, stock)
	}
	return p, nil
}

func UpdateProductCategory(id, categoryID int64) (*productmodel.Product, error) {
	if categoryID <= 0 {
		return nil, codeerror.OrSystem(codeerror.ErrorParam, "UpdateProductCategory invalid categoryID=%d id=%d", categoryID, id)
	}
	if _, err := GetCategoryByID(categoryID); err != nil {
		return nil, codeerror.OrSystem(err, "UpdateProductCategory category id=%d", categoryID)
	}
	p, err := GetProductByID(id)
	if err != nil {
		return nil, codeerror.OrSystem(err, "UpdateProductCategory GetProductByID id=%d", id)
	}
	p.CategoryID = categoryID
	p.Ut = time.Now().Unix()
	if err := db().Model(p).Select("category_id", "ut").Updates(p).Error; err != nil {
		return nil, codeerror.OrSystem(err, "UpdateProductCategory id=%d categoryID=%d", id, categoryID)
	}
	return p, nil
}
