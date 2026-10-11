package cartdao

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"helloworld/base/codeerror"
	cartmodel "helloworld/pkg/model/cart"
)

type CartDao struct {
	db *gorm.DB
}

func (d *CartDao) WithTx(tx *gorm.DB) *CartDao {
	return &CartDao{db: tx}
}

func (d *CartDao) GetByID(id, userID int64) (*cartmodel.CartItem, error) {
	var item cartmodel.CartItem
	err := d.db.Where("id = ? AND user_id = ?", id, userID).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, codeerror.OrSystem(codeerror.ErrorParam.Errorf("cart item not found"), "CartDao.GetByID not found id=%d userID=%d", id, userID)
	}
	if err != nil {
		return nil, codeerror.OrSystem(err, "CartDao.GetByID id=%d userID=%d", id, userID)
	}
	return &item, nil
}

func (d *CartDao) FindByUserProduct(userID, productID int64) (*cartmodel.CartItem, error) {
	var item cartmodel.CartItem
	err := d.db.Where("user_id = ? AND product_id = ?", userID, productID).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, codeerror.OrSystem(err, "CartDao.FindByUserProduct userID=%d productID=%d", userID, productID)
	}
	return &item, nil
}

func (d *CartDao) ListByUserID(userID int64) ([]cartmodel.CartItem, error) {
	var list []cartmodel.CartItem
	if err := d.db.Where("user_id = ?", userID).Order("ut desc, id desc").Find(&list).Error; err != nil {
		return nil, codeerror.OrSystem(err, "CartDao.ListByUserID userID=%d", userID)
	}
	return list, nil
}

func (d *CartDao) Create(item *cartmodel.CartItem) error {
	if item == nil || item.UserID <= 0 || item.ProductID <= 0 || item.Quantity <= 0 {
		return codeerror.OrSystem(codeerror.ErrorParam, "CartDao.Create invalid item")
	}
	now := time.Now().Unix()
	item.Ct = now
	item.Ut = now
	if err := d.db.Create(item).Error; err != nil {
		return codeerror.OrSystem(err, "CartDao.Create userID=%d productID=%d", item.UserID, item.ProductID)
	}
	return nil
}

func (d *CartDao) UpdateQuantity(id, userID int64, quantity int32) (*cartmodel.CartItem, error) {
	item, err := d.GetByID(id, userID)
	if err != nil {
		return nil, codeerror.OrSystem(err, "CartDao.UpdateQuantity GetByID id=%d", id)
	}
	if quantity <= 0 {
		return nil, codeerror.OrSystem(codeerror.ErrorParam, "CartDao.UpdateQuantity invalid quantity=%d id=%d", quantity, id)
	}
	item.Quantity = quantity
	item.Ut = time.Now().Unix()
	if err := d.db.Model(item).Select("quantity", "ut").Updates(item).Error; err != nil {
		return nil, codeerror.OrSystem(err, "CartDao.UpdateQuantity id=%d quantity=%d", id, quantity)
	}
	return item, nil
}

func (d *CartDao) DeleteByID(id, userID int64) error {
	res := d.db.Where("id = ? AND user_id = ?", id, userID).Delete(&cartmodel.CartItem{})
	if res.Error != nil {
		return codeerror.OrSystem(res.Error, "CartDao.DeleteByID id=%d userID=%d", id, userID)
	}
	if res.RowsAffected == 0 {
		return codeerror.OrSystem(codeerror.ErrorParam.Errorf("cart item not found"), "CartDao.DeleteByID not found id=%d userID=%d", id, userID)
	}
	return nil
}

func (d *CartDao) DeleteByUserProduct(userID, productID int64) error {
	res := d.db.Where("user_id = ? AND product_id = ?", userID, productID).Delete(&cartmodel.CartItem{})
	if res.Error != nil {
		return codeerror.OrSystem(res.Error, "CartDao.DeleteByUserProduct userID=%d productID=%d", userID, productID)
	}
	if res.RowsAffected == 0 {
		return codeerror.OrSystem(codeerror.ErrorParam.Errorf("cart item not found"), "CartDao.DeleteByUserProduct not found userID=%d productID=%d", userID, productID)
	}
	return nil
}
