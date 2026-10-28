package biz

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	coinapi "mall/im/api/coin"
	apicommon "mall/im/api/common"
	"mall/im/common/database"
	"mall/im/pkg/bizerror"
	imdao "mall/im/pkg/dao"
	"mall/im/pkg/model"
	"mall/kit/codeerror"
	"mall/kit/logger"

	"gorm.io/gorm"
)

const (
	maxRechargeCoins   int64 = 1000000
	ledgerDefaultLimit       = 20
	ledgerMaxLimit           = 50
	orderNoAttempts          = 5
	remarkRecharge           = "充值到账"
	remarkGiftPrefix         = "送出礼物 "
	giftPreviewPrefix        = "[礼物]"
)

// GiftItem 配置里的一份礼物。
type GiftItem struct {
	ID     int64
	Name   string
	Icon   string
	Effect string
	Price  int64
}

// CoinConfig 金币汇率、支付超时和礼物列表。
type CoinConfig struct {
	CoinsPerYuan  int64
	PayTimeoutSec int64
	Gifts         []GiftItem
}

var (
	coinMu  sync.RWMutex
	coinCfg CoinConfig
)

// SetCoinConfig 由进程启动时写入金币配置。未走 config.Load 的测试也要调用。
func SetCoinConfig(cfg CoinConfig) {
	gifts := make([]GiftItem, len(cfg.Gifts))
	copy(gifts, cfg.Gifts)
	coinMu.Lock()
	coinCfg = CoinConfig{
		CoinsPerYuan:  cfg.CoinsPerYuan,
		PayTimeoutSec: cfg.PayTimeoutSec,
		Gifts:         gifts,
	}
	coinMu.Unlock()
}

func coinConfig() CoinConfig {
	coinMu.RLock()
	defer coinMu.RUnlock()
	return coinCfg
}

func findGift(id int64) (GiftItem, bool) {
	cfg := coinConfig()
	for _, g := range cfg.Gifts {
		if g.ID == id {
			return g, true
		}
	}
	return GiftItem{}, false
}

type coinChange struct {
	UserID       int64
	Change       int64
	ChangeType   apicommon.CoinChangeType
	IdempotentID string
	Remark       string
	Ext          string
}

type rechargeExt struct {
	OrderNo string            `json:"order_no"`
	PayType apicommon.PayType `json:"pay_type"`
}

type giftExt struct {
	GiftID     int64  `json:"gift_id"`
	PeerUserID int64  `json:"peer_user_id"`
	GiftName   string `json:"gift_name"`
}

// Balance 返回当前用户金币余额。没有账户行时余额为 0。
func Balance(userID int64) (*coinapi.BalanceReply, error) {
	if userID <= 0 {
		logger.Errorf("Balance user invalid user=%d", userID)
		return nil, codeerror.ErrorParam
	}
	bal, err := readBalance(nil, userID)
	if err != nil {
		logger.Errorf("Balance user=%d err=%v", userID, err)
		return nil, err
	}
	return &coinapi.BalanceReply{
		Balance:      bal,
		CoinsPerYuan: coinConfig().CoinsPerYuan,
	}, nil
}

// Ledger 按 id 倒序返回流水。beforeID 为 0 表示从最新开始。
func Ledger(userID, beforeID int64, limit int32) (*coinapi.LedgerReply, error) {
	if userID <= 0 {
		logger.Errorf("Ledger user invalid user=%d", userID)
		return nil, codeerror.ErrorParam
	}
	if beforeID < 0 {
		logger.Errorf("Ledger before_id invalid user=%d before=%d", userID, beforeID)
		return nil, codeerror.ErrorParam
	}
	page := ledgerLimit(limit)
	rows, err := imdao.NewCoinLedgerDao(nil).ListDesc(userID, beforeID, page+1)
	if err != nil {
		logger.Errorf("Ledger user=%d before=%d err=%v", userID, beforeID, err)
		return nil, err
	}
	hasMore := len(rows) > page
	if hasMore {
		rows = rows[:page]
	}
	list := make([]*coinapi.LedgerItem, 0, len(rows))
	for i := range rows {
		row := &rows[i]
		list = append(list, &coinapi.LedgerItem{
			Id:         row.ID,
			ChangeType: row.ChangeType,
			Change:     row.Change,
			Balance:    row.Balance,
			Remark:     row.Remark,
			Ext:        row.Ext,
			Ct:         row.Ct,
		})
	}
	return &coinapi.LedgerReply{List: list, HasMore: hasMore}, nil
}

