/** 分 → 元展示 */
export function formatFenToYuan(fen: number | undefined | null): string {
  const value = Number(fen ?? 0)
  return (value / 100).toFixed(2)
}

/** 元输入 → 分 */
export function yuanToFen(yuan: number | string): number {
  const n = typeof yuan === 'string' ? Number(yuan) : yuan
  if (!Number.isFinite(n)) {
    return 0
  }
  return Math.round(n * 100)
}

export function formatUnix(ts: number | undefined | null): string {
  if (!ts) {
    return '-'
  }
  const date = new Date(ts * 1000)
  if (Number.isNaN(date.getTime())) {
    return '-'
  }
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`
}
