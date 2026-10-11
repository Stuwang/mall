/** 分 → 元展示 */
export function formatFenToYuan(fen: number | undefined | null): string {
  const value = Number(fen ?? 0)
  return (value / 100).toFixed(2)
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
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`
}
