package biz

import (
	"sort"
	"time"

	apicommon "mall/im/api/common"
	api "mall/im/api/im"
	"mall/im/common/database"
	"mall/im/pkg/bizerror"
	imdao "mall/im/pkg/dao"
	"mall/im/pkg/model"
	"mall/file/mediastore"
	"mall/kit/codeerror"
	"mall/kit/logger"
	userdao "mall/user/dao"

	"gorm.io/gorm"
)

var textMaxRunes = 2000

// SetTextMaxRunes 由进程启动时写入配置。
func SetTextMaxRunes(n int) {
	if n > 0 {
		textMaxRunes = n
	}
}

// PushItem 提交成功后推给某个用户的一条下行消息。
type PushItem struct {
	UserID int64
	Msg    *api.Msg
}

type writeResult struct {
	Msg    *model.Message
	Pushes []PushItem
}

func requireUser(userID int64) error {
	_, err := userdao.NewUserDao(nil).GetByID(userID)
	if err != nil {
		logger.Errorf("biz requireUser id=%d err=%v", userID, err)
		return err
	}
	return nil
}

func nicknameOf(userID int64) string {
	n, _ := profileOf(userID)
	return n
}

func profileOf(userID int64) (nickname, avatar string) {
	u, err := userdao.NewUserDao(nil).GetByID(userID)
	if err != nil {
		logger.Errorf("biz profileOf id=%d err=%v", userID, err)
		return "", ""
	}
	return u.Nickname, avatarURL(u.Avatar)
}

func avatarURL(uri string) string {
	if uri == "" {
		return ""
	}
	return mediastore.Get().PathURL(uri)
}

// Send 用户发送。双方可见，两侧预览都更新。
func Send(
	senderID, peerUserID int64,
	clientMsgID string,
	contentType apicommon.ContentType,
	text, imageURI string,
) (*api.SendReply, []PushItem, error) {
	if err := validatePeer(senderID, peerUserID); err != nil {
		return nil, nil, err
	}
	if err := validateIdempotencyKey(clientMsgID); err != nil {
		return nil, nil, err
	}
	if err := validateContentType(contentType); err != nil {
		return nil, nil, err
	}
	content, fingerprint, preview, err := prepareBody(contentType, text, imageURI)
	if err != nil {
		return nil, nil, err
	}
	if err := requireUser(peerUserID); err != nil {
		return nil, nil, err
	}

	var out writeResult
	err = database.GetDb().Transaction(func(tx *gorm.DB) error {
		res, err := createMessage(tx, createParams{
			SenderID:            senderID,
			PeerUserID:          peerUserID,
			Origin:              apicommon.Origin_ORIGIN_USER,
			Visibility:          apicommon.Visibility_VISIBILITY_BOTH,
			SenderPreview:       true,
			PeerPreview:         true,
			IdempotencyKey:      clientMsgID,
			ContentType:         contentType,
			Content:             content,
			CompareFingerprint:  fingerprint,
			CompareContentTy:    contentType,
			Preview:             preview,
		})
		if err != nil {
			return err
		}
		out = res
		return nil
	})
	if err != nil {
		logger.Errorf("Send tx sender=%d peer=%d err=%v", senderID, peerUserID, err)
		return nil, nil, err
	}

	member, err := imdao.NewMemberDao(nil).Get(senderID, peerUserID)
	if err != nil {
		return nil, nil, err
	}
	if member == nil {
		logger.Errorf("Send missing member after write sender=%d peer=%d", senderID, peerUserID)
		return nil, nil, codeerror.ErrorSystem
	}
	lastOrigin, lastCT, lastCt, lastUt := lastMsgMeta(out.Msg, member)
	peerNick, peerAvatar := profileOf(peerUserID)
	reply := &api.SendReply{
		Msg:  toAPIMsg(out.Msg, senderID),
		Conv: toAPIConv(member, peerNick, peerAvatar, lastOrigin, lastCT, lastCt, lastUt),
	}
	return reply, out.Pushes, nil
}

