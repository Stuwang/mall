/** 金额（分）→ 可抵扣积分；points_per_yuan：1 元 = N 积分 */
export function centsToPoints(cents: number, pointsPerYuan: number): number {
  if (cents <= 0 || pointsPerYuan <= 0) return 0
  return Math.floor((cents * pointsPerYuan) / 100)
}

/** 积分 → 抵扣金额（分） */
export function pointsToCents(points: number, pointsPerYuan: number): number {
  if (points <= 0 || pointsPerYuan <= 0) return 0
  return Math.floor((points * 100) / pointsPerYuan)
}

/** 本单最多可用积分 */
export function maxUsablePoints(
  totalAmountFen: number,
  pointsPerYuan: number,
  balance: number,
): number {
  const byAmount = centsToPoints(totalAmountFen, pointsPerYuan)
  return Math.max(0, Math.min(byAmount, Math.floor(balance)))
}
