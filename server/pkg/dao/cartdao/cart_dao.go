package cartdao

import (
	"errors"
	"time"

	"gorm.io/gorm"
	cartmodel "helloworld/pkg/model/cart"
	"mall/kit/codeerror"
	"mall/kit/model/common"
)

type CartDao struct {
	db *gorm.DB
}

func (d *CartDao) WithTx(tx *gorm.DB) *CartDao {
	return &CartDao{db: tx}
}

func (d *CartDao) active() *gorm.DB {
	return d.db.Where("deleted = ?", common.DeletedNo)
}

func (d *CartDao) GetByID(id, userID int64) (*cartmodel.CartItem, error) {
	var item cartmodel.CartItem
	err := d.active().Where("id = ? AND user_id = ?", id, userID).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, codeerror.OrSystem(codeerror.ErrorParam.Errorf("cart item not found"), "CartDao.GetByID not found id=%d userID=%d", id, userID)
	}
	if err != nil {
		return nil, codeerror.OrSystem(err, "CartDao.GetByID id=%d userID=%d", id, userID)
	}
	return &item, nil
}

// FindByUserProduct 查询有效购物车项
func (d *CartDao) FindByUserProduct(userID, productID int64) (*cartmodel.CartItem, error) {
	var item cartmodel.CartItem
	err := d.active().Where("user_id = ? AND product_id = ?", userID, productID).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, codeerror.OrSystem(err, "CartDao.FindByUserProduct userID=%d productID=%d", userID, productID)
	}
	return &item, nil
}

// FindAnyByUserProduct 查询含软删除的购物车项（用于复活）
func (d *CartDao) FindAnyByUserProduct(userID, productID int64) (*cartmodel.CartItem, error) {
	var item cartmodel.CartItem
	err := d.db.Where("user_id = ? AND product_id = ?", userID, productID).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, codeerror.OrSystem(err, "CartDao.FindAnyByUserProduct userID=%d productID=%d", userID, productID)
	}
	return &item, nil
}

func (d *CartDao) ListByUserID(userID int64) ([]cartmodel.CartItem, error) {
	var list []cartmodel.CartItem
	if err := d.active().Where("user_id = ?", userID).Order("ut desc, id desc").Find(&list).Error; err != nil {
		return nil, codeerror.OrSystem(err, "CartDao.ListByUserID userID=%d", userID)
	}
	return list, nil
}

func (d *CartDao) Create(item *cartmodel.CartItem) error {
	if item == nil || item.UserID <= 0 || item.ProductID <= 0 || item.Quantity <= 0 {
		return codeerror.OrSystem(codeerror.ErrorParam, "CartDao.Create invalid item")
	}
	now := time.Now().Unix()
	item.Deleted = common.DeletedNo
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

// Revive 复活软删除项并设置数量
func (d *CartDao) Revive(item *cartmodel.CartItem, quantity int32) error {
	if item == nil || quantity <= 0 {
		return codeerror.OrSystem(codeerror.ErrorParam, "CartDao.Revive invalid")
	}
	item.Deleted = common.DeletedNo
	item.Quantity = quantity
	item.Ut = time.Now().Unix()
	if err := d.db.Model(item).Select("deleted", "quantity", "ut").Updates(item).Error; err != nil {
		return codeerror.OrSystem(err, "CartDao.Revive id=%d", item.ID)
	}
	return nil
}

func (d *CartDao) SoftDeleteByID(id, userID int64) error {
	res := d.active().Model(&cartmodel.CartItem{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(map[string]interface{}{
			"deleted": common.DeletedYes,
			"ut":      time.Now().Unix(),
		})
	if res.Error != nil {
		return codeerror.OrSystem(res.Error, "CartDao.SoftDeleteByID id=%d userID=%d", id, userID)
	}
	if res.RowsAffected == 0 {
		return codeerror.OrSystem(codeerror.ErrorParam.Errorf("cart item not found"), "CartDao.SoftDeleteByID not found id=%d userID=%d", id, userID)
	}
	return nil
}

func (d *CartDao) SoftDeleteByUserProduct(userID, productID int64) error {
	res := d.active().Model(&cartmodel.CartItem{}).
		Where("user_id = ? AND product_id = ?", userID, productID).
		Updates(map[string]interface{}{
			"deleted": common.DeletedYes,
			"ut":      time.Now().Unix(),
		})
	if res.Error != nil {
		return codeerror.OrSystem(res.Error, "CartDao.SoftDeleteByUserProduct userID=%d productID=%d", userID, productID)
	}
	if res.RowsAffected == 0 {
		return codeerror.OrSystem(codeerror.ErrorParam.Errorf("cart item not found"), "CartDao.SoftDeleteByUserProduct not found userID=%d productID=%d", userID, productID)
	}
	return nil
}

func (d *CartDao) ListByIDs(userID int64, ids []int64) ([]cartmodel.CartItem, error) {
	if len(ids) == 0 {
		return nil, codeerror.OrSystem(codeerror.ErrorParam, "CartDao.ListByIDs empty ids userID=%d", userID)
	}
	var list []cartmodel.CartItem
	if err := d.active().Where("user_id = ? AND id IN ?", userID, ids).Find(&list).Error; err != nil {
		return nil, codeerror.OrSystem(err, "CartDao.ListByIDs userID=%d", userID)
	}
	return list, nil
}

func (d *CartDao) SoftDeleteByIDs(userID int64, ids []int64) error {
	if len(ids) == 0 {
		return codeerror.OrSystem(codeerror.ErrorParam, "CartDao.SoftDeleteByIDs empty ids userID=%d", userID)
	}
	res := d.active().Model(&cartmodel.CartItem{}).
		Where("user_id = ? AND id IN ?", userID, ids).
		Updates(map[string]interface{}{
			"deleted": common.DeletedYes,
			"ut":      time.Now().Unix(),
		})
	if res.Error != nil {
		return codeerror.OrSystem(res.Error, "CartDao.SoftDeleteByIDs userID=%d", userID)
	}
	return nil
}
