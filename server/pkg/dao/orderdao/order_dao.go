package orderdao

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"helloworld/base/codeerror"
	ordermodel "helloworld/pkg/model/order"
)

type OrderDao struct {
	db *gorm.DB
}

func (d *OrderDao) WithTx(tx *gorm.DB) *OrderDao {
	return &OrderDao{db: tx}
}

func (d *OrderDao) Create(o *ordermodel.Order) error {
	if o == nil || o.UserID <= 0 || o.OrderNo == "" {
		return codeerror.OrSystem(codeerror.ErrorParam, "OrderDao.Create invalid order")
	}
	now := time.Now().Unix()
	if o.Ct == 0 {
		o.Ct = now
	}
	o.Ut = now
	if err := d.db.Create(o).Error; err != nil {
		return codeerror.OrSystem(err, "OrderDao.Create orderNo=%s", o.OrderNo)
	}
	return nil
}

func (d *OrderDao) GetByID(id, userID int64) (*ordermodel.Order, error) {
	var o ordermodel.Order
	q := d.db.Where("id = ?", id)
	if userID > 0 {
		q = q.Where("user_id = ?", userID)
	}
	err := q.First(&o).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, codeerror.OrSystem(codeerror.ErrorParam.Errorf("order not found"), "OrderDao.GetByID not found id=%d userID=%d", id, userID)
	}
	if err != nil {
		return nil, codeerror.OrSystem(err, "OrderDao.GetByID id=%d userID=%d", id, userID)
	}
	return &o, nil
}

func (d *OrderDao) Update(o *ordermodel.Order, fields ...string) error {
	if o == nil || o.ID <= 0 {
		return codeerror.OrSystem(codeerror.ErrorParam, "OrderDao.Update invalid order")
	}
	o.Ut = time.Now().Unix()
	cols := append(append([]string{}, fields...), "ut")
	if err := d.db.Model(o).Select(cols).Updates(o).Error; err != nil {
		return codeerror.OrSystem(err, "OrderDao.Update id=%d", o.ID)
	}
	return nil
}

type ListQuery struct {
	UserID   int64
	Status   ordermodel.OrderStatus // 0=全部
	Page     int
	PageSize int
}

func (d *OrderDao) List(q ListQuery) ([]ordermodel.Order, int64, error) {
	page, pageSize := q.Page, q.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	tx := d.db.Model(&ordermodel.Order{}).Where("user_id = ?", q.UserID)
	if q.Status > 0 {
		tx = tx.Where("status = ?", q.Status)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, codeerror.OrSystem(err, "OrderDao.List count userID=%d", q.UserID)
	}
	var list []ordermodel.Order
	if err := tx.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, codeerror.OrSystem(err, "OrderDao.List userID=%d", q.UserID)
	}
	return list, total, nil
}

// ListExpiredPending 查询已过期仍待支付的订单
func (d *OrderDao) ListExpiredPending(limit int, now int64) ([]ordermodel.Order, error) {
	if limit <= 0 {
		limit = 100
	}
	var list []ordermodel.Order
	err := d.db.Where("status = ? AND pay_expire_at > 0 AND pay_expire_at < ?", ordermodel.OrderStatusPendingPay, now).
		Order("id asc").Limit(limit).Find(&list).Error
	if err != nil {
		return nil, codeerror.OrSystem(err, "OrderDao.ListExpiredPending")
	}
	return list, nil
}
