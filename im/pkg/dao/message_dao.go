package imdao

import (
	"errors"
	"fmt"

	apicommon "mall/im/api/common"
	"mall/im/pkg/model"
	"mall/kit/codeerror"
	"mall/kit/logger"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MessageDao struct {
	db *gorm.DB
}

func NewMessageDao(db *gorm.DB) *MessageDao {
	return &MessageDao{db: defaultDB(db)}
}

func (d *MessageDao) WithTx(tx *gorm.DB) *MessageDao {
	return &MessageDao{db: defaultDB(tx)}
}

func (d *MessageDao) GetByID(id int64) (*model.Message, error) {
	var m model.Message
	err := d.db.Where("id = ?", id).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		logger.Errorf("MessageDao.GetByID id=%d err=%v", id, err)
		return nil, codeerror.OrSystem(err, "MessageDao.GetByID id=%d", id)
	}
	return &m, nil
}

func (d *MessageDao) GetByIDForUpdate(id int64) (*model.Message, error) {
	var m model.Message
	err := d.db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		logger.Errorf("MessageDao.GetByIDForUpdate id=%d err=%v", id, err)
		return nil, codeerror.OrSystem(err, "MessageDao.GetByIDForUpdate id=%d", id)
	}
	return &m, nil
}

func (d *MessageDao) GetByIdempotency(origin apicommon.Origin, senderID int64, key string) (*model.Message, error) {
	var m model.Message
	err := d.db.Where("origin = ? AND sender_id = ? AND idempotency_key = ?", origin, senderID, key).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		logger.Errorf("MessageDao.GetByIdempotency origin=%d sender=%d err=%v", origin, senderID, err)
		return nil, codeerror.OrSystem(err, "MessageDao.GetByIdempotency sender=%d", senderID)
	}
	return &m, nil
}

func (d *MessageDao) Create(m *model.Message) error {
	if err := d.db.Create(m).Error; err != nil {
		logger.Errorf("MessageDao.Create sender=%d conv=%s err=%v", m.SenderID, m.ConvID, err)
		return codeerror.OrSystem(err, "MessageDao.Create")
	}
	return nil
}

func (d *MessageDao) Save(m *model.Message) error {
	if err := d.db.Save(m).Error; err != nil {
		logger.Errorf("MessageDao.Save id=%d err=%v", m.ID, err)
		return codeerror.OrSystem(err, "MessageDao.Save id=%d", m.ID)
	}
	return nil
}

// ListAfterSeq 按当前用户这一侧的同步序号增量拉取。
func (d *MessageDao) ListAfterSeq(userLow, userHigh, viewerID, afterSeq int64, limit int) ([]model.Message, error) {
	col, err := viewerSeqColumn(userLow, userHigh, viewerID)
	if err != nil {
		logger.Errorf("MessageDao.ListAfterSeq bad viewer userLow=%d userHigh=%d viewer=%d", userLow, userHigh, viewerID)
		return nil, codeerror.OrSystem(err, "MessageDao.ListAfterSeq viewer=%d", viewerID)
	}
	q := d.db.Where("user_low = ? AND user_high = ?", userLow, userHigh).
		Where(fmt.Sprintf("%s > ? AND %s > 0", col, col), afterSeq).
		Order(col + " ASC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	var list []model.Message
	if err := q.Find(&list).Error; err != nil {
		logger.Errorf("MessageDao.ListAfterSeq pair=%d_%d viewer=%d err=%v", userLow, userHigh, viewerID, err)
		return nil, codeerror.OrSystem(err, "MessageDao.ListAfterSeq viewer=%d", viewerID)
	}
	return list, nil
}

func viewerSeqColumn(userLow, userHigh, viewerID int64) (string, error) {
	switch viewerID {
	case userLow:
		return "seq_low", nil
	case userHigh:
		return "seq_high", nil
	default:
		return "", fmt.Errorf("viewer %d not in pair %d_%d", viewerID, userLow, userHigh)
	}
}
