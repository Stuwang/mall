package assetdao

import (
	"time"

	"gorm.io/gorm"
	"helloworld/pkg/bizerror"
	assetmodel "helloworld/pkg/model/asset"
	"mall/kit/codeerror"
	"helloworld/common/database"
)

// ChangeReq 资产变更请求（内部能力，供后续发分/扣分使用）
type ChangeReq struct {
	UserID       int64
	Type         int32
	Change       int64
	IdempotentID string // 必填，用于幂等
	BizType      string
	BizID        string
	Remark       string
}

// Change 跨 asset / history 事务变更，按 IdempotentID 幂等
func Change(req ChangeReq) (*assetmodel.UserAsset, error) {
	var result *assetmodel.UserAsset
	err := database.GetDb().Transaction(func(tx *gorm.DB) error {
		a, err := ChangeTx(tx, req)
		if err != nil {
			return err
		}
		result = a
		return nil
	})
	if err != nil {
		return nil, codeerror.OrSystem(err, "Change userID=%d type=%d change=%d idempotentID=%s", req.UserID, req.Type, req.Change, req.IdempotentID)
	}
	return result, nil
}

// ChangeTx 在已有事务中变更资产（幂等）
func ChangeTx(tx *gorm.DB, req ChangeReq) (*assetmodel.UserAsset, error) {
	if tx == nil {
		return nil, codeerror.OrSystem(codeerror.ErrorParam, "ChangeTx nil tx")
	}
	if req.UserID <= 0 || req.Type <= 0 || req.Change == 0 || req.IdempotentID == "" {
		return nil, codeerror.OrSystem(codeerror.ErrorParam, "ChangeTx invalid req userID=%d type=%d change=%d idempotentID=%q", req.UserID, req.Type, req.Change, req.IdempotentID)
	}

	assetDao := NewAssetDao(tx)
	historyDao := NewHistoryDao(tx)

	exist, err := historyDao.GetByIdempotentID(req.IdempotentID)
	if err != nil {
		return nil, err
	}
	if exist != nil {
		return assetDao.GetByUserAndType(exist.UserID, exist.Type)
	}

	now := time.Now().Unix()
	a, err := assetDao.FindByUserAndType(req.UserID, req.Type)
	if err != nil {
		return nil, err
	}
	if a == nil {
		if req.Change < 0 {
			return nil, bizerror.ErrorAssetInsufficient
		}
		a = &assetmodel.UserAsset{
			UserID:  req.UserID,
			Type:    req.Type,
			Balance: req.Change,
			Ct:      now,
			Ut:      now,
		}
		if err := assetDao.Create(a); err != nil {
			return nil, err
		}
	} else {
		next := a.Balance + req.Change
		if next < 0 {
			return nil, bizerror.ErrorAssetInsufficient
		}
		a.Balance = next
		if err := assetDao.UpdateBalance(a); err != nil {
			return nil, err
		}
	}

	if err := historyDao.Create(&assetmodel.UserAssetHistory{
		UserID:       req.UserID,
		Type:         req.Type,
		Change:       req.Change,
		Balance:      a.Balance,
		IdempotentID: req.IdempotentID,
		BizType:      req.BizType,
		BizID:        req.BizID,
		Remark:       req.Remark,
		Ct:           now,
	}); err != nil {
		return nil, err
	}
	return a, nil
}
