package userdao

import (
	"gorm.io/gorm"
	"mall/kit/codeerror"
	usermodel "mall/user/model"
)

var fallbackDB *gorm.DB

// SetDB 由宿主服务在自己完成数据库初始化后注入连接。
func SetDB(db *gorm.DB) {
	fallbackDB = db
}

func defaultDB(db *gorm.DB) *gorm.DB {
	if db != nil {
		return db
	}
	return fallbackDB
}

func AutoMigrate() error {
	if err := defaultDB(nil).AutoMigrate(&usermodel.User{}, &usermodel.UserLogin{}); err != nil {
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
