import type { WithBaseParam } from './common'

export interface ContactInfo {
  phone: string
  email: string
  wechat: string
  address: string
}

export type GetMallInfoRequest = WithBaseParam

export interface GetMallInfoReply {
  /** 1 元 = N 积分 */
  points_per_yuan: number
  contact: ContactInfo
  help_center: string
  about_us: string
}
