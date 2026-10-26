import { appContext } from '@/app/AppContext'
import type { Conversation, Msg, OnlineUser } from '@/api/types'
import { ContentTypeImage, ContentTypeText } from '@/api/types'
import { makeConvId, newClientMsgId } from '@/utils/conv'
import { LocalStore } from './LocalStore'
import type { LocalConv, LocalMsg } from './types'
import { WsClient } from './WsClient'

type ChangeHandler = () => void

/**
 * IM 客户端：本地游标 + HTTP 兜底同步 + WS 推送先展示。
 * 推送/发送响应不推进游标，只由 list/history 成功响应推进。
 * 未读用本机 read_seq，不写服务端。
 */
export class ImClient {
  private readonly store = new LocalStore()
  private readonly ws = new WsClient()
  private userId = 0
  private syncing = false
  private syncTimer: number | null = null
  private listeners = new Set<ChangeHandler>()
  private online: OnlineUser[] = []
  private wsConnected = false
  /** 当前前台打开的对端；0 表示不在聊天页 */
  private activePeerId = 0

  constructor() {
    this.ws.setHandlers(
      (msg) => {
        void this.onPush(msg)
      },
      (ok) => {
        this.wsConnected = ok
        this.emit()
        if (ok) {
          void this.syncAll()
        }
      },
    )
  }

  subscribe(fn: ChangeHandler): () => void {
    this.listeners.add(fn)
    return () => this.listeners.delete(fn)
  }

  isWsConnected(): boolean {
    return this.wsConnected
  }

  getOnline(): OnlineUser[] {
    return this.online
  }

  listConvs(): LocalConv[] {
    return this.store.listConvs()
  }

  getMessages(peerUserId: number): LocalMsg[] {
    const convId = makeConvId(this.userId, peerUserId)
    return this.store.getMessages(convId)
  }

  getConv(peerUserId: number): LocalConv | undefined {
    return this.store.getConv(makeConvId(this.userId, peerUserId))
  }

  unreadOf(peerUserId: number): number {
    if (this.userId <= 0) {
      return 0
    }
    return this.store.unreadCount(makeConvId(this.userId, peerUserId), this.userId)
  }

  totalUnread(): number {
    if (this.userId <= 0) {
      return 0
    }
    return this.store.totalUnread(this.userId)
  }

  /** 进入 / 离开聊天页时调用。peerUserId=0 表示离开。 */
  setActivePeer(peerUserId: number): void {
    this.activePeerId = peerUserId > 0 ? peerUserId : 0
    if (this.activePeerId > 0) {
      this.markPeerRead(this.activePeerId)
    }
    this.emit()
  }

  start(userId: number, token: string): void {
    this.stop()
    this.userId = userId
    this.store.bindUser(userId)
    this.ws.connect(token)
    this.syncTimer = window.setInterval(() => void this.syncAll(false), 30000)
    void this.syncAll(true)
    void this.refreshOnline()
    this.emit()
  }

  stop(): void {
    this.ws.disconnect()
    if (this.syncTimer != null) {
      window.clearInterval(this.syncTimer)
      this.syncTimer = null
    }
    this.store.unbind()
    this.userId = 0
    this.activePeerId = 0
    this.online = []
    this.emit()
  }

  async refreshOnline(): Promise<void> {
    const reply = await appContext.api.im.listOnline()
    this.online = reply.list || []
    for (const u of this.online) {
      this.store.patchPeerProfile(u.user_id, u.nickname, u.avatar)
    }
    this.emit()
  }

  openLocalChat(peerUserId: number, nickname: string, avatar = ''): void {
    const convId = makeConvId(this.userId, peerUserId)
    this.store.ensureLocalConv({
      conv_id: convId,
      peer: { user_id: peerUserId, nickname, avatar },
      last_seq: 0,
      last_msg_id: 0,
      sort_ut: Math.floor(Date.now() / 1000),
      history_cursor: 0,
      read_seq: 0,
      local_only: true,
    })
    this.emit()
  }

  async syncAll(fullList = false): Promise<void> {
    if (this.userId <= 0 || this.syncing) {
      return
    }
    this.syncing = true
    try {
      await this.syncConvList(fullList)
      const convs = this.store.listConvs()
      for (const c of convs) {
        if (c.peer?.user_id) {
          await this.syncHistory(c.peer.user_id)
        }
      }
    } finally {
      this.syncing = false
      if (this.activePeerId > 0) {
        this.markPeerRead(this.activePeerId)
      }
      this.emit()
    }
  }

  async syncPeer(peerUserId: number): Promise<void> {
    await this.syncConvList(false)
    await this.syncHistory(peerUserId)
    if (this.activePeerId === peerUserId) {
      this.markPeerRead(peerUserId)
    }
    this.emit()
  }

  async sendText(
    peerUserId: number,
    text: string,
    peerNickname = '',
    peerAvatar = '',
  ): Promise<void> {
    const trimmed = text.trim()
    if (!trimmed) {
      return
    }
    await this.sendMessage(peerUserId, {
      contentType: ContentTypeText,
      text: trimmed,
      peerNickname,
      peerAvatar,
    })
  }

  async sendImage(
    peerUserId: number,
    file: File,
    peerNickname = '',
    peerAvatar = '',
  ): Promise<void> {
    const uploaded = await appContext.api.auth.uploadMedia(file)
    if (!uploaded?.uri) {
      throw new Error('上传未返回文件地址')
    }
    await this.sendMessage(peerUserId, {
      contentType: ContentTypeImage,
      imageUri: uploaded.uri,
      imageUrl: uploaded.url || '',
      peerNickname,
      peerAvatar,
    })
  }

