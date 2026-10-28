package model

import (
	apicommon "mall/im/api/common"
	kitcommon "mall/kit/model/common"
)

// UserSeq 一个用户的同步序号计数。该用户所有会话共用，只对看得到的变更加一。
type UserSeq struct {
	UserID int64 `gorm:"column:user_id;primaryKey"`
	Seq    int64 `gorm:"column:seq"`
	Ut     int64 `gorm:"column:ut"`
}

func (UserSeq) TableName() string {
	return "im_user_seq"
}

// Member 一个用户在一场私聊中的视图。没有可见消息时不建行。
type Member struct {
	ID           int64  `gorm:"column:id;primaryKey;autoIncrement"`
	UserID       int64  `gorm:"column:user_id;uniqueIndex:uk_user_peer;index:idx_user_seq,priority:1;index:idx_user_sort,priority:1"`
	PeerUserID   int64  `gorm:"column:peer_user_id;uniqueIndex:uk_user_peer"`
	UserLow      int64  `gorm:"column:user_low"`
	UserHigh     int64  `gorm:"column:user_high"`
	ConvID       string `gorm:"column:conv_id;size:64"`
	LastSeq      int64  `gorm:"column:last_seq;index:idx_user_seq,priority:2"`
	LastMsgID    int64  `gorm:"column:last_msg_id"`
	LastSenderID int64  `gorm:"column:last_sender_id"`
	LastPreview  string `gorm:"column:last_preview"`
	SortUt       int64  `gorm:"column:sort_ut;index:idx_user_sort,priority:2"`
	Ct           int64  `gorm:"column:ct"`
	Ut           int64  `gorm:"column:ut"`
}

func (Member) TableName() string {
	return "im_conversation_member"
}

