package userdao

import (
	"strings"
	"time"

	"gorm.io/gorm"
	"helloworld/base/codeerror"
	"helloworld/base/logger"
	"helloworld/common/config"
	"helloworld/common/database"
	usermodel "helloworld/pkg/model/user"
)

// EnsureAdmins 按配置幂等写入管理员：username 已存在则跳过创建（必要时补 is_admin）
func EnsureAdmins(accounts []config.AdminAccount) error {
	for i, acc := range accounts {
		username := strings.TrimSpace(acc.Username)
		password := acc.Password
		nickname := strings.TrimSpace(acc.Nickname)
		if username == "" || password == "" {
			return codeerror.OrSystem(codeerror.ErrorParam, "EnsureAdmins invalid account index=%d", i)
		}
		if nickname == "" {
			nickname = username
		}
		created, err := ensureAdminAccount(username, password, nickname)
		if err != nil {
			return codeerror.OrSystem(err, "EnsureAdmins username=%s", username)
		}
		if created {
			logger.Infof("EnsureAdmins created admin username=%s", username)
		} else {
			logger.Infof("EnsureAdmins skip existing username=%s", username)
		}
	}
	return nil
}

func ensureAdminAccount(username, password, nickname string) (created bool, err error) {
	loginDao := NewLoginDao(nil)
	login, err := loginDao.GetByPasswordAccount(username)
	if err == nil {
		// 已存在：不改密码，确保具备管理员权限
		u, gerr := NewUserDao(nil).GetByID(login.UserID)
		if gerr != nil {
			return false, gerr
		}
		if !u.Admin() {
			if _, serr := NewUserDao(nil).SetAdmin(u.ID, true); serr != nil {
				return false, serr
			}
			logger.Infof("EnsureAdmins promote to admin userID=%d username=%s", u.ID, username)
		}
		return false, nil
	}
	if ce, ok := codeerror.As(err); !ok || !ce.Is(codeerror.ErrorUserNotFound) {
		return false, err
	}

	now := time.Now().Unix()
	u := &usermodel.User{
		Nickname: nickname,
		Sex:      0,
		IsAdmin:  1,
		Ct:       now,
		Ut:       now,
	}
	txErr := database.GetDb().Transaction(func(tx *gorm.DB) error {
		userDao := NewUserDao(tx)
		loginDaoTx := NewLoginDao(tx)
		exists, e := loginDaoTx.ExistsPasswordAccount(username)
		if e != nil {
			return e
		}
		if exists {
			// 并发下可能已被写入，视为已存在
			created = false
			return nil
		}
		if e = userDao.Create(u); e != nil {
			return e
		}
		if e = loginDaoTx.Create(&usermodel.UserLogin{
			UserID:     u.ID,
			Type:       usermodel.LoginTypePassword,
			Identifier: username,
			Credential: password,
			Ct:         now,
			Ut:         now,
		}); e != nil {
			return e
		}
		created = true
		return nil
	})
	if txErr != nil {
		return false, codeerror.OrSystem(txErr, "ensureAdminAccount username=%s", username)
	}
	return created, nil
}
