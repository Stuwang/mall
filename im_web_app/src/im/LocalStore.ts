import type { ImLocalState, LocalConv, LocalMsg } from './types'

function storageKey(userId: number): string {
  return `im.web.local.${userId}`
}

function emptyState(): ImLocalState {
  return { list_cursor: 0, convs: {}, messages: {} }
}

export class LocalStore {
  private userId = 0
  private state: ImLocalState = emptyState()

  bindUser(userId: number): void {
    this.userId = userId
    this.state = this.load()
  }

  /** 仅卸下内存态，保留磁盘缓存供下次登录。 */
  unbind(): void {
    this.userId = 0
    this.state = emptyState()
  }

  getListCursor(): number {
    return this.state.list_cursor
  }

  setListCursor(cursor: number): void {
    this.state.list_cursor = cursor
    this.persist()
  }

  listConvs(): LocalConv[] {
    return Object.values(this.state.convs).sort((a, b) => {
      if (b.sort_ut !== a.sort_ut) {
        return b.sort_ut - a.sort_ut
      }
      return (b.last_msg_id || 0) - (a.last_msg_id || 0)
    })
  }

  getConv(convId: string): LocalConv | undefined {
    return this.state.convs[convId]
  }

  upsertConv(conv: LocalConv): void {
    const prev = this.state.convs[conv.conv_id]
    const peer = {
      user_id: conv.peer?.user_id || prev?.peer?.user_id || 0,
      nickname: conv.peer?.nickname || prev?.peer?.nickname || '',
      avatar: conv.peer?.avatar || prev?.peer?.avatar || '',
    }
    this.state.convs[conv.conv_id] = {
      ...prev,
      ...conv,
      peer,
      history_cursor: conv.history_cursor ?? prev?.history_cursor ?? 0,
      // 服务端快照不得冲掉本机已读水位
      read_seq: conv.read_seq ?? prev?.read_seq ?? 0,
    }
    this.persist()
  }

  ensureLocalConv(conv: LocalConv): void {
    const prev = this.state.convs[conv.conv_id]
    if (prev) {
      if (conv.peer?.avatar || conv.peer?.nickname) {
        this.upsertConv({
          ...prev,
          peer: {
            user_id: prev.peer?.user_id || conv.peer.user_id,
            nickname: conv.peer.nickname || prev.peer?.nickname || '',
            avatar: conv.peer.avatar || prev.peer?.avatar || '',
          },
        })
      }
      return
    }
    this.state.convs[conv.conv_id] = {
      ...conv,
      history_cursor: conv.history_cursor ?? 0,
      read_seq: conv.read_seq ?? 0,
      local_only: true,
    }
    this.persist()
  }

  /** 用在线列表等来源补全会话对方昵称/头像 */
  patchPeerProfile(peerUserId: number, nickname?: string, avatar?: string): void {
    const conv = Object.values(this.state.convs).find((c) => c.peer?.user_id === peerUserId)
    if (!conv?.peer) {
      return
    }
    const nextNick = nickname || conv.peer.nickname || ''
    const nextAvatar = avatar || conv.peer.avatar || ''
    if (nextNick === (conv.peer.nickname || '') && nextAvatar === (conv.peer.avatar || '')) {
      return
    }
    this.upsertConv({
      ...conv,
      peer: {
        ...conv.peer,
        nickname: nextNick,
        avatar: nextAvatar,
      },
    })
  }

  getMessages(convId: string): LocalMsg[] {
    const list = this.state.messages[convId] ?? []
    return [...list].sort((a, b) => {
      if (a.ct !== b.ct) {
        return a.ct - b.ct
      }
      return (a.msg_id || 0) - (b.msg_id || 0)
    })
  }

  upsertMessage(convId: string, msg: LocalMsg): void {
    const list = this.state.messages[convId] ? [...this.state.messages[convId]] : []
    let idx = -1
    if (msg.msg_id > 0) {
      idx = list.findIndex((m) => m.msg_id === msg.msg_id)
    }
    if (idx < 0 && msg.client_msg_id) {
      idx = list.findIndex((m) => m.client_msg_id === msg.client_msg_id)
    }
    if (idx >= 0) {
      list[idx] = { ...list[idx], ...msg, pending: false, failed: false }
    } else {
      list.push(msg)
    }
    this.state.messages[convId] = list
    this.persist()
  }

  markFailed(convId: string, clientMsgId: string): void {
    const list = this.state.messages[convId]
    if (!list) {
      return
    }
    const idx = list.findIndex((m) => m.client_msg_id === clientMsgId)
    if (idx >= 0) {
      list[idx] = { ...list[idx], pending: false, failed: true }
      this.persist()
    }
  }

  /** 未读：seq > read_seq 且非自己发送、已有服务端 msg_id */
  unreadCount(convId: string, selfId: number): number {
    const conv = this.state.convs[convId]
    const readSeq = conv?.read_seq ?? 0
    const list = this.state.messages[convId] ?? []
    let n = 0
    for (const m of list) {
      if (m.msg_id <= 0 || m.pending || m.failed) {
        continue
      }
      if (m.sender_id === selfId) {
        continue
      }
      if (m.seq > readSeq) {
        n += 1
      }
    }
    return n
  }

  totalUnread(selfId: number): number {
    let sum = 0
    for (const id of Object.keys(this.state.convs)) {
      sum += this.unreadCount(id, selfId)
    }
    return sum
  }

  /**
   * 推进已读水位：max(当前 read_seq, 本地消息最大 seq, 会话 last_seq)
   */
  markRead(convId: string): void {
    const conv = this.state.convs[convId]
    if (!conv) {
      return
    }
    let maxMsgSeq = 0
    for (const m of this.state.messages[convId] ?? []) {
      if (m.msg_id > 0 && m.seq > maxMsgSeq) {
        maxMsgSeq = m.seq
      }
    }
    const next = Math.max(conv.read_seq || 0, maxMsgSeq, conv.last_seq || 0)
    if (next === conv.read_seq) {
      return
    }
    this.state.convs[convId] = { ...conv, read_seq: next }
    this.persist()
  }

  private load(): ImLocalState {
    if (this.userId <= 0) {
      return emptyState()
    }
    const raw = localStorage.getItem(storageKey(this.userId))
    if (!raw) {
      return emptyState()
    }
    try {
      const parsed = JSON.parse(raw) as ImLocalState
      const convs: Record<string, LocalConv> = {}
      for (const [id, c] of Object.entries(parsed.convs || {})) {
        convs[id] = {
          ...c,
          history_cursor: c.history_cursor ?? 0,
          read_seq: c.read_seq ?? 0,
        }
      }
      return {
        list_cursor: parsed.list_cursor || 0,
        convs,
        messages: parsed.messages || {},
      }
    } catch {
      return emptyState()
    }
  }

  private persist(): void {
    if (this.userId <= 0) {
      return
    }
    localStorage.setItem(storageKey(this.userId), JSON.stringify(this.state))
  }
}
