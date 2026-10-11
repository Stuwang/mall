package userdao

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"helloworld/base/codeerror"
	"helloworld/common/database"
	usermodel "helloworld/pkg/model/user"
)

func db() *gorm.DB {
	return database.GetDb()
}

// AutoMigrate 迁移用户相关表
func AutoMigrate() error {
	if err := db().AutoMigrate(&usermodel.User{}, &usermodel.UserLogin{}); err != nil {
		return codeerror.OrSystem(err, "AutoMigrate user tables")
	}
	return nil
}

// CreateUserWithPassword 事务创建用户及密码登录方式
func CreateUserWithPassword(nickname, username, password string) (*usermodel.User, error) {
	now := time.Now().Unix()
	u := &usermodel.User{
		Nickname: nickname,
		Sex:      0,
		Ct:       now,
		Ut:       now,
	}
	err := db().Transaction(func(tx *gorm.DB) error {
		var exist usermodel.UserLogin
		err := tx.Where("type = ? AND identifier = ?", usermodel.LoginTypePassword, username).
			First(&exist).Error
		if err == nil {
			return codeerror.ErrorUserExists
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err := tx.Create(u).Error; err != nil {
			return err
		}
		login := &usermodel.UserLogin{
			UserID:     u.ID,
			Type:       usermodel.LoginTypePassword,
			Identifier: username,
			Credential: password,
			Ct:         now,
			Ut:         now,
		}
		if err := tx.Create(login).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, codeerror.OrSystem(err, "CreateUserWithPassword username=%s", username)
	}
	return u, nil
}

// GetLoginByPasswordAccount 按账号密码登录方式查询
func GetLoginByPasswordAccount(username string) (*usermodel.UserLogin, error) {
	var login usermodel.UserLogin
	err := db().Where("type = ? AND identifier = ?", usermodel.LoginTypePassword, username).
		First(&login).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, codeerror.OrSystem(codeerror.ErrorUserNotFound, "GetLoginByPasswordAccount not found username=%s", username)
	}
	if err != nil {
		return nil, codeerror.OrSystem(err, "GetLoginByPasswordAccount username=%s", username)
	}
	return &login, nil
}

// GetByID 按用户 ID 查询
func GetByID(id int64) (*usermodel.User, error) {
	var u usermodel.User
	err := db().Where("id = ?", id).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, codeerror.OrSystem(codeerror.ErrorUserNotFound, "GetByID not found id=%d", id)
	}
	if err != nil {
		return nil, codeerror.OrSystem(err, "GetByID id=%d", id)
	}
	return &u, nil
}

// GetPasswordLoginByUserID 查询用户的密码登录账号
func GetPasswordLoginByUserID(userID int64) (*usermodel.UserLogin, error) {
	var login usermodel.UserLogin
	err := db().Where("user_id = ? AND type = ?", userID, usermodel.LoginTypePassword).
		First(&login).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, codeerror.OrSystem(codeerror.ErrorUserNotFound, "GetPasswordLoginByUserID not found userID=%d", userID)
	}
	if err != nil {
		return nil, codeerror.OrSystem(err, "GetPasswordLoginByUserID userID=%d", userID)
	}
	return &login, nil
}

// UpdateUser 更新用户资料
func UpdateUser(id int64, nickname string, sex int32) (*usermodel.User, error) {
	u, err := GetByID(id)
	if err != nil {
		return nil, codeerror.OrSystem(err, "UpdateUser GetByID id=%d", id)
	}
	u.Nickname = nickname
	u.Sex = sex
	u.Ut = time.Now().Unix()
	if err := db().Model(u).Select("nickname", "sex", "ut").Updates(u).Error; err != nil {
		return nil, codeerror.OrSystem(err, "UpdateUser id=%d", id)
	}
	return u, nil
}