// GiftList 按配置顺序返回礼物。
func GiftList(userID int64) (*coinapi.GiftListReply, error) {
	if userID <= 0 {
		logger.Errorf("GiftList user invalid user=%d", userID)
		return nil, codeerror.ErrorParam
	}
	cfg := coinConfig()
	list := make([]*coinapi.Gift, 0, len(cfg.Gifts))
	for _, g := range cfg.Gifts {
		list = append(list, &coinapi.Gift{
			Id:     g.ID,
			Name:   g.Name,
			Icon:   g.Icon,
			Effect: g.Effect,
			Price:  g.Price,
		})
	}
	return &coinapi.GiftListReply{List: list}, nil
}

// CreateRecharge 创建待支付充值订单，并把当时汇率写入订单。
func CreateRecharge(userID, coins int64, clientOrderID string) (*coinapi.RechargeOrder, error) {
	if userID <= 0 {
		logger.Errorf("CreateRecharge user invalid user=%d", userID)
		return nil, codeerror.ErrorParam
	}
	cfg := coinConfig()
	if coins < 1 || coins > maxRechargeCoins {
		logger.Errorf("CreateRecharge coins out of range user=%d coins=%d", userID, coins)
		return nil, codeerror.ErrorParam
	}
	if cfg.CoinsPerYuan <= 0 || (coins*100)%cfg.CoinsPerYuan != 0 {
		logger.Errorf("CreateRecharge cash indivisible user=%d coins=%d rate=%d", userID, coins, cfg.CoinsPerYuan)
		return nil, codeerror.ErrorParam
	}
	if cfg.PayTimeoutSec <= 0 {
		logger.Errorf("CreateRecharge timeout invalid user=%d", userID)
		return nil, codeerror.ErrorSystem
	}
	cash := coins * 100 / cfg.CoinsPerYuan
	dao := imdao.NewCoinOrderDao(nil)
	if clientOrderID != "" {
		if err := validateIdempotencyKey(clientOrderID); err != nil {
			return nil, err
		}
		existing, err := dao.GetByClient(userID, clientOrderID)
		if err != nil {
			logger.Errorf("CreateRecharge get client user=%d err=%v", userID, err)
			return nil, err
		}
		if existing != nil {
			return sameCoinsOrder(existing, coins)
		}
	}
	now := time.Now().Unix()
	order := &model.CoinOrder{
		UserID:        userID,
		ClientOrderID: clientOrderID,
		Coins:         coins,
		CashAmount:    cash,
		CoinsPerYuan:  cfg.CoinsPerYuan,
		PayType:       apicommon.PayType_PAY_TYPE_UNSPECIFIED,
		Status:        apicommon.CoinOrderStatus_COIN_ORDER_STATUS_PENDING_PAY,
		ExpireAt:      now + cfg.PayTimeoutSec,
		Ct:            now,
		Ut:            now,
	}
	for range orderNoAttempts {
		order.ID = 0
		no, err := newOrderNo()
		if err != nil {
			return nil, codeerror.OrSystem(err, "CreateRecharge order no user=%d", userID)
		}
		order.OrderNo = no
		err = dao.Create(order)
		if err == nil {
			return toRechargeOrder(order), nil
		}
		if !uniqueConflict(err) {
			logger.Errorf("CreateRecharge create user=%d err=%v", userID, err)
			return nil, err
		}
		if clientOrderID != "" {
			existing, gerr := dao.GetByClient(userID, clientOrderID)
			if gerr != nil {
				logger.Errorf("CreateRecharge replay get user=%d err=%v", userID, gerr)
				return nil, gerr
			}
			if existing != nil {
				return sameCoinsOrder(existing, coins)
			}
		}
		logger.Errorf("CreateRecharge order_no conflict user=%d order=%s", userID, order.OrderNo)
	}
	logger.Errorf("CreateRecharge order_no retries exhausted user=%d", userID)
	return nil, codeerror.ErrorSystem
}

