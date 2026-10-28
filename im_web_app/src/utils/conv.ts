/** 与服务端一致：较小用户 id_较大用户 id */
export function makeConvId(a: number, b: number): string {
  const low = a < b ? a : b
  const high = a < b ? b : a
  return `${low}_${high}`
}

export function newClientMsgId(): string {
  const rand = Math.random().toString(36).slice(2, 10)
  return `c-${Date.now().toString(36)}-${rand}`.slice(0, 64)
}

export function newClientPostId(): string {
  const rand = Math.random().toString(36).slice(2, 10)
  return `p-${Date.now().toString(36)}-${rand}`.slice(0, 64)
}

export function newClientCommentId(): string {
  const rand = Math.random().toString(36).slice(2, 10)
  return `cm-${Date.now().toString(36)}-${rand}`.slice(0, 64)
}

export function newClientOrderId(): string {
  const rand = Math.random().toString(36).slice(2, 10)
  return `o-${Date.now().toString(36)}-${rand}`.slice(0, 64)
}

export function formatTime(sec: number): string {
  if (!sec) {
    return ''
  }
  const d = new Date(sec * 1000)
  const now = new Date()
  const sameDay =
    d.getFullYear() === now.getFullYear() &&
    d.getMonth() === now.getMonth() &&
    d.getDate() === now.getDate()
  const hh = String(d.getHours()).padStart(2, '0')
  const mm = String(d.getMinutes()).padStart(2, '0')
  if (sameDay) {
    return `${hh}:${mm}`
  }
  const md = `${d.getMonth() + 1}/${d.getDate()}`
  return `${md} ${hh}:${mm}`
}
