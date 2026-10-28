import type { WithBaseParam } from './common'

export const ContentTypeText = 1
export const ContentTypeImage = 2
export const ContentTypeGift = 3
export const OriginUser = 1
export const OriginSystem = 2

export interface Peer {
  user_id: number
  nickname: string
  avatar?: string
}

export interface MsgGift {
  gift_id: number
  name: string
  icon: string
  effect: string
  price: number
}

export interface MsgContent {
  text?: string
  image_url?: string
  gift?: MsgGift
}

export interface Msg {
  msg_id: number
  conv_id: string
  seq: number
  sender_id: number
  origin: number
  client_msg_id: string
  content_type: number
  content: MsgContent
  ct: number
  ut: number
}

export interface MsgPreview {
  msg_id: number
  sender_id: number
  origin: number
  content_type: number
  preview: string
  ct: number
  ut: number
}

export interface Conversation {
  conv_id: string
  peer: Peer
  last_seq: number
  last_msg_id: number
  sort_ut: number
  last_msg?: MsgPreview | null
}

export interface ListConvRequest extends WithBaseParam {
  after_seq?: number
  limit?: number
}

export interface ListConvReply {
  list: Conversation[]
  has_more: boolean
}

export interface HistoryRequest extends WithBaseParam {
  peer_user_id: number
  after_seq?: number
  limit?: number
}

export interface HistoryReply {
  list: Msg[]
  last_seq: number
  has_more: boolean
}

export interface SendRequest extends WithBaseParam {
  peer_user_id: number
  client_msg_id: string
  content_type: number
  text?: string
  image_uri?: string
}

export interface SendReply {
  msg: Msg
  conv: Conversation
}

export interface OnlineUser {
  user_id: number
  nickname: string
  conn_count: number
  avatar?: string
  /** 相册预览图，从新到旧，最多 4 张 */
  album_urls?: string[]
}

export interface ListOnlineReply {
  list: OnlineUser[]
}
