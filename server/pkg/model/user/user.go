package user

// LoginTypePassword 账号密码登录
const LoginTypePassword int32 = 1

// User 用户表
type User struct {
	ID       int64  `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Nickname string `json:"nickname" gorm:"column:nickname"`
	Sex      int32  `json:"sex" gorm:"column:sex"`
	Ct       int64  `json:"ct" gorm:"column:ct"`
	Ut       int64  `json:"ut" gorm:"column:ut"`
}

func (User) TableName() string {
	return "user"
}

// UserLogin 用户登录方式表（一个用户可绑定多种登录方式）
type UserLogin struct {
	ID         int64  `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	UserID     int64  `json:"user_id" gorm:"column:user_id;index"`
	Type       int32  `json:"type" gorm:"column:type;uniqueIndex:uk_type_identifier"`
	Identifier string `json:"identifier" gorm:"column:identifier;uniqueIndex:uk_type_identifier"`
	Credential string `json:"credential" gorm:"column:credential"`
	Ct         int64  `json:"ct" gorm:"column:ct"`
	Ut         int64  `json:"ut" gorm:"column:ut"`
}

func (UserLogin) TableName() string {
	return "user_login"
}