// PayRecharge 选择支付渠道。微信和支付宝未开通，不改订单。
func PayRecharge(userID, orderID int64, payType apicommon.PayType) (*coinapi.RechargeOrder, error) {
	order, err := loadOwnOrder(userID, orderID)
	if err != nil {
		logger.Errorf("PayRecharge user=%d order=%d err=%v", userID, orderID, err)
		return nil, err
	}
	switch order.Status {
	case apicommon.CoinOrderStatus_COIN_ORDER_STATUS_PAID:
		return toRechargeOrder(order), nil
	case apicommon.CoinOrderStatus_COIN_ORDER_STATUS_CANCELLED:
		logger.Errorf("PayRecharge cancelled user=%d order=%d", userID, orderID)
		return nil, bizerror.ErrorCoinOrderStatus
	case apicommon.CoinOrderStatus_COIN_ORDER_STATUS_PENDING_PAY:
		if time.Now().Unix() > order.ExpireAt {
			if err := cancelCoinOrder(order); err != nil {
				logger.Errorf("PayRecharge cancel user=%d order=%d err=%v", userID, orderID, err)
				return nil, err
			}
			logger.Errorf("PayRecharge expired user=%d order=%d", userID, orderID)
			return nil, bizerror.ErrorCoinOrderExpired
		}
	default:
		logger.Errorf("PayRecharge status user=%d order=%d status=%d", userID, orderID, order.Status)
		return nil, bizerror.ErrorCoinOrderStatus
	}
	switch payType {
	case apicommon.PayType_PAY_TYPE_WECHAT, apicommon.PayType_PAY_TYPE_ALIPAY:
		logger.Errorf("PayRecharge channel closed user=%d order=%d payType=%d", userID, orderID, payType)
		return nil, bizerror.ErrorPayChannelClosed
	case apicommon.PayType_PAY_TYPE_MOCK:
		order.PayType = apicommon.PayType_PAY_TYPE_MOCK
		order.Ut = time.Now().Unix()
		if err := imdao.NewCoinOrderDao(nil).Save(order); err != nil {
			logger.Errorf("PayRecharge save user=%d order=%d err=%v", userID, orderID, err)
			return nil, err
		}
		return toRechargeOrder(order), nil
	default:
		logger.Errorf("PayRecharge pay_type invalid user=%d order=%d payType=%d", userID, orderID, payType)
		return nil, codeerror.ErrorParam
	}
}

