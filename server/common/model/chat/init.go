package chat

import (
	"basic_server/base/logger"
	"gorm.io/gorm"
)

func InitDb(db *gorm.DB) (err error) {
	err = db.Exec(`
		-- 创建OfflineMsg表
		CREATE TABLE IF NOT EXISTS offline_queue (
			seq INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			userid INTEGER DEFAULT 0,
			'from' INTEGER DEFAULT 0,
			'to' INTEGER DEFAULT 0,
			sender INTEGER DEFAULT 0,
			msg_id INTEGER DEFAULT 0,
			sid VARCHAR(60) DEFAULT '',
			content_type INTEGER DEFAULT 0,
			content TEXT DEFAULT NULL,
			ct INTEGER DEFAULT 0
		);

		-- 为OfflineMsg表创建索引
		CREATE UNIQUE INDEX IF NOT EXISTS idx_offline_queue_userid_seq ON offline_queue (userid, seq);

		`).Error
	if err != nil {
		logger.Panicf("error %v", err)
	}

	err = db.Exec(`
		-- 创建HistoryMsg表
		CREATE TABLE IF NOT EXISTS history_msg (
			sid VARCHAR(60) NOT NULL DEFAULT '0',
			msg_id INTEGER PRIMARY KEY ,
			'from' INTEGER DEFAULT 0,
			'to' INTEGER DEFAULT 0,
			sender INTEGER DEFAULT 0,
			hide_mid INTEGER DEFAULT 0,
			content_type INTEGER DEFAULT 0,
			content TEXT DEFAULT NULL,
			ct INTEGER DEFAULT 0
		);

		-- 为HistoryMsg表创建索引
		CREATE UNIQUE INDEX IF NOT EXISTS idx_history_msg_sid_msgid ON history_msg (sid, msg_id);


		`).Error
	if err != nil {
		logger.Panicf("error %v", err)
	}

	err = db.Exec(`
		-- 创建Session表
		CREATE TABLE IF NOT EXISTS session_info (
			userid INTEGER NOT NULL,
			sid VARCHAR NOT NULL DEFAULT "",
			peer INTEGER DEFAULT 0,
			begin_msg_id INTEGER DEFAULT 0,
			content_type INTEGER DEFAULT 0, -- last msg
			content TEXT DEFAULT NULL, -- last msg
			ut INTEGER DEFAULT 0,
			ct INTEGER DEFAULT 0,
			PRIMARY KEY (userid, sid)
		);

		-- 为Session表创建索引
		CREATE UNIQUE INDEX IF NOT EXISTS idx_session_info_userid_peer ON session_info (userid, peer);
		`).Error
	if err != nil {
		logger.Panicf("error %v", err)
	}
	return
}
