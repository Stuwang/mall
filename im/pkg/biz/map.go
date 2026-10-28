package biz

import (
	"encoding/json"
	"strconv"

	apicommon "mall/im/api/common"
	api "mall/im/api/im"
	"mall/file/mediastore"
	"mall/im/pkg/model"
	"mall/kit/logger"
)

type msgPayload struct {
	Text       string `json:"text,omitempty"`
	ImageURI   string `json:"image_uri,omitempty"`
	GiftID     int64  `json:"gift_id,omitempty"`
	GiftName   string `json:"gift_name,omitempty"`
	GiftIcon   string `json:"gift_icon,omitempty"`
	GiftEffect string `json:"gift_effect,omitempty"`
	GiftPrice  int64  `json:"gift_price,omitempty"`
}

const imagePreview = "[图片]"

func encodeTextContent(text string) (string, error) {
	bs, err := json.Marshal(msgPayload{Text: text})
	if err != nil {
		logger.Errorf("encodeTextContent marshal err=%v", err)
		return "", err
	}
	return string(bs), nil
}

func encodeImageContent(imageURI string) (string, error) {
	bs, err := json.Marshal(msgPayload{ImageURI: imageURI})
	if err != nil {
		logger.Errorf("encodeImageContent marshal err=%v", err)
		return "", err
	}
	return string(bs), nil
}

func encodeGiftContent(id int64, name, icon, effect string, price int64) (string, error) {
	bs, err := json.Marshal(msgPayload{
		GiftID:     id,
		GiftName:   name,
		GiftIcon:   icon,
		GiftEffect: effect,
		GiftPrice:  price,
	})
	if err != nil {
		logger.Errorf("encodeGiftContent marshal err=%v", err)
		return "", err
	}
	return string(bs), nil
}

func decodePayload(content string) msgPayload {
	var c msgPayload
	if err := json.Unmarshal([]byte(content), &c); err != nil {
		logger.Errorf("decodePayload unmarshal err=%v", err)
		return msgPayload{}
	}
	return c
}

func decodeText(content string) string {
	return decodePayload(content).Text
}

func decodeImageURI(content string) string {
	return decodePayload(content).ImageURI
}

func contentFingerprint(ct apicommon.ContentType, content string) string {
	p := decodePayload(content)
	switch ct {
	case apicommon.ContentType_CONTENT_TYPE_IMAGE:
		return p.ImageURI
	case apicommon.ContentType_CONTENT_TYPE_GIFT:
		return strconv.FormatInt(p.GiftID, 10)
	default:
		return p.Text
	}
}

func toAPIContent(ct apicommon.ContentType, raw string) *api.MsgContent {
	p := decodePayload(raw)
	out := &api.MsgContent{Text: p.Text}
	if ct == apicommon.ContentType_CONTENT_TYPE_IMAGE && p.ImageURI != "" {
		out.ImageUrl = mediastore.Get().PathURL(p.ImageURI)
	}
	if ct == apicommon.ContentType_CONTENT_TYPE_GIFT {
		out.Gift = &api.GiftBody{
			GiftId: p.GiftID,
			Name:   p.GiftName,
			Icon:   p.GiftIcon,
			Effect: p.GiftEffect,
			Price:  p.GiftPrice,
		}
	}
	return out
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
		Content:     toAPIContent(m.ContentType, m.Content),
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