// InternalSend 系统代发。
func InternalSend(req *api.InternalSendRequest) (*api.InternalSendReply, []PushItem, error) {
	senderID := req.GetSenderId()
	peerID := req.GetPeerUserId()
	if err := validatePeer(senderID, peerID); err != nil {
		return nil, nil, err
	}
	if err := validateIdempotencyKey(req.GetRequestId()); err != nil {
		return nil, nil, err
	}
	if err := validateContentType(req.GetContentType()); err != nil {
		return nil, nil, err
	}
	if err := validateVisibility(req.GetVisibility()); err != nil {
		return nil, nil, err
	}
	vis := req.GetVisibility()
	if req.GetSenderPreview() && !visibleFor(senderID, senderID, peerID, vis) {
		logger.Errorf("InternalSend sender_preview on invisible side sender=%d", senderID)
		return nil, nil, bizerror.ErrorBadMessage
	}
	if req.GetPeerPreview() && !visibleFor(peerID, senderID, peerID, vis) {
		logger.Errorf("InternalSend peer_preview on invisible side peer=%d", peerID)
		return nil, nil, bizerror.ErrorBadMessage
	}
	content, fingerprint, preview, err := prepareBody(req.GetContentType(), req.GetText(), req.GetImageUri())
	if err != nil {
		return nil, nil, err
	}
	if err := requireUser(senderID); err != nil {
		return nil, nil, err
	}
	if err := requireUser(peerID); err != nil {
		return nil, nil, err
	}

	var out writeResult
	err = database.GetDb().Transaction(func(tx *gorm.DB) error {
		res, err := createMessage(tx, createParams{
			SenderID:           senderID,
			PeerUserID:         peerID,
			Origin:             apicommon.Origin_ORIGIN_SYSTEM,
			Visibility:         vis,
			SenderPreview:      req.GetSenderPreview(),
			PeerPreview:        req.GetPeerPreview(),
			IdempotencyKey:     req.GetRequestId(),
			ContentType:        req.GetContentType(),
			Content:            content,
			CompareFingerprint: fingerprint,
			CompareContentTy:   req.GetContentType(),
			Preview:            preview,
		})
		if err != nil {
			return err
		}
		out = res
		return nil
	})
	if err != nil {
		logger.Errorf("InternalSend tx sender=%d peer=%d err=%v", senderID, peerID, err)
		return nil, nil, err
	}

	return &api.InternalSendReply{
		MsgId:     out.Msg.ID,
		ConvId:    out.Msg.ConvID,
		SenderSeq: seqForViewer(out.Msg, senderID),
		PeerSeq:   seqForViewer(out.Msg, peerID),
	}, out.Pushes, nil
}

// InternalUpdate 更新已有消息正文。
func InternalUpdate(req *api.InternalUpdateRequest) (*api.InternalSendReply, []PushItem, error) {
	if req.GetMsgId() <= 0 {
		logger.Errorf("InternalUpdate msg_id invalid")
		return nil, nil, bizerror.ErrorBadMessage
	}
	if err := validateContentType(req.GetContentType()); err != nil {
		return nil, nil, err
	}
	content, _, preview, err := prepareBody(req.GetContentType(), req.GetText(), req.GetImageUri())
	if err != nil {
		return nil, nil, err
	}

	var out writeResult
	var senderID, peerID int64
	err = database.GetDb().Transaction(func(tx *gorm.DB) error {
		msgDao := imdao.NewMessageDao(tx)
		msg, err := msgDao.GetByIDForUpdate(req.GetMsgId())
		if err != nil {
			return err
		}
		if msg == nil {
			logger.Errorf("InternalUpdate message not found id=%d", req.GetMsgId())
			return bizerror.ErrorMsgNotFound
		}
		senderID = msg.SenderID
		if msg.UserLow == senderID {
			peerID = msg.UserHigh
		} else {
			peerID = msg.UserLow
		}
		vis := msg.Visibility
		if req.GetSenderPreview() && !visibleFor(senderID, senderID, peerID, vis) {
			logger.Errorf("InternalUpdate sender_preview on invisible side msg=%d", msg.ID)
			return bizerror.ErrorBadMessage
		}
		if req.GetPeerPreview() && !visibleFor(peerID, senderID, peerID, vis) {
			logger.Errorf("InternalUpdate peer_preview on invisible side msg=%d", msg.ID)
			return bizerror.ErrorBadMessage
		}

		res, err := updateMessage(tx, msg, updateParams{
			Content:       content,
			ContentType:   req.GetContentType(),
			SenderPreview: req.GetSenderPreview(),
			PeerPreview:   req.GetPeerPreview(),
			SenderID:      senderID,
			PeerID:        peerID,
			Preview:       preview,
		})
		if err != nil {
			return err
		}
		out = res
		return nil
	})
	if err != nil {
		logger.Errorf("InternalUpdate tx msgID=%d err=%v", req.GetMsgId(), err)
		return nil, nil, err
	}

	return &api.InternalSendReply{
		MsgId:     out.Msg.ID,
		ConvId:    out.Msg.ConvID,
		SenderSeq: seqForViewer(out.Msg, senderID),
		PeerSeq:   seqForViewer(out.Msg, peerID),
	}, out.Pushes, nil
}

