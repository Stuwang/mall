package imdao

import (
	"mall/im/common/database"
	"mall/im/pkg/model"
	"mall/kit/codeerror"

	"gorm.io/gorm"
)

func defaultDB(db *gorm.DB) *gorm.DB {
	if db != nil {
		return db
	}
	return database.GetDb()
}

func AutoMigrate() error {
	db := defaultDB(nil)
	if err := db.AutoMigrate(&model.Member{}, &model.Message{}); err != nil {
		return codeerror.OrSystem(err, "AutoMigrate im tables")
	}
	// 视图序号 0 表示不可见，多行都可以是 0，所以唯一索引排除 0。
	if err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_msg_seq_low ON im_message(user_low, user_high, seq_low) WHERE seq_low > 0`).Error; err != nil {
		return codeerror.OrSystem(err, "AutoMigrate uk_msg_seq_low")
	}
	if err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_msg_seq_high ON im_message(user_low, user_high, seq_high) WHERE seq_high > 0`).Error; err != nil {
		return codeerror.OrSystem(err, "AutoMigrate uk_msg_seq_high")
	}
	return nil
}

func NewMemberDao(db *gorm.DB) *MemberDao {
	return &MemberDao{db: defaultDB(db)}
}

func NewMessageDao(db *gorm.DB) *MessageDao {
	return &MessageDao{db: defaultDB(db)}
}
