package productdao

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	productmodel "helloworld/pkg/model/product"
	"mall/kit/codeerror"
)

type CategoryDao struct {
	db *gorm.DB
}

func (d *CategoryDao) WithTx(tx *gorm.DB) *CategoryDao {
	return &CategoryDao{db: tx}
}

func (d *CategoryDao) Create(name string, sort int32, status int32) (*productmodel.Category, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, codeerror.OrSystem(codeerror.ErrorParam, "CategoryDao.Create empty name")
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
	if err := d.db.Create(c).Error; err != nil {
		return nil, codeerror.OrSystem(err, "CategoryDao.Create name=%s", name)
	}
	return c, nil
}

func (d *CategoryDao) Update(id int64, name string, sort int32, status int32) (*productmodel.Category, error) {
	c, err := d.GetByID(id)
	if err != nil {
		return nil, codeerror.OrSystem(err, "CategoryDao.Update GetByID id=%d", id)
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, codeerror.OrSystem(codeerror.ErrorParam, "CategoryDao.Update empty name id=%d", id)
	}
	if status != productmodel.StatusOn && status != productmodel.StatusOff {
		return nil, codeerror.OrSystem(codeerror.ErrorParam, "CategoryDao.Update invalid status=%d id=%d", status, id)
	}
	c.Name = name
	c.Sort = sort
	c.Status = status
	c.Ut = time.Now().Unix()
	if err := d.db.Model(c).Select("name", "sort", "status", "ut").Updates(c).Error; err != nil {
		return nil, codeerror.OrSystem(err, "CategoryDao.Update id=%d", id)
	}
	return c, nil
}

func (d *CategoryDao) GetByID(id int64) (*productmodel.Category, error) {
	var c productmodel.Category
	err := d.db.Where("id = ?", id).First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, codeerror.OrSystem(codeerror.ErrorParam.Errorf("category not found"), "CategoryDao.GetByID not found id=%d", id)
	}
	if err != nil {
		return nil, codeerror.OrSystem(err, "CategoryDao.GetByID id=%d", id)
	}
	return &c, nil
}

func (d *CategoryDao) List(onlyOn bool) ([]productmodel.Category, error) {
	q := d.db.Model(&productmodel.Category{})
	if onlyOn {
		q = q.Where("status = ?", productmodel.StatusOn)
	}
	var list []productmodel.Category
	if err := q.Order("sort desc, id asc").Find(&list).Error; err != nil {
		return nil, codeerror.OrSystem(err, "CategoryDao.List onlyOn=%v", onlyOn)
	}
	return list, nil
}
