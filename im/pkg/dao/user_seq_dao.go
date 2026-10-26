package imdao

import (
	"time"

	"mall/kit/codeerror"
	"mall/kit/logger"

	"gorm.io/gorm"
)

type UserSeqDao struct {
	db *gorm.DB
}

func NewUserSeqDao(db *gorm.DB) *UserSeqDao {
	return &UserSeqDao{db: defaultDB(db)}
}

func (d *UserSeqDao) WithTx(tx *gorm.DB) *UserSeqDao {
	return &UserSeqDao{db: defaultDB(tx)}
}

// Next 为 userID 分配下一个同步序号。
func (d *UserSeqDao) Next(userID int64) (int64, error) {
	now := time.Now().Unix()
	var seq int64
	err := d.db.Raw(`
INSERT INTO im_user_seq(user_id, seq, ut) VALUES(?, 1, ?)
ON CONFLICT(user_id) DO UPDATE SET seq = seq + 1, ut = excluded.ut
RETURNING seq`, userID, now).Scan(&seq).Error
	if err != nil {
		logger.Errorf("UserSeqDao.Next userID=%d err=%v", userID, err)
		return 0, codeerror.OrSystem(err, "UserSeqDao.Next userID=%d", userID)
	}
	if seq <= 0 {
		logger.Errorf("UserSeqDao.Next empty seq userID=%d", userID)
		return 0, codeerror.ErrorSystem
	}
	return seq, nil
}