// ListOnlineUsers 组装在线用户资料。entries 由 Hub 提供。
func ListOnlineUsers(selfID int64, entries []OnlineEntry) (*api.ListOnlineReply, error) {
	out := make([]*api.OnlineUser, 0, len(entries))
	for _, e := range entries {
		if e.UserID == selfID || e.UserID <= 0 {
			continue
		}
		nick, avatar := profileOf(e.UserID)
		out = append(out, &api.OnlineUser{
			UserId:    e.UserID,
			Nickname:  nick,
			ConnCount: int32(e.ConnCount),
			Avatar:    avatar,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Nickname == out[j].Nickname {
			return out[i].UserId < out[j].UserId
		}
		return out[i].Nickname < out[j].Nickname
	})
	return &api.ListOnlineReply{List: out}, nil
}

// OnlineEntry 与 hub 解耦的在线条目。
type OnlineEntry struct {
	UserID    int64
	ConnCount int
}

// ListConv 按 after_seq 增量返回会话。
func ListConv(userID, afterSeq int64, limit int32) (*api.ListConvReply, error) {
	if afterSeq < 0 {
		logger.Errorf("ListConv after_seq invalid user=%d after=%d", userID, afterSeq)
		return nil, codeerror.ErrorParam
	}
	fetch := 0
	if limit > 0 {
		fetch = int(limit) + 1
	}
	list, err := imdao.NewMemberDao(nil).ListAfterSeq(userID, afterSeq, fetch)
	if err != nil {
		logger.Errorf("ListConv dao user=%d err=%v", userID, err)
		return nil, err
	}
	hasMore := false
	if limit > 0 && len(list) > int(limit) {
		hasMore = true
		list = list[:limit]
	}
	out := make([]*api.Conversation, 0, len(list))
	msgDao := imdao.NewMessageDao(nil)
	for i := range list {
		m := &list[i]
		var lastOrigin apicommon.Origin
		var lastCT apicommon.ContentType
		var lastCt, lastUt int64
		if m.LastMsgID > 0 {
			msg, err := msgDao.GetByID(m.LastMsgID)
			if err != nil {
				return nil, err
			}
			if msg != nil {
				lastOrigin = msg.Origin
				lastCT = msg.ContentType
				lastCt = msg.Ct
				lastUt = msg.Ut
			}
		}
		peerNick, peerAvatar := profileOf(m.PeerUserID)
		out = append(out, toAPIConv(m, peerNick, peerAvatar, lastOrigin, lastCT, lastCt, lastUt))
	}
	return &api.ListConvReply{List: out, HasMore: hasMore}, nil
}

// History 按 after_seq 增量返回一场会话的历史。
func History(userID, peerUserID, afterSeq int64, limit int32) (*api.HistoryReply, error) {
	if err := validatePeer(userID, peerUserID); err != nil {
		return nil, err
	}
	if afterSeq < 0 {
		logger.Errorf("History after_seq invalid user=%d after=%d", userID, afterSeq)
		return nil, codeerror.ErrorParam
	}
	if err := requireUser(peerUserID); err != nil {
		return nil, err
	}
	pageSize := int(limit)
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 100 {
		pageSize = 100
	}

	member, err := imdao.NewMemberDao(nil).Get(userID, peerUserID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return &api.HistoryReply{List: []*api.Msg{}, LastSeq: 0, HasMore: false}, nil
	}

	low, high := PairIDs(userID, peerUserID)
	msgs, err := imdao.NewMessageDao(nil).ListAfterSeq(low, high, userID, afterSeq, pageSize+1)
	if err != nil {
		logger.Errorf("History dao user=%d peer=%d err=%v", userID, peerUserID, err)
		return nil, err
	}
	hasMore := false
	if len(msgs) > pageSize {
		hasMore = true
		msgs = msgs[:pageSize]
	}
	out := make([]*api.Msg, 0, len(msgs))
	for i := range msgs {
		out = append(out, toAPIMsg(&msgs[i], userID))
	}
	return &api.HistoryReply{List: out, LastSeq: member.LastSeq, HasMore: hasMore}, nil
}

type createParams struct {
	SenderID           int64
	PeerUserID         int64
	Origin             apicommon.Origin
	Visibility         apicommon.Visibility
	SenderPreview      bool
	PeerPreview        bool
	IdempotencyKey     string
	ContentType        apicommon.ContentType
	Content            string
	CompareFingerprint string
	CompareContentTy   apicommon.ContentType
	Preview            string
}

type updateParams struct {
	Content       string
	ContentType   apicommon.ContentType
	SenderPreview bool
	PeerPreview   bool
	SenderID      int64
	PeerID        int64
	Preview       string
}

// prepareBody 按类型校验并编码正文，返回库内 JSON、幂等指纹、列表预览文案。
func prepareBody(ct apicommon.ContentType, text, imageURI string) (content, fingerprint, preview string, err error) {
	switch ct {
	case apicommon.ContentType_CONTENT_TYPE_TEXT:
		if err = validateText(text, textMaxRunes); err != nil {
			return "", "", "", err
		}
		content, err = encodeTextContent(text)
		if err != nil {
			return "", "", "", codeerror.OrSystem(err, "prepareBody encodeText")
		}
		return content, text, PreviewText(text, 50), nil
	case apicommon.ContentType_CONTENT_TYPE_IMAGE:
		uri, nerr := normalizeImageURI(imageURI)
		if nerr != nil {
			return "", "", "", nerr
		}
		content, err = encodeImageContent(uri)
		if err != nil {
			return "", "", "", codeerror.OrSystem(err, "prepareBody encodeImage")
		}
		return content, uri, imagePreview, nil
	default:
		logger.Errorf("prepareBody unsupported type=%d", ct)
		return "", "", "", bizerror.ErrorBadMessage
	}
}

func createMessage(tx *gorm.DB, p createParams) (writeResult, error) {
	msgDao := imdao.NewMessageDao(tx)
	existing, err := msgDao.GetByIdempotency(p.Origin, p.SenderID, p.IdempotencyKey)
	if err != nil {
		return writeResult{}, err
	}
	if existing != nil {
		if existing.ContentType != p.CompareContentTy || contentFingerprint(existing.ContentType, existing.Content) != p.CompareFingerprint {
			logger.Errorf("createMessage idempotency conflict sender=%d key=%s", p.SenderID, p.IdempotencyKey)
			return writeResult{}, bizerror.ErrorIdempotency
		}
		return writeResult{Msg: existing, Pushes: nil}, nil
	}

	low, high := PairIDs(p.SenderID, p.PeerUserID)
	convID := ConvID(p.SenderID, p.PeerUserID)
	now := time.Now().Unix()

	visibleIDs := visibleUserIDs(p.SenderID, p.PeerUserID, p.Visibility)
	sort.Slice(visibleIDs, func(i, j int) bool { return visibleIDs[i] < visibleIDs[j] })

	seqDao := imdao.NewUserSeqDao(tx)
	seqs := make(map[int64]int64, len(visibleIDs))
	for _, uid := range visibleIDs {
		seq, err := seqDao.Next(uid)
		if err != nil {
			return writeResult{}, err
		}
		seqs[uid] = seq
	}

	memberDao := imdao.NewMemberDao(tx)
	members := make(map[int64]*model.Member, len(visibleIDs))
	for _, uid := range visibleIDs {
		peer := p.PeerUserID
		if uid == p.PeerUserID {
			peer = p.SenderID
		}
		m, err := memberDao.Ensure(uid, peer, low, high, convID)
		if err != nil {
			return writeResult{}, err
		}
		members[uid] = m
	}

	msg := &model.Message{
		UserLow:        low,
		UserHigh:       high,
		ConvID:         convID,
		SenderID:       p.SenderID,
		Origin:         p.Origin,
		Visibility:     p.Visibility,
		SeqLow:         seqs[low],
		SeqHigh:        seqs[high],
		IdempotencyKey: p.IdempotencyKey,
		ContentType:    p.ContentType,
		Content:        p.Content,
		Ct:             now,
		Ut:             now,
	}
	if err := msgDao.Create(msg); err != nil {
		// 并发幂等：唯一键冲突后读回。
		again, getErr := msgDao.GetByIdempotency(p.Origin, p.SenderID, p.IdempotencyKey)
		if getErr != nil {
			return writeResult{}, getErr
		}
		if again != nil {
			if again.ContentType != p.CompareContentTy || contentFingerprint(again.ContentType, again.Content) != p.CompareFingerprint {
				return writeResult{}, bizerror.ErrorIdempotency
			}
			return writeResult{Msg: again}, nil
		}
		return writeResult{}, err
	}

	preview := p.Preview
	for _, uid := range visibleIDs {
		m := members[uid]
		m.LastSeq = seqs[uid]
		wantPreview := p.PeerPreview
		if uid == p.SenderID {
			wantPreview = p.SenderPreview
		}
		if wantPreview {
			updatePreview := m.LastMsgID == 0
			if !updatePreview {
				old, err := msgDao.GetByID(m.LastMsgID)
				if err != nil {
					return writeResult{}, err
				}
				updatePreview = old == nil || isNewerMsg(msg.Ct, msg.ID, old.Ct, old.ID)
			}
			if updatePreview {
				m.LastMsgID = msg.ID
				m.LastSenderID = msg.SenderID
				m.LastPreview = preview
				m.SortUt = msg.Ut
			}
		}
		if err := memberDao.Save(m); err != nil {
			return writeResult{}, err
		}
	}

	pushes := make([]PushItem, 0, len(visibleIDs))
	for _, uid := range visibleIDs {
		pushes = append(pushes, PushItem{UserID: uid, Msg: toAPIMsg(msg, uid)})
	}
	return writeResult{Msg: msg, Pushes: pushes}, nil
}

func updateMessage(tx *gorm.DB, msg *model.Message, p updateParams) (writeResult, error) {
	msgDao := imdao.NewMessageDao(tx)
	memberDao := imdao.NewMemberDao(tx)
	seqDao := imdao.NewUserSeqDao(tx)

	visibleIDs := visibleUserIDs(p.SenderID, p.PeerID, msg.Visibility)
	sort.Slice(visibleIDs, func(i, j int) bool { return visibleIDs[i] < visibleIDs[j] })

	now := time.Now().Unix()
	seqs := make(map[int64]int64, len(visibleIDs))
	for _, uid := range visibleIDs {
		seq, err := seqDao.Next(uid)
		if err != nil {
			return writeResult{}, err
		}
		seqs[uid] = seq
	}

	msg.Content = p.Content
	msg.ContentType = p.ContentType
	msg.Ut = now
	if seq, ok := seqs[msg.UserLow]; ok {
		msg.SeqLow = seq
	}
	if seq, ok := seqs[msg.UserHigh]; ok {
		msg.SeqHigh = seq
	}
	if err := msgDao.Save(msg); err != nil {
		return writeResult{}, err
	}

	preview := p.Preview
	for _, uid := range visibleIDs {
		peer := p.PeerID
		if uid == p.PeerID {
			peer = p.SenderID
		}
		m, err := memberDao.GetForUpdate(uid, peer)
		if err != nil {
			return writeResult{}, err
		}
		if m == nil {
			// 可见侧理论上已有行；若缺失则补建。
			m, err = memberDao.Ensure(uid, peer, msg.UserLow, msg.UserHigh, msg.ConvID)
			if err != nil {
				return writeResult{}, err
			}
		}
		m.LastSeq = seqs[uid]
		wantPreview := false
		if uid == p.SenderID {
			wantPreview = p.SenderPreview
		} else {
			wantPreview = p.PeerPreview
		}
		if wantPreview && m.LastMsgID == msg.ID {
			m.LastSenderID = msg.SenderID
			m.LastPreview = preview
			m.SortUt = msg.Ut
		}
		if err := memberDao.Save(m); err != nil {
			return writeResult{}, err
		}
	}

	pushes := make([]PushItem, 0, len(visibleIDs))
	for _, uid := range visibleIDs {
		pushes = append(pushes, PushItem{UserID: uid, Msg: toAPIMsg(msg, uid)})
	}
	return writeResult{Msg: msg, Pushes: pushes}, nil
}

func visibleUserIDs(senderID, peerID int64, v apicommon.Visibility) []int64 {
	switch v {
	case apicommon.Visibility_VISIBILITY_BOTH:
		return []int64{senderID, peerID}
	case apicommon.Visibility_VISIBILITY_PEER_ONLY:
		return []int64{peerID}
	case apicommon.Visibility_VISIBILITY_SENDER_ONLY:
		return []int64{senderID}
	default:
		return nil
	}
}

func lastMsgMeta(msg *model.Message, member *model.Member) (apicommon.Origin, apicommon.ContentType, int64, int64) {
	if member.LastMsgID == 0 {
		return 0, 0, 0, 0
	}
	if msg != nil && msg.ID == member.LastMsgID {
		return msg.Origin, msg.ContentType, msg.Ct, msg.Ut
	}
	old, err := imdao.NewMessageDao(nil).GetByID(member.LastMsgID)
	if err != nil || old == nil {
		return 0, 0, 0, 0
	}
	return old.Origin, old.ContentType, old.Ct, old.Ut
}
