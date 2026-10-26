package biz_test

import (
	"path/filepath"
	"testing"
	"time"

	apicommon "mall/im/api/common"
	api "mall/im/api/im"
	"mall/im/common/database"
	"mall/im/pkg/biz"
	"mall/im/pkg/bizerror"
	imdao "mall/im/pkg/dao"
	"mall/kit/codeerror"
	"mall/kit/logger"
	userdao "mall/user/dao"
	usermodel "mall/user/model"
)

func setup(t *testing.T) (userA, userB, userC int64) {
	t.Helper()
	logger.InitLogger("")
	dir := t.TempDir()
	database.InitDb(filepath.Join(dir, "test.db"))
	t.Cleanup(database.Close)
	userdao.SetDB(database.GetDb())
	if err := userdao.AutoMigrate(); err != nil {
		t.Fatalf("user migrate: %v", err)
	}
	if err := imdao.AutoMigrate(); err != nil {
		t.Fatalf("im migrate: %v", err)
	}
	biz.SetTextMaxRunes(2000)

	now := time.Now().Unix()
	for _, nick := range []string{"A", "B", "C"} {
		u := &usermodel.User{Nickname: nick, Ct: now, Ut: now}
		if err := userdao.NewUserDao(nil).Create(u); err != nil {
			t.Fatalf("create user %s: %v", nick, err)
		}
		switch nick {
		case "A":
			userA = u.ID
		case "B":
			userB = u.ID
		case "C":
			userC = u.ID
		}
	}
	return userA, userB, userC
}

