package user

import "time"

// LoginType 登录方式
type LoginType int32

const (
	LoginTypePassword LoginType = 1 // 账号密码
)

// User 用户表
type User struct {
	ID       int64  `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Nickname string `json:"nickname" gorm:"column:nickname"`
	Sex      int32  `json:"sex" gorm:"column:sex"`
	// Avatar 头像逻辑 URI（如 local_file:///key），空表示未设置
	Avatar  string `json:"avatar" gorm:"column:avatar"`
	IsAdmin int32  `json:"is_admin" gorm:"column:is_admin"` // 1=管理员
	// BanUntil 封禁截止 unix 时间戳；0 表示未封禁；到期后自动视为解封
	BanUntil int64 `json:"ban_until" gorm:"column:ban_until"`
	Ct       int64 `json:"ct" gorm:"column:ct"`
	Ut       int64 `json:"ut" gorm:"column:ut"`
}

func (u *User) Admin() bool {
	return u != nil && u.IsAdmin == 1
}

// Banned 是否处于封禁期内
func (u *User) Banned() bool {
	return u != nil && u.BanUntil > 0 && u.BanUntil > time.Now().Unix()
}

func (User) TableName() string {
	return "user"
}

// UserLogin 用户登录方式表（一个用户可绑定多种登录方式）
type UserLogin struct {
	ID         int64     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	UserID     int64     `json:"user_id" gorm:"column:user_id;index"`
	Type       LoginType `json:"type" gorm:"column:type;uniqueIndex:uk_type_identifier"`
	Identifier string    `json:"identifier" gorm:"column:identifier;uniqueIndex:uk_type_identifier"`
	Credential string    `json:"credential" gorm:"column:credential"`
	Ct         int64     `json:"ct" gorm:"column:ct"`
	Ut         int64     `json:"ut" gorm:"column:ut"`
}

func (UserLogin) TableName() string {
	return "user_login"
}
