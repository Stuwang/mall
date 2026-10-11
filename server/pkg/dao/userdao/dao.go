package userdao

import (
	"gorm.io/gorm"
	"helloworld/base/codeerror"
	"helloworld/common/database"
	usermodel "helloworld/pkg/model/user"
)

func defaultDB(db *gorm.DB) *gorm.DB {
	if db != nil {
		return db
	}
	return database.GetDb()
}

func AutoMigrate() error {
	if err := database.GetDb().AutoMigrate(&usermodel.User{}, &usermodel.UserLogin{}); err != nil {
		return codeerror.OrSystem(err, "AutoMigrate user tables")
	}
	return nil
}

func NewUserDao(db *gorm.DB) *UserDao {
	return &UserDao{db: defaultDB(db)}
}

func NewLoginDao(db *gorm.DB) *LoginDao {
	return &LoginDao{db: defaultDB(db)}
}
