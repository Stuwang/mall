package productdao

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"helloworld/base/codeerror"
	productmodel "helloworld/pkg/model/product"
)

type ProductDao struct {
	db *gorm.DB
}

func (d *ProductDao) WithTx(tx *gorm.DB) *ProductDao {
	return &ProductDao{db: tx}
}

type ProductQuery struct {
	CategoryID int64
	Status     int32 // 0=全部
	Keyword    string
	Page       int
	PageSize   int
}

func (d *ProductDao) Create(p *productmodel.Product) (*productmodel.Product, error) {
	if p == nil || strings.TrimSpace(p.Name) == "" || p.CategoryID <= 0 || p.Price < 0 || p.Stock < 0 {
		return nil, codeerror.OrSystem(codeerror.ErrorParam, "ProductDao.Create invalid param")
	}
	if p.Status == 0 {
		p.Status = productmodel.StatusOff
	}
	if p.Status != productmodel.StatusOn && p.Status != productmodel.StatusOff {
		return nil, codeerror.OrSystem(codeerror.ErrorParam, "ProductDao.Create invalid status=%d", p.Status)
	}
	now := time.Now().Unix()
	p.Name = strings.TrimSpace(p.Name)
	p.Ct = now
	p.Ut = now
	if err := d.db.Create(p).Error; err != nil {
		return nil, codeerror.OrSystem(err, "ProductDao.Create name=%s", p.Name)
	}
	return p, nil
}

func (d *ProductDao) GetByID(id int64) (*productmodel.Product, error) {
	var p productmodel.Product
	err := d.db.Where("id = ?", id).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, codeerror.OrSystem(codeerror.ErrorParam.Errorf("product not found"), "ProductDao.GetByID not found id=%d", id)
	}
	if err != nil {
		return nil, codeerror.OrSystem(err, "ProductDao.GetByID id=%d", id)
	}
	return &p, nil
}

func (d *ProductDao) List(q ProductQuery) ([]productmodel.Product, int64, error) {
	page, pageSize := normalizePage(q.Page, q.PageSize)
	tx := d.db.Model(&productmodel.Product{})
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
		return nil, 0, codeerror.OrSystem(err, "ProductDao.List count")
	}
	var list []productmodel.Product
	if err := tx.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, codeerror.OrSystem(err, "ProductDao.List page=%d", page)
	}
	return list, total, nil
}

func (d *ProductDao) Update(id int64, name, description string, price int64) (*productmodel.Product, error) {
	p, err := d.GetByID(id)
	if err != nil {
		return nil, codeerror.OrSystem(err, "ProductDao.Update GetByID id=%d", id)
	}
	name = strings.TrimSpace(name)
	if name == "" || price < 0 {
		return nil, codeerror.OrSystem(codeerror.ErrorParam, "ProductDao.Update invalid param id=%d", id)
	}
	p.Name = name
	p.Description = description
	p.Price = price
	p.Ut = time.Now().Unix()
	if err := d.db.Model(p).Select("name", "description", "price", "ut").Updates(p).Error; err != nil {
		return nil, codeerror.OrSystem(err, "ProductDao.Update id=%d", id)
	}
	return p, nil
}

func (d *ProductDao) SetStatus(id int64, status int32) (*productmodel.Product, error) {
	if status != productmodel.StatusOn && status != productmodel.StatusOff {
		return nil, codeerror.OrSystem(codeerror.ErrorParam, "ProductDao.SetStatus invalid status=%d id=%d", status, id)
	}
	p, err := d.GetByID(id)
	if err != nil {
		return nil, codeerror.OrSystem(err, "ProductDao.SetStatus GetByID id=%d", id)
	}
	p.Status = status
	p.Ut = time.Now().Unix()
	if err := d.db.Model(p).Select("status", "ut").Updates(p).Error; err != nil {
		return nil, codeerror.OrSystem(err, "ProductDao.SetStatus id=%d status=%d", id, status)
	}
	return p, nil
}

func (d *ProductDao) UpdateStock(id int64, stock int32) (*productmodel.Product, error) {
	if stock < 0 {
		return nil, codeerror.OrSystem(codeerror.ErrorParam, "ProductDao.UpdateStock invalid stock=%d id=%d", stock, id)
	}
	p, err := d.GetByID(id)
	if err != nil {
		return nil, codeerror.OrSystem(err, "ProductDao.UpdateStock GetByID id=%d", id)
	}
	p.Stock = stock
	p.Ut = time.Now().Unix()
	if err := d.db.Model(p).Select("stock", "ut").Updates(p).Error; err != nil {
		return nil, codeerror.OrSystem(err, "ProductDao.UpdateStock id=%d stock=%d", id, stock)
	}
	return p, nil
}

// DecStock 扣减库存（库存不足返回 ErrorStockInsufficient）
func (d *ProductDao) DecStock(id int64, qty int32) error {
	if id <= 0 || qty <= 0 {
		return codeerror.OrSystem(codeerror.ErrorParam, "ProductDao.DecStock invalid id=%d qty=%d", id, qty)
	}
	res := d.db.Model(&productmodel.Product{}).
		Where("id = ? AND stock >= ?", id, qty).
		Updates(map[string]interface{}{
			"stock": gorm.Expr("stock - ?", qty),
			"ut":    time.Now().Unix(),
		})
	if res.Error != nil {
		return codeerror.OrSystem(res.Error, "ProductDao.DecStock id=%d qty=%d", id, qty)
	}
	if res.RowsAffected == 0 {
		return codeerror.OrSystem(codeerror.ErrorStockInsufficient, "ProductDao.DecStock insufficient id=%d qty=%d", id, qty)
	}
	return nil
}

// IncStock 回补库存
func (d *ProductDao) IncStock(id int64, qty int32) error {
	if id <= 0 || qty <= 0 {
		return codeerror.OrSystem(codeerror.ErrorParam, "ProductDao.IncStock invalid id=%d qty=%d", id, qty)
	}
	res := d.db.Model(&productmodel.Product{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"stock": gorm.Expr("stock + ?", qty),
			"ut":    time.Now().Unix(),
		})
	if res.Error != nil {
		return codeerror.OrSystem(res.Error, "ProductDao.IncStock id=%d qty=%d", id, qty)
	}
	if res.RowsAffected == 0 {
		return codeerror.OrSystem(codeerror.ErrorParam.Errorf("product not found"), "ProductDao.IncStock not found id=%d", id)
	}
	return nil
}

func (d *ProductDao) UpdateCategory(id, categoryID int64) (*productmodel.Product, error) {
	if categoryID <= 0 {
		return nil, codeerror.OrSystem(codeerror.ErrorParam, "ProductDao.UpdateCategory invalid categoryID=%d id=%d", categoryID, id)
	}
	p, err := d.GetByID(id)
	if err != nil {
		return nil, codeerror.OrSystem(err, "ProductDao.UpdateCategory GetByID id=%d", id)
	}
	p.CategoryID = categoryID
	p.Ut = time.Now().Unix()
	if err := d.db.Model(p).Select("category_id", "ut").Updates(p).Error; err != nil {
		return nil, codeerror.OrSystem(err, "ProductDao.UpdateCategory id=%d categoryID=%d", id, categoryID)
	}
	return p, nil
}
