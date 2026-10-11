package userdao

import (
	"time"

	"gorm.io/gorm"
	"helloworld/base/codeerror"
	"helloworld/common/database"
	usermodel "helloworld/pkg/model/user"
)

// CreateUserWithPassword 跨 user / user_login 事务注册
func CreateUserWithPassword(nickname, username, password string) (*usermodel.User, error) {
	now := time.Now().Unix()
	u := &usermodel.User{
		Nickname: nickname,
		Sex:      0,
		Ct:       now,
		Ut:       now,
	}
	err := database.GetDb().Transaction(func(tx *gorm.DB) error {
		userDao := NewUserDao(tx)
		loginDao := NewLoginDao(tx)

		exists, err := loginDao.ExistsPasswordAccount(username)
		if err != nil {
			return err
		}
		if exists {
			return codeerror.ErrorUserExists
		}
		if err := userDao.Create(u); err != nil {
			return err
		}
		return loginDao.Create(&usermodel.UserLogin{
			UserID:     u.ID,
			Type:       usermodel.LoginTypePassword,
			Identifier: username,
			Credential: password,
			Ct:         now,
			Ut:         now,
		})
	})
	if err != nil {
		return nil, codeerror.OrSystem(err, "CreateUserWithPassword username=%s", username)
	}
	return u, nil
}
