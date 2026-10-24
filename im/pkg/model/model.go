package model

import apicommon "mall/im/api/common"

// Member 一个用户在一场私聊中的视图。不建 im_conversation。
type Member struct {
	ID           int64  `gorm:"column:id;primaryKey;autoIncrement"`
	UserID       int64  `gorm:"column:user_id;uniqueIndex:uk_user_peer;index:idx_user_last_msg,priority:1"`
	PeerUserID   int64  `gorm:"column:peer_user_id;uniqueIndex:uk_user_peer"`
	UserLow      int64  `gorm:"column:user_low"`
	UserHigh     int64  `gorm:"column:user_high"`
	LastSeq      int64  `gorm:"column:last_seq"`
	LastSyncSeq  int64  `gorm:"column:last_sync_seq"`
	LastMsgID    int64  `gorm:"column:last_msg_id"`
	LastSenderID int64  `gorm:"column:last_sender_id"`
	LastPreview  string `gorm:"column:last_preview"`
	LastMsgAt    int64  `gorm:"column:last_msg_at;index:idx_user_last_msg,priority:2"`
	Ct           int64  `gorm:"column:ct"`
	Ut           int64  `gorm:"column:ut"`
}

func (Member) TableName() string {
	return "im_conversation_member"
}

// Message 一条消息。Seq 是同步序号，更新正文时改成更大的新值。
// SeqLow、SeqHigh 是两侧视图序号，创建时写入，之后不改。0 表示该侧不可见。
type Message struct {
	ID             int64                 `gorm:"column:id;primaryKey;autoIncrement"`
	UserLow        int64                 `gorm:"column:user_low;uniqueIndex:uk_pair_seq,priority:1"`
	UserHigh       int64                 `gorm:"column:user_high;uniqueIndex:uk_pair_seq,priority:2"`
	SenderID       int64                 `gorm:"column:sender_id;uniqueIndex:uk_idem,priority:2"`
	Origin         apicommon.Origin      `gorm:"column:origin;uniqueIndex:uk_idem,priority:1"`
	Visibility     apicommon.Visibility  `gorm:"column:visibility"`
	Seq            int64                 `gorm:"column:seq;uniqueIndex:uk_pair_seq,priority:3"`
	SeqLow         int64                 `gorm:"column:seq_low"`
	SeqHigh        int64                 `gorm:"column:seq_high"`
	IdempotencyKey string                `gorm:"column:idempotency_key;size:64;uniqueIndex:uk_idem,priority:3"`
	ContentType    apicommon.ContentType `gorm:"column:content_type"`
	Content        string                `gorm:"column:content;type:text"`
	Ct             int64                 `gorm:"column:ct"`
	Ut             int64                 `gorm:"column:ut"`
}

func (Message) TableName() string {
	return "im_message"
}
