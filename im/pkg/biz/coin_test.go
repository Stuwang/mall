package biz_test

import (
	"testing"

	apicommon "mall/im/api/common"
	"mall/im/pkg/biz"
	"mall/im/pkg/bizerror"
	imdao "mall/im/pkg/dao"
	"mall/kit/codeerror"
)

func setTestCoin(t *testing.T) {
	t.Helper()
	biz.SetCoinConfig(biz.CoinConfig{
		CoinsPerYuan:  10,
		PayTimeoutSec: 900,
		Gifts: []biz.GiftItem{
			{ID: 1, Name: "玫瑰", Icon: "🌹", Effect: "", Price: 1},
			{ID: 2, Name: "啤酒", Icon: "🍺", Effect: "cheers", Price: 10},
			{ID: 3, Name: "跑车", Icon: "🏎️", Effect: "car", Price: 99},
			{ID: 4, Name: "城堡", Icon: "🏰", Effect: "castle", Price: 520},
		},
	})
}

func TestRechargeMockPayOnce(t *testing.T) {
	a, _, _ := setup(t)
	setTestCoin(t)

	order, err := biz.CreateRecharge(a, 10, "")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if order.GetCoins() != 10 || order.GetCoinsPerYuan() != 10 || order.GetCashAmount() != 100 {
		t.Fatalf("snapshot coins=%d rate=%d cash=%d", order.GetCoins(), order.GetCoinsPerYuan(), order.GetCashAmount())
	}
	if order.GetStatus() != apicommon.CoinOrderStatus_COIN_ORDER_STATUS_PENDING_PAY {
		t.Fatalf("status=%d", order.GetStatus())
	}

	paying, err := biz.PayRecharge(a, order.GetId(), apicommon.PayType_PAY_TYPE_MOCK)
	if err != nil {
		t.Fatalf("pay: %v", err)
	}
	if paying.GetPayType() != apicommon.PayType_PAY_TYPE_MOCK ||
		paying.GetStatus() != apicommon.CoinOrderStatus_COIN_ORDER_STATUS_PENDING_PAY {
		t.Fatalf("after pay type=%d status=%d", paying.GetPayType(), paying.GetStatus())
	}
	bal, err := biz.Balance(a)
	if err != nil {
		t.Fatalf("balance after pay: %v", err)
	}
	if bal.GetBalance() != 0 {
		t.Fatalf("pay should not credit, balance=%d", bal.GetBalance())
	}

	paid, err := biz.MockPaySuccess(a, order.GetId())
	if err != nil {
		t.Fatalf("mock: %v", err)
	}
	if paid.GetStatus() != apicommon.CoinOrderStatus_COIN_ORDER_STATUS_PAID {
		t.Fatalf("paid status=%d", paid.GetStatus())
	}
	bal, err = biz.Balance(a)
	if err != nil {
		t.Fatalf("balance: %v", err)
	}
	if bal.GetBalance() != 10 {
		t.Fatalf("balance=%d", bal.GetBalance())
	}

	if _, err := biz.MockPaySuccess(a, order.GetId()); err != nil {
		t.Fatalf("mock again: %v", err)
	}
	bal, err = biz.Balance(a)
	if err != nil {
		t.Fatalf("balance again: %v", err)
	}
	if bal.GetBalance() != 10 {
		t.Fatalf("double credit balance=%d", bal.GetBalance())
	}

	ledger, err := biz.Ledger(a, 0, 20)
	if err != nil {
		t.Fatalf("ledger: %v", err)
	}
	if len(ledger.GetList()) != 1 {
		t.Fatalf("ledger len=%d", len(ledger.GetList()))
	}
	row := ledger.GetList()[0]
	if row.GetChangeType() != apicommon.CoinChangeType_COIN_CHANGE_TYPE_RECHARGE || row.GetChange() != 10 {
		t.Fatalf("ledger row type=%d change=%d", row.GetChangeType(), row.GetChange())
	}
}

func TestPayWechatClosed(t *testing.T) {
	a, _, _ := setup(t)
	setTestCoin(t)

	order, err := biz.CreateRecharge(a, 10, "")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	_, err = biz.PayRecharge(a, order.GetId(), apicommon.PayType_PAY_TYPE_WECHAT)
	ce, ok := codeerror.As(err)
	if !ok || !ce.Is(bizerror.ErrorPayChannelClosed) {
		t.Fatalf("want channel closed, got %v", err)
	}
	bal, err := biz.Balance(a)
	if err != nil {
		t.Fatalf("balance: %v", err)
	}
	if bal.GetBalance() != 0 {
		t.Fatalf("balance=%d", bal.GetBalance())
	}
	got, err := biz.GetRecharge(a, order.GetId())
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.GetStatus() != apicommon.CoinOrderStatus_COIN_ORDER_STATUS_PENDING_PAY ||
		got.GetPayType() != apicommon.PayType_PAY_TYPE_UNSPECIFIED {
		t.Fatalf("order changed status=%d pay=%d", got.GetStatus(), got.GetPayType())
	}
}

