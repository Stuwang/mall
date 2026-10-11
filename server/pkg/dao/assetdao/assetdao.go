package assetdao

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"helloworld/base/codeerror"
	"helloworld/common/database"
	assetmodel "helloworld/pkg/model/asset"
)

func db() *gorm.DB {
	return database.GetDb()
}

// AutoMigrate 迁移资产相关表
func AutoMigrate() error {
	if err := db().AutoMigrate(&assetmodel.UserAsset{}, &assetmodel.UserAssetHistory{}); err != nil {
		return codeerror.OrSystem(err, "AutoMigrate asset tables")
	}
	return nil
}

// ListByUserID 查询用户资产；assetType=0 表示全部
func ListByUserID(userID int64, assetType int32) ([]assetmodel.UserAsset, error) {
	q := db().Where("user_id = ?", userID)
	if assetType > 0 {
		q = q.Where("type = ?", assetType)
	}
	var list []assetmodel.UserAsset
	if err := q.Order("type asc").Find(&list).Error; err != nil {
		return nil, codeerror.OrSystem(err, "ListByUserID userID=%d type=%d", userID, assetType)
	}
	return list, nil
}

// GetByUserAndType 查询指定类型资产
func GetByUserAndType(userID int64, assetType int32) (*assetmodel.UserAsset, error) {
	var a assetmodel.UserAsset
	err := db().Where("user_id = ? AND type = ?", userID, assetType).First(&a).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, codeerror.OrSystem(codeerror.ErrorParam.Errorf("asset not found"), "GetByUserAndType not found userID=%d type=%d", userID, assetType)
	}
	if err != nil {
		return nil, codeerror.OrSystem(err, "GetByUserAndType userID=%d type=%d", userID, assetType)
	}
	return &a, nil
}

// ListHistory 分页查询资产历史；assetType=0 表示全部
func ListHistory(userID int64, assetType int32, page, pageSize int) ([]assetmodel.UserAssetHistory, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	q := db().Model(&assetmodel.UserAssetHistory{}).Where("user_id = ?", userID)
	if assetType > 0 {
		q = q.Where("type = ?", assetType)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, codeerror.OrSystem(err, "ListHistory count userID=%d type=%d", userID, assetType)
	}

	var list []assetmodel.UserAssetHistory
	err := q.Order("id desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&list).Error
	if err != nil {
		return nil, 0, codeerror.OrSystem(err, "ListHistory userID=%d type=%d page=%d", userID, assetType, page)
	}
	return list, total, nil
}

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

// Change 变更资产并写入历史（事务，按 IdempotentID 幂等）
func Change(req ChangeReq) (*assetmodel.UserAsset, error) {
	if req.UserID <= 0 || req.Type <= 0 || req.Change == 0 || req.IdempotentID == "" {
		return nil, codeerror.OrSystem(codeerror.ErrorParam, "Change invalid req userID=%d type=%d change=%d idempotentID=%q", req.UserID, req.Type, req.Change, req.IdempotentID)
	}

	var result *assetmodel.UserAsset
	err := db().Transaction(func(tx *gorm.DB) error {
		// 幂等：已存在相同 idempotent_id 则直接返回当前资产，不再变更
		var exist assetmodel.UserAssetHistory
		err := tx.Where("idempotent_id = ?", req.IdempotentID).First(&exist).Error
		if err == nil {
			var a assetmodel.UserAsset
			if err := tx.Where("user_id = ? AND type = ?", exist.UserID, exist.Type).First(&a).Error; err != nil {
				return err
			}
			result = &a
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		now := time.Now().Unix()
		var a assetmodel.UserAsset
		err = tx.Where("user_id = ? AND type = ?", req.UserID, req.Type).First(&a).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if req.Change < 0 {
				return codeerror.ErrorParam.Errorf("insufficient balance")
			}
			a = assetmodel.UserAsset{
				UserID:  req.UserID,
				Type:    req.Type,
				Balance: req.Change,
				Ct:      now,
				Ut:      now,
			}
			if err := tx.Create(&a).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else {
			next := a.Balance + req.Change
			if next < 0 {
				return codeerror.ErrorParam.Errorf("insufficient balance")
			}
			a.Balance = next
			a.Ut = now
			if err := tx.Model(&a).Select("balance", "ut").Updates(&a).Error; err != nil {
				return err
			}
		}

		h := &assetmodel.UserAssetHistory{
			UserID:       req.UserID,
			Type:         req.Type,
			Change:       req.Change,
			Balance:      a.Balance,
			IdempotentID: req.IdempotentID,
			BizType:      req.BizType,
			BizID:        req.BizID,
			Remark:       req.Remark,
			Ct:           now,
		}
		if err := tx.Create(h).Error; err != nil {
			return err
		}
		result = &a
		return nil
	})
	if err != nil {
		return nil, codeerror.OrSystem(err, "Change userID=%d type=%d change=%d idempotentID=%s", req.UserID, req.Type, req.Change, req.IdempotentID)
	}
	return result, nil
}
