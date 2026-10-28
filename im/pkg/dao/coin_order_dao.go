package imdao

import (
	"errors"

	"mall/im/pkg/model"
	"mall/kit/codeerror"
	"mall/kit/logger"

	"gorm.io/gorm"
)

type CoinOrderDao struct {
	db *gorm.DB
}

func NewCoinOrderDao(db *gorm.DB) *CoinOrderDao {
	return &CoinOrderDao{db: defaultDB(db)}
}

func (d *CoinOrderDao) WithTx(tx *gorm.DB) *CoinOrderDao {
	return &CoinOrderDao{db: defaultDB(tx)}
}

func (d *CoinOrderDao) GetByID(id int64) (*model.CoinOrder, error) {
	var row model.CoinOrder
	err := d.db.Where("id = ?", id).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		logger.Errorf("CoinOrderDao.GetByID id=%d err=%v", id, err)
		return nil, codeerror.OrSystem(err, "CoinOrderDao.GetByID id=%d", id)
	}
	return &row, nil
}

func (d *CoinOrderDao) GetByClient(userID int64, clientOrderID string) (*model.CoinOrder, error) {
	var row model.CoinOrder
	err := d.db.Where("user_id = ? AND client_order_id = ?", userID, clientOrderID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		logger.Errorf("CoinOrderDao.GetByClient user=%d err=%v", userID, err)
		return nil, codeerror.OrSystem(err, "CoinOrderDao.GetByClient user=%d", userID)
	}
	return &row, nil
}

func (d *CoinOrderDao) Create(row *model.CoinOrder) error {
	if err := d.db.Create(row).Error; err != nil {
		logger.Errorf("CoinOrderDao.Create user=%d order=%s err=%v", row.UserID, row.OrderNo, err)
		if uniqueConflict(err) {
			return err
		}
		return codeerror.OrSystem(err, "CoinOrderDao.Create user=%d", row.UserID)
	}
	return nil
}

func (d *CoinOrderDao) Save(row *model.CoinOrder) error {
	if err := d.db.Save(row).Error; err != nil {
		logger.Errorf("CoinOrderDao.Save id=%d err=%v", row.ID, err)
		return codeerror.OrSystem(err, "CoinOrderDao.Save id=%d", row.ID)
	}
	return nil
}