func TestSendGiftDebitAndReplay(t *testing.T) {
	a, b, _ := setup(t)
	setTestCoin(t)
	if err := rechargeCoins(t, a, 10); err != nil {
		t.Fatal(err)
	}

	reply, pushes, err := biz.SendGift(a, b, 1, "giftmsg01")
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	if reply.GetBalance() != 9 {
		t.Fatalf("balance=%d", reply.GetBalance())
	}
	msg := reply.GetMsg()
	if msg.GetContentType() != apicommon.ContentType_CONTENT_TYPE_GIFT {
		t.Fatalf("content type=%d", msg.GetContentType())
	}
	if msg.GetContent().GetGift().GetGiftId() != 1 || msg.GetContent().GetGift().GetName() != "玫瑰" {
		t.Fatalf("gift content %#v", msg.GetContent().GetGift())
	}
	if len(pushes) != 2 {
		t.Fatalf("pushes=%d", len(pushes))
	}
	gotUsers := map[int64]bool{}
	for _, p := range pushes {
		gotUsers[p.UserID] = true
	}
	if !gotUsers[a] || !gotUsers[b] {
		t.Fatalf("push users %#v", pushes)
	}

	again, pushes, err := biz.SendGift(a, b, 1, "giftmsg01")
	if err != nil {
		t.Fatalf("replay: %v", err)
	}
	if again.GetMsg().GetMsgId() != msg.GetMsgId() {
		t.Fatalf("msg id changed %d %d", msg.GetMsgId(), again.GetMsg().GetMsgId())
	}
	if again.GetBalance() != 9 {
		t.Fatalf("replay balance=%d", again.GetBalance())
	}
	if len(pushes) != 0 {
		t.Fatalf("replay pushes=%d", len(pushes))
	}
}

func TestSendGiftInsufficient(t *testing.T) {
	a, b, _ := setup(t)
	setTestCoin(t)

	_, _, err := biz.SendGift(a, b, 1, "giftkey04")
	ce, ok := codeerror.As(err)
	if !ok || !ce.Is(bizerror.ErrorCoinInsufficient) {
		t.Fatalf("want insufficient, got %v", err)
	}
	rec, err := imdao.NewGiftRecordDao(nil).GetBySenderClient(a, "giftkey04")
	if err != nil {
		t.Fatalf("record: %v", err)
	}
	if rec != nil {
		t.Fatalf("gift record written: %#v", rec)
	}
	bal, err := biz.Balance(a)
	if err != nil {
		t.Fatalf("balance: %v", err)
	}
	if bal.GetBalance() != 0 {
		t.Fatalf("balance=%d", bal.GetBalance())
	}
}

func TestSendGiftIdempotencyConflict(t *testing.T) {
	a, b, _ := setup(t)
	setTestCoin(t)
	if err := rechargeCoins(t, a, 10); err != nil {
		t.Fatal(err)
	}

	first, _, err := biz.SendGift(a, b, 1, "giftidem1")
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	_, _, err = biz.SendGift(a, b, 2, "giftidem1")
	ce, ok := codeerror.As(err)
	if !ok || !ce.Is(bizerror.ErrorIdempotency) {
		t.Fatalf("want idempotency, got %v", err)
	}
	if first.GetBalance() != 9 {
		t.Fatalf("first balance=%d", first.GetBalance())
	}
	bal, err := biz.Balance(a)
	if err != nil {
		t.Fatalf("balance: %v", err)
	}
	if bal.GetBalance() != 9 {
		t.Fatalf("balance=%d", bal.GetBalance())
	}
}

func rechargeCoins(t *testing.T, userID, coins int64) error {
	t.Helper()
	order, err := biz.CreateRecharge(userID, coins, "")
	if err != nil {
		return err
	}
	if _, err := biz.PayRecharge(userID, order.GetId(), apicommon.PayType_PAY_TYPE_MOCK); err != nil {
		return err
	}
	_, err = biz.MockPaySuccess(userID, order.GetId())
	return err
}
