package common

// DeletedFlag 软删除标记
type DeletedFlag int32

const (
	DeletedNo  DeletedFlag = 0 // 未删除
	DeletedYes DeletedFlag = 1 // 已删除
)
