package imdao

import (
	"errors"

	"mall/im/pkg/model"
	"mall/kit/codeerror"
	"mall/kit/logger"

	"gorm.io/gorm"
)

type UserCoinDao struct {
	db *gorm.DB
}

func NewUserCoinDao(db *gorm.DB) *UserCoinDao {
	return &UserCoinDao{db: defaultDB(db)}
}

func (d *UserCoinDao) WithTx(tx *gorm.DB) *UserCoinDao {
	return &UserCoinDao{db: defaultDB(tx)}
}

func (d *UserCoinDao) Get(userID int64) (*model.UserCoin, error) {
	var row model.UserCoin
	err := d.db.Where("user_id = ?", userID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		logger.Errorf("UserCoinDao.Get user=%d err=%v", userID, err)
		return nil, codeerror.OrSystem(err, "UserCoinDao.Get user=%d", userID)
	}
	return &row, nil
}

func (d *UserCoinDao) Create(row *model.UserCoin) error {
	if err := d.db.Create(row).Error; err != nil {
		logger.Errorf("UserCoinDao.Create user=%d err=%v", row.UserID, err)
		if uniqueConflict(err) {
			return err
		}
		return codeerror.OrSystem(err, "UserCoinDao.Create user=%d", row.UserID)
	}
	return nil
}

// AddBalance 在余额足够时加上 delta（可负）。影响行数为 0 表示余额不足或没有账户行。
func (d *UserCoinDao) AddBalance(userID, delta, now int64) (int64, error) {
	res := d.db.Exec(
		`UPDATE im_user_coin SET balance = balance + ?, ut = ? WHERE user_id = ? AND balance + ? >= 0`,
		delta,
		now,
		userID,
		delta,
	)
	if res.Error != nil {
		logger.Errorf("UserCoinDao.AddBalance user=%d delta=%d err=%v", userID, delta, res.Error)
		return 0, codeerror.OrSystem(res.Error, "UserCoinDao.AddBalance user=%d", userID)
	}
	return res.RowsAffected, nil
}
