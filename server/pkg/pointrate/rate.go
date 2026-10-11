package pointrate

// PointsToCents 积分转金额（分）；points_per_yuan 表示 1 元（100 分）对应多少积分
func PointsToCents(points, pointsPerYuan int64) int64 {
	if points <= 0 || pointsPerYuan <= 0 {
		return 0
	}
	return points * 100 / pointsPerYuan
}

// CentsToPoints 金额（分）转积分
func CentsToPoints(cents, pointsPerYuan int64) int64 {
	if cents <= 0 || pointsPerYuan <= 0 {
		return 0
	}
	return cents * pointsPerYuan / 100
}
