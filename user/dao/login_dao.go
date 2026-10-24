package userdao

import (
	"errors"

	"gorm.io/gorm"
	"mall/kit/codeerror"
	usermodel "mall/user/model"
)

type LoginDao struct {
	db *gorm.DB
}

func (d *LoginDao) WithTx(tx *gorm.DB) *LoginDao {
	return &LoginDao{db: tx}
}

func (d *LoginDao) Create(login *usermodel.UserLogin) error {
	if login == nil {
		return codeerror.OrSystem(codeerror.ErrorParam, "LoginDao.Create nil login")
	}
	if err := d.db.Create(login).Error; err != nil {
		return codeerror.OrSystem(err, "LoginDao.Create userID=%d", login.UserID)
	}
	return nil
}

func (d *LoginDao) GetByPasswordAccount(username string) (*usermodel.UserLogin, error) {
	var login usermodel.UserLogin
	err := d.db.Where("type = ? AND identifier = ?", usermodel.LoginTypePassword, username).
		First(&login).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, codeerror.OrSystem(codeerror.ErrorUserNotFound, "LoginDao.GetByPasswordAccount not found username=%s", username)
	}
	if err != nil {
		return nil, codeerror.OrSystem(err, "LoginDao.GetByPasswordAccount username=%s", username)
	}
	return &login, nil
}

func (d *LoginDao) GetPasswordByUserID(userID int64) (*usermodel.UserLogin, error) {
	var login usermodel.UserLogin
	err := d.db.Where("user_id = ? AND type = ?", userID, usermodel.LoginTypePassword).
		First(&login).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, codeerror.OrSystem(codeerror.ErrorUserNotFound, "LoginDao.GetPasswordByUserID not found userID=%d", userID)
	}
	if err != nil {
		return nil, codeerror.OrSystem(err, "LoginDao.GetPasswordByUserID userID=%d", userID)
	}
	return &login, nil
}

func (d *LoginDao) ExistsPasswordAccount(username string) (bool, error) {
	var login usermodel.UserLogin
	err := d.db.Where("type = ? AND identifier = ?", usermodel.LoginTypePassword, username).
		First(&login).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, codeerror.OrSystem(err, "LoginDao.ExistsPasswordAccount username=%s", username)
	}
	return true, nil
}

// MapPasswordUsernameByUserIDs 批量查密码账号用户名
func (d *LoginDao) MapPasswordUsernameByUserIDs(userIDs []int64) (map[int64]string, error) {
	out := make(map[int64]string, len(userIDs))
	if len(userIDs) == 0 {
		return out, nil
	}
	var list []usermodel.UserLogin
	err := d.db.Where("user_id IN ? AND type = ?", userIDs, usermodel.LoginTypePassword).
		Find(&list).Error
	if err != nil {
		return nil, codeerror.OrSystem(err, "LoginDao.MapPasswordUsernameByUserIDs count=%d", len(userIDs))
	}
	for i := range list {
		out[list[i].UserID] = list[i].Identifier
	}
	return out, nil
}
