package chat

import (
	"basic_server/api/chatservice"
	"fmt"
)

type OfflineMsg struct {
	Userid      int64                          `json:"userid"`
	Seq         int64                          `json:"seq"`
	From        int64                          `json:"from"`
	To          int64                          `json:"to"`
	Sender      int64                          `json:"sender"`
	MsgId       int64                          `json:"msg_id"`
	Sid         string                         `json:"sid"`
	ContentType chatservice.ChatMsgContentType `json:"content_type"`
	Content     string                         `json:"content"`
	Ct          int64                          `json:"ct"`
}

func (o OfflineMsg) TableName() string {
	return "offline_queue"
}

type HistoryMsg struct {
	Sid         string                         `json:"sid"`
	MsgId       int64                          `json:"msg_id"`
	From        int64                          `json:"from"`
	To          int64                          `json:"to"`
	Sender      int64                          `json:"sender"`
	HideMid     int64                          `json:"hide_mid"` // 消息隐藏的人
	ContentType chatservice.ChatMsgContentType `json:"content_type"`
	Content     string                         `json:"content"`
	Ct          int64                          `json:"ct"`
}

func SessionId(userid1, userid2 int64) string {
	s, b := userid1, userid2
	if s > b {
		s, b = b, s
	}
	return fmt.Sprintf("%d:%d", s, b)
}

func (HistoryMsg) TableName() string {
	return "history_msg"
}