// MockPaySuccess 模拟支付成功并入账一次。
func MockPaySuccess(userID, orderID int64) (*coinapi.RechargeOrder, error) {
	order, err := loadOwnOrder(userID, orderID)
	if err != nil {
		logger.Errorf("MockPaySuccess user=%d order=%d err=%v", userID, orderID, err)
		return nil, err
	}
	if order.Status == apicommon.CoinOrderStatus_COIN_ORDER_STATUS_PAID &&
		order.PayType == apicommon.PayType_PAY_TYPE_MOCK {
		return toRechargeOrder(order), nil
	}
	if order.PayType != apicommon.PayType_PAY_TYPE_MOCK {
		logger.Errorf("MockPaySuccess pay_type user=%d order=%d payType=%d", userID, orderID, order.PayType)
		return nil, bizerror.ErrorCoinOrderStatus
	}
	if order.Status != apicommon.CoinOrderStatus_COIN_ORDER_STATUS_PENDING_PAY {
		logger.Errorf("MockPaySuccess status user=%d order=%d status=%d", userID, orderID, order.Status)
		return nil, bizerror.ErrorCoinOrderStatus
	}
	if time.Now().Unix() > order.ExpireAt {
		if err := cancelCoinOrder(order); err != nil {
			logger.Errorf("MockPaySuccess cancel user=%d order=%d err=%v", userID, orderID, err)
			return nil, err
		}
		logger.Errorf("MockPaySuccess expired user=%d order=%d", userID, orderID)
		return nil, bizerror.ErrorCoinOrderExpired
	}

	var saved *model.CoinOrder
	err = database.GetDb().Transaction(func(tx *gorm.DB) error {
		cur, err := imdao.NewCoinOrderDao(tx).GetByID(order.ID)
		if err != nil {
			return err
		}
		if cur == nil || cur.UserID != userID {
			logger.Errorf("MockPaySuccess missing in tx user=%d order=%d", userID, orderID)
			return bizerror.ErrorCoinOrderNotFound
		}
		if cur.Status == apicommon.CoinOrderStatus_COIN_ORDER_STATUS_PAID &&
			cur.PayType == apicommon.PayType_PAY_TYPE_MOCK {
			saved = cur
			return nil
		}
		if cur.PayType != apicommon.PayType_PAY_TYPE_MOCK ||
			cur.Status != apicommon.CoinOrderStatus_COIN_ORDER_STATUS_PENDING_PAY {
			logger.Errorf(
				"MockPaySuccess status in tx user=%d order=%d status=%d pay=%d",
				userID,
				orderID,
				cur.Status,
				cur.PayType,
			)
			return bizerror.ErrorCoinOrderStatus
		}
		if time.Now().Unix() > cur.ExpireAt {
			logger.Errorf("MockPaySuccess expired in tx user=%d order=%d", userID, orderID)
			return bizerror.ErrorCoinOrderExpired
		}
		extBytes, err := json.Marshal(rechargeExt{OrderNo: cur.OrderNo, PayType: cur.PayType})
		if err != nil {
			return codeerror.OrSystem(err, "MockPaySuccess ext order=%d", orderID)
		}
		if _, _, err := applyCoinChange(tx, coinChange{
			UserID:       cur.UserID,
			Change:       cur.Coins,
			ChangeType:   apicommon.CoinChangeType_COIN_CHANGE_TYPE_RECHARGE,
			IdempotentID: rechargeIdempotent(cur.OrderNo),
			Remark:       remarkRecharge,
			Ext:          string(extBytes),
		}); err != nil {
			return err
		}
		now := time.Now().Unix()
		cur.Status = apicommon.CoinOrderStatus_COIN_ORDER_STATUS_PAID
		cur.PaidAt = now
		cur.Ut = now
		if err := imdao.NewCoinOrderDao(tx).Save(cur); err != nil {
			return err
		}
		saved = cur
		return nil
	})
	if err != nil {
		logger.Errorf("MockPaySuccess tx user=%d order=%d err=%v", userID, orderID, err)
		return nil, err
	}
	if saved == nil {
		logger.Errorf("MockPaySuccess empty order user=%d order=%d", userID, orderID)
		return nil, codeerror.ErrorSystem
	}
	return toRechargeOrder(saved), nil
}

// GetRecharge 读取调用者自己的充值订单。
func GetRecharge(userID, orderID int64) (*coinapi.RechargeOrder, error) {
	order, err := loadOwnOrder(userID, orderID)
	if err != nil {
		logger.Errorf("GetRecharge user=%d order=%d err=%v", userID, orderID, err)
		return nil, err
	}
	return toRechargeOrder(order), nil
}

