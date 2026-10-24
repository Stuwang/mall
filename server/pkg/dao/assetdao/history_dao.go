package assetdao

import (
	"errors"

	"gorm.io/gorm"
	assetmodel "helloworld/pkg/model/asset"
	"mall/kit/codeerror"
)

type HistoryDao struct {
	db *gorm.DB
}

func (d *HistoryDao) WithTx(tx *gorm.DB) *HistoryDao {
	return &HistoryDao{db: tx}
}

func (d *HistoryDao) GetByIdempotentID(idempotentID string) (*assetmodel.UserAssetHistory, error) {
	var h assetmodel.UserAssetHistory
	err := d.db.Where("idempotent_id = ?", idempotentID).First(&h).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, codeerror.OrSystem(err, "HistoryDao.GetByIdempotentID id=%s", idempotentID)
	}
	return &h, nil
}

func (d *HistoryDao) Create(h *assetmodel.UserAssetHistory) error {
	if h == nil {
		return codeerror.OrSystem(codeerror.ErrorParam, "HistoryDao.Create nil")
	}
	if err := d.db.Create(h).Error; err != nil {
		return codeerror.OrSystem(err, "HistoryDao.Create userID=%d idempotentID=%s", h.UserID, h.IdempotentID)
	}
	return nil
}

func (d *HistoryDao) List(userID int64, assetType int32, page, pageSize int) ([]assetmodel.UserAssetHistory, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	q := d.db.Model(&assetmodel.UserAssetHistory{}).Where("user_id = ?", userID)
	if assetType > 0 {
		q = q.Where("type = ?", assetType)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, codeerror.OrSystem(err, "HistoryDao.List count userID=%d type=%d", userID, assetType)
	}

	var list []assetmodel.UserAssetHistory
	err := q.Order("id desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&list).Error
	if err != nil {
		return nil, 0, codeerror.OrSystem(err, "HistoryDao.List userID=%d type=%d page=%d", userID, assetType, page)
	}
	return list, total, nil
}
