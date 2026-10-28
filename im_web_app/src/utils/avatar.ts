const PALETTE = [
  '#0f8f86',
  '#2a6fdb',
  '#e8873a',
  '#c45c6a',
  '#5b6fd6',
  '#1f9d6a',
  '#d4a017',
  '#3d7ea6',
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
