package imdao

import (
	"errors"

	"mall/im/pkg/model"
	"mall/kit/codeerror"
	"mall/kit/logger"

	"gorm.io/gorm"
)

type GiftRecordDao struct {
	db *gorm.DB
}

func NewGiftRecordDao(db *gorm.DB) *GiftRecordDao {
	return &GiftRecordDao{db: defaultDB(db)}
}

func (d *GiftRecordDao) WithTx(tx *gorm.DB) *GiftRecordDao {
	return &GiftRecordDao{db: defaultDB(tx)}
}

func (d *GiftRecordDao) GetBySenderClient(senderID int64, clientMsgID string) (*model.GiftRecord, error) {
	var row model.GiftRecord
	err := d.db.Where("sender_id = ? AND client_msg_id = ?", senderID, clientMsgID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		logger.Errorf("GiftRecordDao.GetBySenderClient sender=%d err=%v", senderID, err)
		return nil, codeerror.OrSystem(err, "GiftRecordDao.GetBySenderClient sender=%d", senderID)
	}
	return &row, nil
}

func (d *GiftRecordDao) Create(row *model.GiftRecord) error {
	if err := d.db.Create(row).Error; err != nil {
		logger.Errorf("GiftRecordDao.Create sender=%d err=%v", row.SenderID, err)
		if uniqueConflict(err) {
			return err
		}
		return codeerror.OrSystem(err, "GiftRecordDao.Create sender=%d", row.SenderID)
	}
	return nil
}
