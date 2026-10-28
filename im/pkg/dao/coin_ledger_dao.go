package imdao

import (
	"errors"

	"mall/im/pkg/model"
	"mall/kit/codeerror"
	"mall/kit/logger"

	"gorm.io/gorm"
)

type CoinLedgerDao struct {
	db *gorm.DB
}

func NewCoinLedgerDao(db *gorm.DB) *CoinLedgerDao {
	return &CoinLedgerDao{db: defaultDB(db)}
}

func (d *CoinLedgerDao) WithTx(tx *gorm.DB) *CoinLedgerDao {
	return &CoinLedgerDao{db: defaultDB(tx)}
}

func (d *CoinLedgerDao) GetByIdempotentID(idempotentID string) (*model.CoinLedger, error) {
	var row model.CoinLedger
	err := d.db.Where("idempotent_id = ?", idempotentID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		logger.Errorf("CoinLedgerDao.GetByIdempotentID idempotent=%s err=%v", idempotentID, err)
		return nil, codeerror.OrSystem(err, "CoinLedgerDao.GetByIdempotentID")
	}
	return &row, nil
}

func (d *CoinLedgerDao) Create(row *model.CoinLedger) error {
	if err := d.db.Create(row).Error; err != nil {
		logger.Errorf("CoinLedgerDao.Create user=%d idempotent=%s err=%v", row.UserID, row.IdempotentID, err)
		if uniqueConflict(err) {
			return err
		}
		return codeerror.OrSystem(err, "CoinLedgerDao.Create user=%d", row.UserID)
	}
	return nil
}

// ListDesc 按 id 倒序。beforeID 大于 0 时只取更小的 id。
func (d *CoinLedgerDao) ListDesc(userID, beforeID int64, limit int) ([]model.CoinLedger, error) {
	q := d.db.Where("user_id = ?", userID)
	if beforeID > 0 {
		q = q.Where("id < ?", beforeID)
	}
	if limit > 0 {
		q = q.Limit(limit)
	}
	list := []model.CoinLedger{}
	if err := q.Order("id DESC").Find(&list).Error; err != nil {
		logger.Errorf("CoinLedgerDao.ListDesc user=%d err=%v", userID, err)
		return nil, codeerror.OrSystem(err, "CoinLedgerDao.ListDesc user=%d", userID)
	}
	return list, nil
}