func TestFieldDocExample(t *testing.T) {
	a, b, _ := setup(t)

	reply, pushes, err := biz.Send(a, b, "client01", apicommon.ContentType_CONTENT_TYPE_TEXT, "你好", "")
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	if reply.Msg.GetSeq() != 1 || reply.Conv.GetLastSeq() != 1 {
		t.Fatalf("sender seq want 1 got msg=%d conv=%d", reply.Msg.GetSeq(), reply.Conv.GetLastSeq())
	}
	if len(pushes) != 2 {
		t.Fatalf("pushes want 2 got %d", len(pushes))
	}
	msgID1 := reply.Msg.GetMsgId()

	internal, pushes, err := biz.InternalSend(&api.InternalSendRequest{
		SenderId:      a,
		PeerUserId:    b,
		Visibility:    apicommon.Visibility_VISIBILITY_PEER_ONLY,
		SenderPreview: false,
		PeerPreview:   true,
		RequestId:     "sys-0001",
		ContentType:   apicommon.ContentType_CONTENT_TYPE_TEXT,
		Text:          "只给B",
	})
	if err != nil {
		t.Fatalf("internal send: %v", err)
	}
	if internal.GetSenderSeq() != 0 || internal.GetPeerSeq() != 2 {
		t.Fatalf("internal seqs sender=%d peer=%d", internal.GetSenderSeq(), internal.GetPeerSeq())
	}
	if len(pushes) != 1 || pushes[0].UserID != b {
		t.Fatalf("peer-only push %#v", pushes)
	}

	histA, err := biz.History(a, b, 0, 50)
	if err != nil {
		t.Fatalf("history A: %v", err)
	}
	if len(histA.List) != 1 || histA.List[0].GetMsgId() != msgID1 {
		t.Fatalf("A history %#v", histA.List)
	}
	histB, err := biz.History(b, a, 0, 50)
	if err != nil {
		t.Fatalf("history B: %v", err)
	}
	if len(histB.List) != 2 {
		t.Fatalf("B history len=%d", len(histB.List))
	}

	upd, pushes, err := biz.InternalUpdate(&api.InternalUpdateRequest{
		MsgId:         msgID1,
		ContentType:   apicommon.ContentType_CONTENT_TYPE_TEXT,
		Text:          "你好改了",
		SenderPreview: true,
		PeerPreview:   true,
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if upd.GetSenderSeq() != 2 {
		t.Fatalf("A update seq want 2 got %d", upd.GetSenderSeq())
	}
	if upd.GetPeerSeq() != 3 {
		t.Fatalf("B update seq want 3 got %d", upd.GetPeerSeq())
	}
	if len(pushes) != 2 {
		t.Fatalf("update pushes %d", len(pushes))
	}

	histA2, err := biz.History(a, b, 1, 50)
	if err != nil {
		t.Fatalf("history A after update: %v", err)
	}
	if len(histA2.List) != 1 || histA2.List[0].GetMsgId() != msgID1 || histA2.List[0].GetSeq() != 2 {
		t.Fatalf("A delta after update %#v", histA2.List)
	}
	if histA2.List[0].GetContent().GetText() != "你好改了" {
		t.Fatalf("text=%s", histA2.List[0].GetContent().GetText())
	}

	listB, err := biz.ListConv(b, 0, 0)
	if err != nil {
		t.Fatalf("list B: %v", err)
	}
	if len(listB.List) != 1 || listB.List[0].GetLastSeq() != 3 {
		t.Fatalf("list B %#v", listB.List)
	}
	if listB.List[0].GetLastMsgId() == msgID1 {
		t.Fatalf("B last msg should still be peer-only message, got %d", listB.List[0].GetLastMsgId())
	}
}

func TestIdempotency(t *testing.T) {
	a, b, _ := setup(t)
	r1, _, err := biz.Send(a, b, "client02", apicommon.ContentType_CONTENT_TYPE_TEXT, "same", "")
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	r2, pushes, err := biz.Send(a, b, "client02", apicommon.ContentType_CONTENT_TYPE_TEXT, "same", "")
	if err != nil {
		t.Fatalf("retry: %v", err)
	}
	if r1.Msg.GetMsgId() != r2.Msg.GetMsgId() {
		t.Fatalf("idempotent msg id mismatch")
	}
	if len(pushes) != 0 {
		t.Fatalf("idempotent retry should not push, got %d", len(pushes))
	}
	_, _, err = biz.Send(a, b, "client02", apicommon.ContentType_CONTENT_TYPE_TEXT, "different", "")
	ce, ok := codeerror.As(err)
	if !ok || !ce.Is(bizerror.ErrorIdempotency) {
		t.Fatalf("want idempotency error, got %v", err)
	}
}

func TestListCursorAcrossConvs(t *testing.T) {
	a, b, c := setup(t)
	if _, _, err := biz.Send(a, b, "ab-msg-01", apicommon.ContentType_CONTENT_TYPE_TEXT, "toB", ""); err != nil {
		t.Fatalf("send B: %v", err)
	}
	if _, _, err := biz.Send(a, c, "ac-msg-01", apicommon.ContentType_CONTENT_TYPE_TEXT, "toC", ""); err != nil {
		t.Fatalf("send C: %v", err)
	}
	// A 的序号跨会话递增：与 B 是 1，与 C 是 2。
	list, err := biz.ListConv(a, 1, 0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list.List) != 1 || list.List[0].GetPeer().GetUserId() != c {
		t.Fatalf("after_seq=1 should return C conv, got %#v", list.List)
	}
	list0, err := biz.ListConv(a, 0, 0)
	if err != nil {
		t.Fatalf("list0: %v", err)
	}
	if len(list0.List) != 2 {
		t.Fatalf("want 2 convs got %d", len(list0.List))
	}
	if list0.List[0].GetLastSeq() > list0.List[1].GetLastSeq() {
		t.Fatalf("list should be ascending by last_seq")
	}
}

func TestHistoryEmptyWithoutMember(t *testing.T) {
	a, b, _ := setup(t)
	h, err := biz.History(a, b, 0, 50)
	if err != nil {
		t.Fatalf("history: %v", err)
	}
	if len(h.List) != 0 || h.GetLastSeq() != 0 || h.GetHasMore() {
		t.Fatalf("want empty history %#v", h)
	}
}

func TestInternalSenderOnlyNoPreview(t *testing.T) {
	a, b, _ := setup(t)
	if _, _, err := biz.Send(a, b, "seed-msg1", apicommon.ContentType_CONTENT_TYPE_TEXT, "seed", ""); err != nil {
		t.Fatalf("seed: %v", err)
	}
	listBefore, err := biz.ListConv(a, 0, 0)
	if err != nil {
		t.Fatalf("list before: %v", err)
	}
	oldLast := listBefore.List[0].GetLastMsgId()

	_, _, err = biz.InternalSend(&api.InternalSendRequest{
		SenderId:      a,
		PeerUserId:    b,
		Visibility:    apicommon.Visibility_VISIBILITY_SENDER_ONLY,
		SenderPreview: false,
		PeerPreview:   false,
		RequestId:     "sys-only-a",
		ContentType:   apicommon.ContentType_CONTENT_TYPE_TEXT,
		Text:          "hidden preview",
	})
	if err != nil {
		t.Fatalf("internal: %v", err)
	}
	listAfter, err := biz.ListConv(a, 0, 0)
	if err != nil {
		t.Fatalf("list after: %v", err)
	}
	if listAfter.List[0].GetLastMsgId() != oldLast {
		t.Fatalf("preview should stay, before=%d after=%d", oldLast, listAfter.List[0].GetLastMsgId())
	}
	if listAfter.List[0].GetLastSeq() <= listBefore.List[0].GetLastSeq() {
		t.Fatalf("last_seq should advance")
	}
	hist, err := biz.History(a, b, listBefore.List[0].GetLastSeq(), 50)
	if err != nil {
		t.Fatalf("history: %v", err)
	}
	if len(hist.List) != 1 || hist.List[0].GetContent().GetText() != "hidden preview" {
		t.Fatalf("history %#v", hist.List)
	}
}