// SendGift 扣发送者金币并写入礼物消息。不给接收者加币。
func SendGift(senderID, peerUserID, giftID int64, clientMsgID string) (*coinapi.SendGiftReply, []PushItem, error) {
	if senderID <= 0 {
		logger.Errorf("SendGift sender invalid sender=%d", senderID)
		return nil, nil, codeerror.ErrorParam
	}
	if err := validatePeer(senderID, peerUserID); err != nil {
		return nil, nil, err
	}
	if err := validateIdempotencyKey(clientMsgID); err != nil {
		return nil, nil, err
	}
	if err := requireUser(peerUserID); err != nil {
		return nil, nil, err
	}
	gift, giftOK := findGift(giftID)

	var (
		balance int64
		msg     *model.Message
		pushes  []PushItem
	)
	err := database.GetDb().Transaction(func(tx *gorm.DB) error {
		rec, err := imdao.NewGiftRecordDao(tx).GetBySenderClient(senderID, clientMsgID)
		if err != nil {
			return err
		}
		if rec != nil {
			if rec.PeerUserID != peerUserID || rec.GiftID != giftID {
				logger.Errorf("SendGift idempotency sender=%d key=%s", senderID, clientMsgID)
				return bizerror.ErrorIdempotency
			}
			old, err := imdao.NewMessageDao(tx).GetByID(rec.MsgID)
			if err != nil {
				return err
			}
			if old == nil {
				logger.Errorf("SendGift record without message sender=%d msg=%d", senderID, rec.MsgID)
				return codeerror.ErrorSystem
			}
			bal, err := readBalance(tx, senderID)
			if err != nil {
				return err
			}
			balance = bal
			msg = old
			pushes = nil
			return nil
		}
		if !giftOK {
			logger.Errorf("SendGift gift not found sender=%d gift=%d", senderID, giftID)
			return bizerror.ErrorGiftNotFound
		}
		if gift.Price <= 0 {
			logger.Errorf("SendGift gift price invalid sender=%d gift=%d price=%d", senderID, giftID, gift.Price)
			return codeerror.ErrorParam
		}
		extBytes, err := json.Marshal(giftExt{
			GiftID:     gift.ID,
			PeerUserID: peerUserID,
			GiftName:   gift.Name,
		})
		if err != nil {
			return codeerror.OrSystem(err, "SendGift ext sender=%d", senderID)
		}
		bal, _, err := applyCoinChange(tx, coinChange{
			UserID:       senderID,
			Change:       -gift.Price,
			ChangeType:   apicommon.CoinChangeType_COIN_CHANGE_TYPE_GIFT_SEND,
			IdempotentID: giftIdempotent(senderID, clientMsgID),
			Remark:       remarkGiftPrefix + gift.Name,
			Ext:          string(extBytes),
		})
		if err != nil {
			return err
		}
		balance = bal
		content, err := encodeGiftContent(
			gift.ID,
			gift.Name,
			gift.Icon,
			gift.Effect,
			gift.Price,
		)
		if err != nil {
			return codeerror.OrSystem(err, "SendGift encode gift=%d", gift.ID)
		}
		res, err := createMessage(tx, createParams{
			SenderID:           senderID,
			PeerUserID:         peerUserID,
			Origin:             apicommon.Origin_ORIGIN_USER,
			Visibility:         apicommon.Visibility_VISIBILITY_BOTH,
			SenderPreview:      true,
			PeerPreview:        true,
			IdempotencyKey:     clientMsgID,
			ContentType:        apicommon.ContentType_CONTENT_TYPE_GIFT,
			Content:            content,
			CompareFingerprint: strconv.FormatInt(gift.ID, 10),
			CompareContentTy:   apicommon.ContentType_CONTENT_TYPE_GIFT,
			Preview:            PreviewText(giftPreviewPrefix+gift.Name, 50),
		})
		if err != nil {
			return err
		}
		if res.Msg == nil {
			logger.Errorf("SendGift createMessage empty sender=%d", senderID)
			return codeerror.ErrorSystem
		}
		msg = res.Msg
		pushes = res.Pushes
		row := &model.GiftRecord{
			SenderID:    senderID,
			ClientMsgID: clientMsgID,
			PeerUserID:  peerUserID,
			GiftID:      gift.ID,
			GiftName:    gift.Name,
			GiftIcon:    gift.Icon,
			GiftEffect:  gift.Effect,
			Price:       gift.Price,
			MsgID:       res.Msg.ID,
			Ct:          time.Now().Unix(),
		}
		if err := imdao.NewGiftRecordDao(tx).Create(row); err != nil {
			if uniqueConflict(err) {
				logger.Errorf("SendGift gift record unique sender=%d key=%s err=%v", senderID, clientMsgID, err)
				return bizerror.ErrorIdempotency
			}
			return err
		}
		return nil
	})
	if err != nil {
		logger.Errorf("SendGift tx sender=%d peer=%d gift=%d err=%v", senderID, peerUserID, giftID, err)
		return nil, nil, err
	}
	if msg == nil {
		logger.Errorf("SendGift empty message sender=%d peer=%d", senderID, peerUserID)
		return nil, nil, codeerror.ErrorSystem
	}
	return &coinapi.SendGiftReply{
		Balance: balance,
		Msg:     toAPIMsg(msg, senderID),
	}, pushes, nil
}

