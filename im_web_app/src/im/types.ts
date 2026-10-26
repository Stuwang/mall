import type { Conversation, Msg } from '@/api/types'

export interface LocalMsg extends Msg {
  /** 本地发送中、尚未拿到服务端 msg_id */
  pending?: boolean
  failed?: boolean
}

export interface LocalConv extends Conversation {
  /** 本场历史游标（已确认的前缀） */
  history_cursor: number
  /** 仅本地草稿会话（还没服务端消息） */
  local_only?: boolean
}

export interface ImLocalState {
  list_cursor: number
  convs: Record<string, LocalConv>
  messages: Record<string, LocalMsg[]>
}
