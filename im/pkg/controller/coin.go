package controller

import (
	"context"

	coinapi "mall/im/api/coin"
	"mall/im/pkg/biz"
	"mall/kit/auth"
	"mall/kit/logger"
)

// CoinController 金币余额、充值和送礼。
type CoinController struct{}

func NewCoinController() *CoinController {
	return &CoinController{}
}

func (c *CoinController) Balance(ctx context.Context, _ *coinapi.BalanceRequest) (*coinapi.BalanceReply, error) {
	a := auth.MustFromContext(ctx)
	reply, err := biz.Balance(a.UserID)
	if err != nil {
		logger.Errorf("Balance user=%d err=%v", a.UserID, err)
		return nil, err
	}
	return reply, nil
}

func (c *CoinController) Ledger(ctx context.Context, req *coinapi.LedgerRequest) (*coinapi.LedgerReply, error) {
	a := auth.MustFromContext(ctx)
	reply, err := biz.Ledger(a.UserID, req.GetBeforeId(), req.GetLimit())
	if err != nil {
		logger.Errorf("Ledger user=%d before=%d err=%v", a.UserID, req.GetBeforeId(), err)
		return nil, err
	}
	return reply, nil
}

func (c *CoinController) GiftList(ctx context.Context, _ *coinapi.GiftListRequest) (*coinapi.GiftListReply, error) {
	a := auth.MustFromContext(ctx)
	reply, err := biz.GiftList(a.UserID)
	if err != nil {
		logger.Errorf("GiftList user=%d err=%v", a.UserID, err)
		return nil, err
	}
	return reply, nil
}

func (c *CoinController) SendGift(ctx context.Context, req *coinapi.SendGiftRequest) (*coinapi.SendGiftReply, error) {
	a := auth.MustFromContext(ctx)
	reply, pushes, err := biz.SendGift(a.UserID, req.GetPeerUserId(), req.GetGiftId(), req.GetClientMsgId())
	if err != nil {
		logger.Errorf("SendGift user=%d peer=%d gift=%d err=%v", a.UserID, req.GetPeerUserId(), req.GetGiftId(), err)
		return nil, err
	}
	pushAll(pushes)
	return reply, nil
}

func (c *CoinController) CreateRecharge(
	ctx context.Context,
	req *coinapi.CreateRechargeRequest,
) (*coinapi.RechargeOrder, error) {
	a := auth.MustFromContext(ctx)
	reply, err := biz.CreateRecharge(a.UserID, req.GetCoins(), req.GetClientOrderId())
	if err != nil {
		logger.Errorf("CreateRecharge user=%d coins=%d err=%v", a.UserID, req.GetCoins(), err)
		return nil, err
	}
	return reply, nil
}

func (c *CoinController) PayRecharge(
	ctx context.Context,
	req *coinapi.PayRechargeRequest,
) (*coinapi.RechargeOrder, error) {
	a := auth.MustFromContext(ctx)
	reply, err := biz.PayRecharge(a.UserID, req.GetOrderId(), req.GetPayType())
	if err != nil {
		logger.Errorf("PayRecharge user=%d order=%d err=%v", a.UserID, req.GetOrderId(), err)
		return nil, err
	}
	return reply, nil
}

func (c *CoinController) MockPaySuccess(
	ctx context.Context,
	req *coinapi.MockPaySuccessRequest,
) (*coinapi.RechargeOrder, error) {
	a := auth.MustFromContext(ctx)
	reply, err := biz.MockPaySuccess(a.UserID, req.GetOrderId())
	if err != nil {
		logger.Errorf("MockPaySuccess user=%d order=%d err=%v", a.UserID, req.GetOrderId(), err)
		return nil, err
	}
	return reply, nil
}

func (c *CoinController) GetRecharge(
	ctx context.Context,
	req *coinapi.GetRechargeRequest,
) (*coinapi.RechargeOrder, error) {
	a := auth.MustFromContext(ctx)
	reply, err := biz.GetRecharge(a.UserID, req.GetOrderId())
	if err != nil {
		logger.Errorf("GetRecharge user=%d order=%d err=%v", a.UserID, req.GetOrderId(), err)
		return nil, err
	}
	return reply, nil
}

var _ coinapi.CoinHTTPServer = (*CoinController)(nil)
