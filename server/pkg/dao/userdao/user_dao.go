package userdao

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"helloworld/base/codeerror"
	usermodel "helloworld/pkg/model/user"
)

type UserDao struct {
	db *gorm.DB
}

func (d *UserDao) WithTx(tx *gorm.DB) *UserDao {
	return &UserDao{db: tx}
}

func (d *UserDao) Create(u *usermodel.User) error {
	if u == nil {
		return codeerror.OrSystem(codeerror.ErrorParam, "UserDao.Create nil user")
	}
	if err := d.db.Create(u).Error; err != nil {
		return codeerror.OrSystem(err, "UserDao.Create")
	}
	return nil
}

func (d *UserDao) GetByID(id int64) (*usermodel.User, error) {
	var u usermodel.User
	err := d.db.Where("id = ?", id).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, codeerror.OrSystem(codeerror.ErrorUserNotFound, "UserDao.GetByID not found id=%d", id)
	}
	if err != nil {
		return nil, codeerror.OrSystem(err, "UserDao.GetByID id=%d", id)
	}
	return &u, nil
}

func (d *UserDao) Update(id int64, nickname string, sex int32) (*usermodel.User, error) {
	u, err := d.GetByID(id)
	if err != nil {
		return nil, codeerror.OrSystem(err, "UserDao.Update GetByID id=%d", id)
	}
	u.Nickname = nickname
	u.Sex = sex
	u.Ut = time.Now().Unix()
	if err := d.db.Model(u).Select("nickname", "sex", "ut").Updates(u).Error; err != nil {
		return nil, codeerror.OrSystem(err, "UserDao.Update id=%d", id)
	}
	return u, nil
}
