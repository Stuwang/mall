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
    this.state.convs[conv.conv_id] = {
      ...prev,
      ...conv,
      history_cursor: conv.history_cursor ?? prev?.history_cursor ?? 0,
    }
    this.persist()
  }

  ensureLocalConv(conv: LocalConv): void {
    if (this.state.convs[conv.conv_id]) {
      return
    }
    this.state.convs[conv.conv_id] = { ...conv, history_cursor: 0, local_only: true }
    this.persist()
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
      return {
        list_cursor: parsed.list_cursor || 0,
        convs: parsed.convs || {},
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
