export function formatYuanFromFen(fen: number): string {
  const n = Number(fen)
  if (!Number.isFinite(n)) {
    return '0.00'
  }
  return (n / 100).toFixed(2)
}

export function formatCoinChange(change: number): string {
  if (change > 0) {
    return `+${change}`
  }
  return String(change)
}

export function isGiftIconUrl(icon?: string): boolean {
  if (!icon) {
    return false
  }
  return icon.startsWith('http://') || icon.startsWith('https://') || icon.startsWith('/')
}

export function parseRouteId(param: unknown): number {
  const raw = Array.isArray(param) ? param[0] : param
  const n = typeof raw === 'number' ? raw : Number(raw)
  if (!Number.isSafeInteger(n) || n <= 0) {
    return 0
  }
  return n
}
