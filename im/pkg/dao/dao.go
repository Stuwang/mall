package imdao

import (
	"errors"
	"strings"

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
	if err := db.AutoMigrate(
		&model.UserSeq{},
		&model.Member{},
		&model.Message{},
		&model.AlbumPhoto{},
		&model.Post{},
		&model.PostComment{},
		&model.PostLike{},
		&model.UserCoin{},
		&model.CoinLedger{},
		&model.CoinOrder{},
		&model.GiftRecord{},
	); err != nil {
		return codeerror.OrSystem(err, "AutoMigrate im tables")
	}
	// 看不见的一侧序号是 0，多行都可以是 0，所以唯一索引排除 0。
	// 更新时这一侧序号会改成更大的新值，旧值不再占用。
	if err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_msg_seq_low ON im_message(user_low, user_high, seq_low) WHERE seq_low > 0`).Error; err != nil {
		return codeerror.OrSystem(err, "AutoMigrate uk_msg_seq_low")
	}
	if err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_msg_seq_high ON im_message(user_low, user_high, seq_high) WHERE seq_high > 0`).Error; err != nil {
		return codeerror.OrSystem(err, "AutoMigrate uk_msg_seq_high")
	}
	// 空的 client_order_id 不参与唯一约束，避免多个未带幂等键的订单互相冲突。
	if err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_coin_order_client ON im_coin_order(user_id, client_order_id) WHERE client_order_id != ''`).Error; err != nil {
		return codeerror.OrSystem(err, "AutoMigrate uk_coin_order_client")
	}
	return nil
}

// uniqueConflict 判断是否为唯一约束冲突。
// 当前库没有打开 gorm TranslateError，SQLite 会返回 UNIQUE constraint failed。
func uniqueConflict(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "UNIQUE constraint failed") || strings.Contains(msg, "PRIMARY KEY constraint failed")
}
