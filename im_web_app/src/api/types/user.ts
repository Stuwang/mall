import type { WithBaseParam } from './common'

export interface UserInfo {
  id: number
  username: string
  nickname: string
  sex: number
  ct: number
  ut: number
  is_admin: boolean
  ban_until: number
  banned: boolean
  avatar?: string
}

export interface UpdateUserInfoRequest extends WithBaseParam {
  nickname?: string
  sex?: number
  avatar?: string
}

export interface UpdateUserInfoReply {
  user: UserInfo
}

export interface MediaUploadReply {
  uri: string
  url: string
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
