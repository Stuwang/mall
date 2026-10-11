package orderdao

import (
	"time"

	"gorm.io/gorm"
	"helloworld/base/codeerror"
	ordermodel "helloworld/pkg/model/order"
)

type ItemDao struct {
	db *gorm.DB
}

func (d *ItemDao) WithTx(tx *gorm.DB) *ItemDao {
	return &ItemDao{db: tx}
}

func (d *ItemDao) CreateBatch(items []ordermodel.OrderItem) error {
	if len(items) == 0 {
		return codeerror.OrSystem(codeerror.ErrorParam, "ItemDao.CreateBatch empty")
	}
	now := time.Now().Unix()
	for i := range items {
		if items[i].Ct == 0 {
			items[i].Ct = now
		}
	}
	if err := d.db.Create(&items).Error; err != nil {
		return codeerror.OrSystem(err, "ItemDao.CreateBatch count=%d", len(items))
	}
	return nil
}

func (d *ItemDao) ListByOrderID(orderID int64) ([]ordermodel.OrderItem, error) {
	var list []ordermodel.OrderItem
	if err := d.db.Where("order_id = ?", orderID).Order("id asc").Find(&list).Error; err != nil {
		return nil, codeerror.OrSystem(err, "ItemDao.ListByOrderID orderID=%d", orderID)
	}
	return list, nil
}

func (d *ItemDao) ListByOrderIDs(orderIDs []int64) (map[int64][]ordermodel.OrderItem, error) {
	out := make(map[int64][]ordermodel.OrderItem, len(orderIDs))
	if len(orderIDs) == 0 {
		return out, nil
	}
	var list []ordermodel.OrderItem
	if err := d.db.Where("order_id IN ?", orderIDs).Order("id asc").Find(&list).Error; err != nil {
		return nil, codeerror.OrSystem(err, "ItemDao.ListByOrderIDs count=%d", len(orderIDs))
	}
	for i := range list {
		oid := list[i].OrderID
		out[oid] = append(out[oid], list[i])
	}
	return out, nil
}
