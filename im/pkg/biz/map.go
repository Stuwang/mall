package biz

import (
	"encoding/json"

	apicommon "mall/im/api/common"
	api "mall/im/api/im"
	"mall/im/pkg/model"
	"mall/kit/logger"
)

type textContent struct {
	Text string `json:"text"`
}

func encodeContent(text string) (string, error) {
	bs, err := json.Marshal(textContent{Text: text})
	if err != nil {
		logger.Errorf("encodeContent marshal err=%v", err)
		return "", err
	}
	return string(bs), nil
}

func decodeText(content string) string {
	var c textContent
	if err := json.Unmarshal([]byte(content), &c); err != nil {
		logger.Errorf("decodeText unmarshal err=%v", err)
		return ""
	}
	return c.Text
}

func seqForViewer(m *model.Message, viewerID int64) int64 {
	if viewerID == m.UserLow {
		return m.SeqLow
	}
	if viewerID == m.UserHigh {
		return m.SeqHigh
	}
	return 0
}

func clientMsgIDForUser(m *model.Message) string {
	if m.Origin != apicommon.Origin_ORIGIN_USER {
		return ""
	}
	return m.IdempotencyKey
}

func toAPIMsg(m *model.Message, viewerID int64) *api.Msg {
	return &api.Msg{
		MsgId:       m.ID,
		ConvId:      m.ConvID,
		Seq:         seqForViewer(m, viewerID),
		SenderId:    m.SenderID,
		Origin:      m.Origin,
		ClientMsgId: clientMsgIDForUser(m),
		ContentType: m.ContentType,
		Content:     &api.MsgContent{Text: decodeText(m.Content)},
		Ct:          m.Ct,
		Ut:          m.Ut,
	}
}

func toAPIConv(
	member *model.Member,
	peerNickname, peerAvatar string,
	lastOrigin apicommon.Origin,
	lastContentType apicommon.ContentType,
	lastCt, lastUt int64,
) *api.Conversation {
	conv := &api.Conversation{
		ConvId: member.ConvID,
		Peer: &api.Peer{
			UserId:   member.PeerUserID,
			Nickname: peerNickname,
			Avatar:   peerAvatar,
		},
		LastSeq:   member.LastSeq,
		LastMsgId: member.LastMsgID,
		SortUt:    member.SortUt,
	}
	if member.LastMsgID > 0 {
		conv.LastMsg = &api.MsgPreview{
			MsgId:       member.LastMsgID,
			SenderId:    member.LastSenderID,
			Origin:      lastOrigin,
			ContentType: lastContentType,
			Preview:     member.LastPreview,
			Ct:          lastCt,
			Ut:          lastUt,
		}
	}
	return conv
}