  private async sendMessage(
    peerUserId: number,
    opts: {
      contentType: number
      text?: string
      imageUri?: string
      imageUrl?: string
      peerNickname?: string
      peerAvatar?: string
    },
  ): Promise<void> {
    const convId = makeConvId(this.userId, peerUserId)
    const clientMsgId = newClientMsgId()
    const now = Math.floor(Date.now() / 1000)
    this.store.ensureLocalConv({
      conv_id: convId,
      peer: {
        user_id: peerUserId,
        nickname: opts.peerNickname || `用户${peerUserId}`,
        avatar: opts.peerAvatar || '',
      },
      last_seq: 0,
      last_msg_id: 0,
      sort_ut: now,
      history_cursor: 0,
      read_seq: 0,
      local_only: true,
    })
    const pending: LocalMsg = {
      msg_id: 0,
      conv_id: convId,
      seq: 0,
      sender_id: this.userId,
      origin: 1,
      client_msg_id: clientMsgId,
      content_type: opts.contentType,
      content: {
        text: opts.text || '',
        image_url: opts.imageUrl || '',
      },
      ct: now,
      ut: now,
      pending: true,
    }
    this.store.upsertMessage(convId, pending)
    this.emit()

    try {
      const reply = await appContext.api.im.send({
        peer_user_id: peerUserId,
        client_msg_id: clientMsgId,
        content_type: opts.contentType,
        text: opts.text || '',
        image_uri: opts.imageUri || '',
      })
      this.applyServerMsg(reply.msg)
      this.applyServerConv(reply.conv)
      void this.syncPeer(peerUserId)
    } catch (err) {
      this.store.markFailed(convId, clientMsgId)
      this.emit()
      throw err
    }
  }

  private markPeerRead(peerUserId: number): void {
    const convId = makeConvId(this.userId, peerUserId)
    this.store.markRead(convId)
  }

  private async onPush(msg: Msg): Promise<void> {
    this.applyServerMsg(msg)
    const peerId = this.peerFromMsg(msg)
    if (peerId > 0 && this.activePeerId === peerId) {
      this.markPeerRead(peerId)
    }
    this.emit()
    if (peerId > 0) {
      await this.syncPeer(peerId)
    } else {
      await this.syncConvList(false)
      this.emit()
    }
  }

  private peerFromMsg(msg: Msg): number {
    if (!msg.conv_id) {
      return 0
    }
    const [a, b] = msg.conv_id.split('_').map((x) => Number(x))
    if (!a || !b) {
      return 0
    }
    return a === this.userId ? b : a
  }

  private applyServerMsg(msg: Msg): void {
    if (!msg?.conv_id) {
      return
    }
    this.store.upsertMessage(msg.conv_id, { ...msg, pending: false, failed: false })
  }

  private applyServerConv(conv: Conversation): void {
    if (!conv?.conv_id) {
      return
    }
    const prev = this.store.getConv(conv.conv_id)
    const next: LocalConv = {
      ...conv,
      history_cursor: prev?.history_cursor ?? 0,
      read_seq: prev?.read_seq ?? 0,
      local_only: false,
    }
    this.store.upsertConv(next)
  }

  private async syncConvList(full = false): Promise<void> {
    let cursor = full ? 0 : this.store.getListCursor()
    for (;;) {
      const reply = await appContext.api.im.listConv({ after_seq: cursor, limit: 100 })
      const list = reply.list || []
      for (const c of list) {
        this.applyServerConv(c)
      }
      if (!list.length) {
        if (full) {
          this.store.setListCursor(cursor)
        }
        break
      }
      const maxSeq = Math.max(...list.map((c) => c.last_seq))
      if (reply.has_more) {
        cursor = list[list.length - 1].last_seq
        this.store.setListCursor(cursor)
        continue
      }
      this.store.setListCursor(maxSeq)
      break
    }
  }

  private async syncHistory(peerUserId: number): Promise<void> {
    const convId = makeConvId(this.userId, peerUserId)
    let cursor = this.store.getConv(convId)?.history_cursor ?? 0
    for (;;) {
      const reply = await appContext.api.im.history({
        peer_user_id: peerUserId,
        after_seq: cursor,
        limit: 50,
      })
      const list = reply.list || []
      for (const m of list) {
        this.applyServerMsg(m)
      }
      const existing = this.store.getConv(convId)
      if (!existing && list.length === 0 && reply.last_seq === 0) {
        return
      }
      if (existing) {
        this.store.upsertConv({
          ...existing,
          last_seq: Math.max(existing.last_seq, reply.last_seq),
          read_seq: existing.read_seq,
        })
      }
      if (list.length >= 50 || reply.has_more) {
        cursor = list[list.length - 1].seq
        const cur = this.store.getConv(convId) as LocalConv
        this.store.upsertConv({
          ...cur,
          history_cursor: cursor,
          read_seq: cur.read_seq,
        })
        continue
      }
      const finalConv = this.store.getConv(convId)
      if (finalConv) {
        this.store.upsertConv({
          ...finalConv,
          history_cursor: reply.last_seq,
          last_seq: Math.max(finalConv.last_seq, reply.last_seq),
          read_seq: finalConv.read_seq,
        })
      }
      break
    }
  }

  private emit(): void {
    for (const fn of this.listeners) {
      fn()
    }
  }
}

export const imClient = new ImClient()
