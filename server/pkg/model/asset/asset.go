package asset

import apicommon "helloworld/api/common"

// UserAsset 用户资产表
type UserAsset struct {
	ID      int64               `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	UserID  int64               `json:"user_id" gorm:"column:user_id;uniqueIndex:uk_user_type"`
	Type    apicommon.AssetType `json:"type" gorm:"column:type;uniqueIndex:uk_user_type"`
	Balance int64               `json:"balance" gorm:"column:balance"`
	Ct      int64               `json:"ct" gorm:"column:ct"`
	Ut      int64               `json:"ut" gorm:"column:ut"`
}

func (UserAsset) TableName() string {
	return "user_asset"
}

// UserAssetHistory 用户资产变更历史
type UserAssetHistory struct {
	ID           int64               `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	UserID       int64               `json:"user_id" gorm:"column:user_id;index:idx_user_type_ct"`
	Type         apicommon.AssetType `json:"type" gorm:"column:type;index:idx_user_type_ct"`
	Change       int64               `json:"change" gorm:"column:change"`
	Balance      int64               `json:"balance" gorm:"column:balance"`
	IdempotentID string              `json:"idempotent_id" gorm:"column:idempotent_id;uniqueIndex:uk_idempotent_id"`
	BizType      string              `json:"biz_type" gorm:"column:biz_type"`
	BizID        string              `json:"biz_id" gorm:"column:biz_id"`
	Remark       string              `json:"remark" gorm:"column:remark"`
	Ct           int64               `json:"ct" gorm:"column:ct;index:idx_user_type_ct"`
}

func (UserAssetHistory) TableName() string {
	return "user_asset_history"
}
