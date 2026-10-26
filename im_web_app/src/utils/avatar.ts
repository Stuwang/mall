const PALETTE = [
  '#1989fa',
  '#07c160',
  '#ff976a',
  '#ee0a24',
  '#7232dd',
  '#0fc6c2',
  '#f5a623',
  '#2f54eb',
]

/** 由用户 id 得到稳定底色 */
export function avatarColor(userId: number): string {
  const id = Math.abs(userId || 0)
  return PALETTE[id % PALETTE.length]
}

/** 头像上显示的字：昵称首字，否则用 id 末位 */
export function avatarLabel(nickname: string | undefined, userId: number): string {
  const name = (nickname || '').trim()
  if (name) {
    return name.slice(0, 1).toUpperCase()
  }
  if (userId > 0) {
    return String(userId % 10)
  }
  return '?'
}
