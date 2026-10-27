import type { WithBaseParam } from './common'

/** 与 user.v1.UserInfo 对齐 */
export interface UserInfo {
  id: number
  username: string
  nickname: string
  sex: number
  ct: number
  ut: number
  is_admin: boolean
  /** 封禁截止时间戳；0=未封禁 */
  ban_until: number
  /** 当前是否处于封禁中 */
  banned: boolean
}

export interface RegisterRequest extends WithBaseParam {
  username: string
  password: string
  nickname?: string
}

export interface RegisterReply {
  user: UserInfo
}

export interface LoginRequest extends WithBaseParam {
  username: string
  password: string
}

export interface LoginReply {
  user: UserInfo
  access_token: string
}

export interface RefreshTokenReply {
  access_token: string
}

export interface GetUserInfoRequest extends WithBaseParam {
  id?: number
}

export interface GetUserInfoReply {
  user: UserInfo
}

export interface UpdateUserInfoRequest extends WithBaseParam {
  id?: number
  nickname?: string
  sex?: number
}

export interface UpdateUserInfoReply {
  user: UserInfo
}
