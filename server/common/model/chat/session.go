package chat

import "basic_server/api/chatservice"

type Session struct {
	Sid         string                         `json:"sid"`
	Userid      int64                          `json:"userid"`
	Peer        int64                          `json:"peer"`
	BeginMsgId  int64                          `json:"begin_msg_id"`
	ContentType chatservice.ChatMsgContentType `json:"content_type"` // last msg
	Content     string                         `json:"content"`      // last msg
	Ut          int64                          `json:"ut"`
	Ct          int64                          `json:"ct"`
}

func (Session) TableName() string {
	return "session_info"
}