func applyCoinChange(tx *gorm.DB, in coinChange) (int64, bool, error) {
	ledgerDao := imdao.NewCoinLedgerDao(tx)
	if in.IdempotentID != "" {
		exist, err := ledgerDao.GetByIdempotentID(in.IdempotentID)
		if err != nil {
			return 0, false, err
		}
		if exist != nil {
			bal, err := readBalance(tx, in.UserID)
			if err != nil {
				return 0, false, err
			}
			return bal, false, nil
		}
	}
	if in.UserID <= 0 || in.Change == 0 || in.IdempotentID == "" {
		logger.Errorf("applyCoinChange param user=%d change=%d idempotent=%q", in.UserID, in.Change, in.IdempotentID)
		return 0, false, codeerror.ErrorParam
	}

	coinDao := imdao.NewUserCoinDao(tx)
	row, err := coinDao.Get(in.UserID)
	if err != nil {
		return 0, false, err
	}
	now := time.Now().Unix()
	var balance int64
	if row == nil {
		if in.Change < 0 {
			logger.Errorf("applyCoinChange insufficient user=%d change=%d", in.UserID, in.Change)
			return 0, false, bizerror.ErrorCoinInsufficient
		}
		created := &model.UserCoin{
			UserID:  in.UserID,
			Balance: in.Change,
			Ct:      now,
			Ut:      now,
		}
		if err := coinDao.Create(created); err != nil {
			if uniqueConflict(err) {
				logger.Errorf("applyCoinChange balance unique user=%d err=%v", in.UserID, err)
				return 0, false, codeerror.ErrorSystem
			}
			return 0, false, err
		}
		balance = in.Change
	} else {
		n, err := coinDao.AddBalance(in.UserID, in.Change, now)
		if err != nil {
			return 0, false, err
		}
		if n == 0 {
			logger.Errorf("applyCoinChange insufficient user=%d change=%d balance=%d", in.UserID, in.Change, row.Balance)
			return 0, false, bizerror.ErrorCoinInsufficient
		}
		got, err := coinDao.Get(in.UserID)
		if err != nil {
			return 0, false, err
		}
		if got == nil {
			logger.Errorf("applyCoinChange missing balance after update user=%d", in.UserID)
			return 0, false, codeerror.ErrorSystem
		}
		balance = got.Balance
	}

	entry := &model.CoinLedger{
		UserID:       in.UserID,
		ChangeType:   in.ChangeType,
		Change:       in.Change,
		Balance:      balance,
		IdempotentID: in.IdempotentID,
		Ext:          in.Ext,
		Remark:       in.Remark,
		Ct:           now,
	}
	if err := ledgerDao.Create(entry); err != nil {
		if uniqueConflict(err) {
			logger.Errorf("applyCoinChange ledger unique user=%d idempotent=%s err=%v", in.UserID, in.IdempotentID, err)
			return 0, false, codeerror.ErrorSystem
		}
		return 0, false, err
	}
	return balance, true, nil
}