// Message 一条消息。SeqLow、SeqHigh 是两侧各自的同步序号。
// 看不见则为 0。更新正文时，看得到的那一侧改成更大的新序号。
type Message struct {
	ID             int64                 `gorm:"column:id;primaryKey;autoIncrement"`
	UserLow        int64                 `gorm:"column:user_low"`
	UserHigh       int64                 `gorm:"column:user_high"`
	ConvID         string                `gorm:"column:conv_id;size:64"`
	SenderID       int64                 `gorm:"column:sender_id;uniqueIndex:uk_idem,priority:2"`
	Origin         apicommon.Origin      `gorm:"column:origin;uniqueIndex:uk_idem,priority:1"`
	Visibility     apicommon.Visibility  `gorm:"column:visibility"`
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

// AlbumPhoto 用户相册中的一张照片。删除只改 deleted。
type AlbumPhoto struct {
	ID      int64                 `gorm:"column:id;primaryKey;autoIncrement;index:idx_album_user_del_id,priority:3"`
	UserID  int64                 `gorm:"column:user_id;index:idx_album_user_del_id,priority:1"`
	URI     string                `gorm:"column:uri;size:512"`
	Caption string                `gorm:"column:caption;size:320"`
	Deleted kitcommon.DeletedFlag `gorm:"column:deleted;index:idx_album_user_del_id,priority:2"`
	Ct      int64                 `gorm:"column:ct"`
	Ut      int64                 `gorm:"column:ut"`
}

func (AlbumPhoto) TableName() string {
	return "im_album_photo"
}

// Post 一条帖子。删除只改 deleted。计数只由服务端改。
// Images 存逻辑 URI 的 JSON 数组，如 ["local_file:///a.jpg"]。
type Post struct {
	ID             int64                 `gorm:"column:id;primaryKey;autoIncrement;index:idx_post_rec,priority:3"`
	UserID         int64                 `gorm:"column:user_id;uniqueIndex:uk_post_idem,priority:1"`
	ClientPostID   string                `gorm:"column:client_post_id;size:64;uniqueIndex:uk_post_idem,priority:2"`
	Text           string                `gorm:"column:text;type:text"`
	Images         string                `gorm:"column:images;type:text"`
	LikeCount      int64                 `gorm:"column:like_count"`
	CommentCount   int64                 `gorm:"column:comment_count"`
	RecommendScore int64                 `gorm:"column:recommend_score;index:idx_post_rec,priority:2"`
	Deleted        kitcommon.DeletedFlag `gorm:"column:deleted;index:idx_post_rec,priority:1"`
	Ct             int64                 `gorm:"column:ct"`
	Ut             int64                 `gorm:"column:ut"`
}

func (Post) TableName() string {
	return "im_post"
}

// PostComment 帖子下的平级评论。
type PostComment struct {
	ID              int64                 `gorm:"column:id;primaryKey;autoIncrement;index:idx_cmt_post,priority:3"`
	PostID          int64                 `gorm:"column:post_id;uniqueIndex:uk_cmt_idem,priority:1;index:idx_cmt_post,priority:1"`
	UserID          int64                 `gorm:"column:user_id;uniqueIndex:uk_cmt_idem,priority:2"`
	ClientCommentID string                `gorm:"column:client_comment_id;size:64;uniqueIndex:uk_cmt_idem,priority:3"`
	Text            string                `gorm:"column:text;type:text"`
	Deleted         kitcommon.DeletedFlag `gorm:"column:deleted;index:idx_cmt_post,priority:2"`
	Ct              int64                 `gorm:"column:ct"`
	Ut              int64                 `gorm:"column:ut"`
}

func (PostComment) TableName() string {
	return "im_post_comment"
}

// PostLike 一人一帖一行。取消后再次点赞复活。
type PostLike struct {
	ID      int64                 `gorm:"column:id;primaryKey;autoIncrement"`
	PostID  int64                 `gorm:"column:post_id;uniqueIndex:uk_like,priority:1"`
	UserID  int64                 `gorm:"column:user_id;uniqueIndex:uk_like,priority:2"`
	Deleted kitcommon.DeletedFlag `gorm:"column:deleted"`
	Ct      int64                 `gorm:"column:ct"`
	Ut      int64                 `gorm:"column:ut"`
}

func (PostLike) TableName() string {
	return "im_post_like"
}

// UserCoin 用户金币余额。没有行表示余额为 0。
type UserCoin struct {
	UserID  int64 `gorm:"column:user_id;primaryKey"`
	Balance int64 `gorm:"column:balance"`
	Ct      int64 `gorm:"column:ct"`
	Ut      int64 `gorm:"column:ut"`
}

func (UserCoin) TableName() string {
	return "im_user_coin"
}

// CoinLedger 金币流水。idempotent_id 唯一，重复请求不重复入账。
type CoinLedger struct {
	ID           int64                    `gorm:"column:id;primaryKey;autoIncrement;index:idx_coin_ledger_user,priority:2"`
	UserID       int64                    `gorm:"column:user_id;index:idx_coin_ledger_user,priority:1"`
	ChangeType   apicommon.CoinChangeType `gorm:"column:change_type"`
	Change       int64                    `gorm:"column:change"`
	Balance      int64                    `gorm:"column:balance"`
	IdempotentID string                   `gorm:"column:idempotent_id;size:128;uniqueIndex"`
	Ext          string                   `gorm:"column:ext;type:text"`
	Remark       string                   `gorm:"column:remark;size:128"`
	Ct           int64                    `gorm:"column:ct"`
}

func (CoinLedger) TableName() string {
	return "im_coin_ledger"
}

// CoinOrder 金币充值订单。client_order_id 为空时不参与唯一约束。
type CoinOrder struct {
	ID            int64                     `gorm:"column:id;primaryKey;autoIncrement"`
	OrderNo       string                    `gorm:"column:order_no;size:64;uniqueIndex"`
	UserID        int64                     `gorm:"column:user_id;index"`
	ClientOrderID string                    `gorm:"column:client_order_id;size:64"`
	Coins         int64                     `gorm:"column:coins"`
	CashAmount    int64                     `gorm:"column:cash_amount"`
	CoinsPerYuan  int64                     `gorm:"column:coins_per_yuan"`
	PayType       apicommon.PayType         `gorm:"column:pay_type"`
	Status        apicommon.CoinOrderStatus `gorm:"column:status"`
	ExpireAt      int64                     `gorm:"column:expire_at"`
	PaidAt        int64                     `gorm:"column:paid_at"`
	Ct            int64                     `gorm:"column:ct"`
	Ut            int64                     `gorm:"column:ut"`
}

func (CoinOrder) TableName() string {
	return "im_coin_order"
}

// GiftRecord 一次送礼。同一发送者的 client_msg_id 只成功一次。
type GiftRecord struct {
	ID          int64  `gorm:"column:id;primaryKey;autoIncrement"`
	SenderID    int64  `gorm:"column:sender_id;uniqueIndex:uk_gift_client,priority:1"`
	ClientMsgID string `gorm:"column:client_msg_id;size:64;uniqueIndex:uk_gift_client,priority:2"`
	PeerUserID  int64  `gorm:"column:peer_user_id"`
	GiftID      int64  `gorm:"column:gift_id"`
	GiftName    string `gorm:"column:gift_name;size:64"`
	GiftIcon    string `gorm:"column:gift_icon;size:512"`
	GiftEffect  string `gorm:"column:gift_effect;size:128"`
	Price       int64  `gorm:"column:price"`
	MsgID       int64  `gorm:"column:msg_id"`
	Ct          int64  `gorm:"column:ct"`
}

func (GiftRecord) TableName() string {
	return "im_gift_record"
}
