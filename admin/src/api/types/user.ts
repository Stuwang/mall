import type { PageQuery, WithBaseParam } from './common'

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

export interface LoginRequest extends WithBaseParam {
  username: string
  password: string
}

export interface LoginReply {
  user: UserInfo
  access_token: string
}

export interface GetUserInfoRequest extends WithBaseParam {
  id?: number
}

export interface GetUserInfoReply {
  user: UserInfo
}

export interface AdminListUsersRequest extends WithBaseParam, PageQuery {
  /** 按用户 id 精确查；0=不限 */
  user_id?: number
  /** 匹配昵称或登录账号 */
  keyword?: string
}

export interface AdminListUsersReply {
  list: UserInfo[]
  total: number
}

export interface AdminGetUserRequest extends WithBaseParam {
  user_id: number
}

export interface AdminBanUserRequest extends WithBaseParam {
  user_id: number
  /** 封禁截止 unix 时间戳；传 0 解封 */
  ban_until: number
}