func readBalance(tx *gorm.DB, userID int64) (int64, error) {
	row, err := imdao.NewUserCoinDao(tx).Get(userID)
	if err != nil {
		return 0, err
	}
	if row == nil {
		return 0, nil
	}
	return row.Balance, nil
}

func loadOwnOrder(userID, orderID int64) (*model.CoinOrder, error) {
	if userID <= 0 {
		logger.Errorf("coin order user invalid user=%d", userID)
		return nil, codeerror.ErrorParam
	}
	if orderID <= 0 {
		logger.Errorf("coin order id invalid user=%d order=%d", userID, orderID)
		return nil, bizerror.ErrorCoinOrderNotFound
	}
	order, err := imdao.NewCoinOrderDao(nil).GetByID(orderID)
	if err != nil {
		logger.Errorf("coin order get user=%d order=%d err=%v", userID, orderID, err)
		return nil, err
	}
	if order == nil || order.UserID != userID {
		logger.Errorf("coin order not found user=%d order=%d", userID, orderID)
		return nil, bizerror.ErrorCoinOrderNotFound
	}
	return order, nil
}

func cancelCoinOrder(order *model.CoinOrder) error {
	order.Status = apicommon.CoinOrderStatus_COIN_ORDER_STATUS_CANCELLED
	order.Ut = time.Now().Unix()
	if err := imdao.NewCoinOrderDao(nil).Save(order); err != nil {
		logger.Errorf("cancelCoinOrder id=%d err=%v", order.ID, err)
		return err
	}
	return nil
}

func sameCoinsOrder(existing *model.CoinOrder, coins int64) (*coinapi.RechargeOrder, error) {
	if existing.Coins != coins {
		logger.Errorf(
			"CreateRecharge idempotency user=%d client=%s coins=%d existing=%d",
			existing.UserID,
			existing.ClientOrderID,
			coins,
			existing.Coins,
		)
		return nil, bizerror.ErrorIdempotency
	}
	return toRechargeOrder(existing), nil
}

func toRechargeOrder(order *model.CoinOrder) *coinapi.RechargeOrder {
	if order == nil {
		return nil
	}
	return &coinapi.RechargeOrder{
		Id:           order.ID,
		OrderNo:      order.OrderNo,
		Coins:        order.Coins,
		CashAmount:   order.CashAmount,
		CoinsPerYuan: order.CoinsPerYuan,
		PayType:      order.PayType,
		Status:       order.Status,
		ExpireAt:     order.ExpireAt,
		PaidAt:       order.PaidAt,
		Ct:           order.Ct,
	}
}

func ledgerLimit(limit int32) int {
	if limit <= 0 {
		return ledgerDefaultLimit
	}
	if int(limit) > ledgerMaxLimit {
		return ledgerMaxLimit
	}
	return int(limit)
}

func newOrderNo() (string, error) {
	buf := make([]byte, 4)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return fmt.Sprintf("C%d%s", time.Now().UnixMilli(), hex.EncodeToString(buf)), nil
}

func rechargeIdempotent(orderNo string) string {
	return "recharge:" + orderNo
}

func giftIdempotent(userID int64, clientMsgID string) string {
	return fmt.Sprintf("gift:%d:%s", userID, clientMsgID)
}

func uniqueConflict(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "UNIQUE constraint failed") || strings.Contains(msg, "PRIMARY KEY constraint failed")
}
