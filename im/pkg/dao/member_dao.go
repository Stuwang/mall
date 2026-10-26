package imdao

import (
	"errors"
	"time"

	"mall/im/pkg/model"
	"mall/kit/codeerror"
	"mall/kit/logger"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MemberDao struct {
	db *gorm.DB
}

func NewMemberDao(db *gorm.DB) *MemberDao {
	return &MemberDao{db: defaultDB(db)}
}

func (d *MemberDao) WithTx(tx *gorm.DB) *MemberDao {
	return &MemberDao{db: defaultDB(tx)}
}

func (d *MemberDao) Get(userID, peerUserID int64) (*model.Member, error) {
	var m model.Member
	err := d.db.Where("user_id = ? AND peer_user_id = ?", userID, peerUserID).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		logger.Errorf("MemberDao.Get userID=%d peer=%d err=%v", userID, peerUserID, err)
		return nil, codeerror.OrSystem(err, "MemberDao.Get userID=%d peer=%d", userID, peerUserID)
	}
	return &m, nil
}

func (d *MemberDao) GetForUpdate(userID, peerUserID int64) (*model.Member, error) {
	var m model.Member
	err := d.db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ? AND peer_user_id = ?", userID, peerUserID).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		logger.Errorf("MemberDao.GetForUpdate userID=%d peer=%d err=%v", userID, peerUserID, err)
		return nil, codeerror.OrSystem(err, "MemberDao.GetForUpdate userID=%d peer=%d", userID, peerUserID)
	}
	return &m, nil
}

// Ensure 保证成员行存在。冲突时读回已有行。
func (d *MemberDao) Ensure(userID, peerUserID int64, userLow, userHigh int64, convID string) (*model.Member, error) {
	now := time.Now().Unix()
	m := &model.Member{
		UserID:     userID,
		PeerUserID: peerUserID,
		UserLow:    userLow,
		UserHigh:   userHigh,
		ConvID:     convID,
		Ct:         now,
		Ut:         now,
	}
	err := d.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "peer_user_id"}},
		DoNothing: true,
	}).Create(m).Error
	if err != nil {
		logger.Errorf("MemberDao.Ensure insert userID=%d peer=%d err=%v", userID, peerUserID, err)
		return nil, codeerror.OrSystem(err, "MemberDao.Ensure userID=%d peer=%d", userID, peerUserID)
	}
	got, err := d.GetForUpdate(userID, peerUserID)
	if err != nil {
		return nil, err
	}
	if got == nil {
		logger.Errorf("MemberDao.Ensure missing after upsert userID=%d peer=%d", userID, peerUserID)
		return nil, codeerror.ErrorSystem
	}
	return got, nil
}

func (d *MemberDao) Save(m *model.Member) error {
	m.Ut = time.Now().Unix()
	if err := d.db.Save(m).Error; err != nil {
		logger.Errorf("MemberDao.Save id=%d userID=%d err=%v", m.ID, m.UserID, err)
		return codeerror.OrSystem(err, "MemberDao.Save id=%d", m.ID)
	}
	return nil
}

func (d *MemberDao) ListAfterSeq(userID, afterSeq int64, limit int) ([]model.Member, error) {
	q := d.db.Where("user_id = ? AND last_seq > ?", userID, afterSeq).
		Order("last_seq ASC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	var list []model.Member
	if err := q.Find(&list).Error; err != nil {
		logger.Errorf("MemberDao.ListAfterSeq userID=%d afterSeq=%d err=%v", userID, afterSeq, err)
		return nil, codeerror.OrSystem(err, "MemberDao.ListAfterSeq userID=%d", userID)
	}
	return list, nil
}
