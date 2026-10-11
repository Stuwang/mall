package user

type User struct {
	Id   int64 `json:"id" gorm:"column:id;primaryKey;autoIncrement;"`
	Sex  int32
	Name string `json:"name" gorm:"column:name;"`
	Ct   int64
	Ut   int64
}

/**
  账号： 一个用户绑定多个账号，例如手机号，邮箱等
 */
type UserAccount struct {
    Id int64 
    Userid int64 
    Type int32
}
