package userdao

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	usermodel "helloworld/pkg/model/user"
	"mall/kit/codeerror"
)

type UserDao struct {
	db *gorm.DB
}

func (d *UserDao) WithTx(tx *gorm.DB) *UserDao {
	return &UserDao{db: tx}
}

type UserQuery struct {
	UserID   int64
	Keyword  string
	Page     int
	PageSize int
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

func (d *UserDao) List(q UserQuery) ([]usermodel.User, int64, error) {
	page, pageSize := q.Page, q.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	tx := d.db.Model(&usermodel.User{})
	if q.UserID > 0 {
		tx = tx.Where("id = ?", q.UserID)
	}
	if kw := strings.TrimSpace(q.Keyword); kw != "" {
		like := "%" + kw + "%"
		sub := d.db.Model(&usermodel.UserLogin{}).
			Select("user_id").
			Where("type = ? AND identifier LIKE ?", usermodel.LoginTypePassword, like)
		tx = tx.Where("nickname LIKE ? OR id IN (?)", like, sub)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, codeerror.OrSystem(err, "UserDao.List count")
	}
	var list []usermodel.User
	if err := tx.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, codeerror.OrSystem(err, "UserDao.List page=%d", page)
	}
	return list, total, nil
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

// SetBanUntil 设置封禁截止时间；banUntil=0 表示解封
func (d *UserDao) SetBanUntil(id, banUntil int64) (*usermodel.User, error) {
	u, err := d.GetByID(id)
	if err != nil {
		return nil, codeerror.OrSystem(err, "UserDao.SetBanUntil GetByID id=%d", id)
	}
	u.BanUntil = banUntil
	u.Ut = time.Now().Unix()
	if err := d.db.Model(u).Select("ban_until", "ut").Updates(u).Error; err != nil {
		return nil, codeerror.OrSystem(err, "UserDao.SetBanUntil id=%d banUntil=%d", id, banUntil)
	}
	return u, nil
}

// SetAdmin 设置/取消管理员标记
func (d *UserDao) SetAdmin(id int64, isAdmin bool) (*usermodel.User, error) {
	u, err := d.GetByID(id)
	if err != nil {
		return nil, codeerror.OrSystem(err, "UserDao.SetAdmin GetByID id=%d", id)
	}
	flag := int32(0)
	if isAdmin {
		flag = 1
	}
	u.IsAdmin = flag
	u.Ut = time.Now().Unix()
	if err := d.db.Model(u).Select("is_admin", "ut").Updates(u).Error; err != nil {
		return nil, codeerror.OrSystem(err, "UserDao.SetAdmin id=%d isAdmin=%v", id, isAdmin)
	}
	return u, nil
}
