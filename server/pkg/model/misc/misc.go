package misc

import apicommon "helloworld/api/common"

// Banner 首页轮播图（暂不支持跳转）
type Banner struct {
	ID       int64                `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	ImageURL string               `json:"image_url" gorm:"column:image_url"`
	Sort     int32                `json:"sort" gorm:"column:sort"`
	Status   apicommon.MiscStatus `json:"status" gorm:"column:status;index"`
	Ct       int64                `json:"ct" gorm:"column:ct"`
	Ut       int64                `json:"ut" gorm:"column:ut"`
}

func (Banner) TableName() string {
	return "misc_banner"
}

// Announcement 首页公告
type Announcement struct {
	ID      int64                `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Content string               `json:"content" gorm:"column:content;type:text"`
	Sort    int32                `json:"sort" gorm:"column:sort"`
	Status  apicommon.MiscStatus `json:"status" gorm:"column:status;index"`
	Ct      int64                `json:"ct" gorm:"column:ct"`
	Ut      int64                `json:"ut" gorm:"column:ut"`
}

func (Announcement) TableName() string {
	return "misc_announcement"
}
