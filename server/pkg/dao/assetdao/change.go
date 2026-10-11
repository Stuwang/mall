package assetdao

import (
	"time"

	"gorm.io/gorm"
	"helloworld/base/codeerror"
	"helloworld/common/database"
	assetmodel "helloworld/pkg/model/asset"
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
	if req.UserID <= 0 || req.Type <= 0 || req.Change == 0 || req.IdempotentID == "" {
		return nil, codeerror.OrSystem(codeerror.ErrorParam, "Change invalid req userID=%d type=%d change=%d idempotentID=%q", req.UserID, req.Type, req.Change, req.IdempotentID)
	}

	var result *assetmodel.UserAsset
	err := database.GetDb().Transaction(func(tx *gorm.DB) error {
		assetDao := NewAssetDao(tx)
		historyDao := NewHistoryDao(tx)

		exist, err := historyDao.GetByIdempotentID(req.IdempotentID)
		if err != nil {
			return err
		}
		if exist != nil {
			a, err := assetDao.GetByUserAndType(exist.UserID, exist.Type)
			if err != nil {
				return err
			}
			result = a
			return nil
		}

		now := time.Now().Unix()
		a, err := assetDao.FindByUserAndType(req.UserID, req.Type)
		if err != nil {
			return err
		}
		if a == nil {
			if req.Change < 0 {
				return codeerror.ErrorParam.Errorf("insufficient balance")
			}
			a = &assetmodel.UserAsset{
				UserID:  req.UserID,
				Type:    req.Type,
				Balance: req.Change,
				Ct:      now,
				Ut:      now,
			}
			if err := assetDao.Create(a); err != nil {
				return err
			}
		} else {
			next := a.Balance + req.Change
			if next < 0 {
				return codeerror.ErrorParam.Errorf("insufficient balance")
			}
			a.Balance = next
			if err := assetDao.UpdateBalance(a); err != nil {
				return err
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
