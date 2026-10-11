import type { WithBaseParam } from './common'

/** misc.v1.MiscStatus */
export const MiscStatus = {
  UNSPECIFIED: 0,
  ON: 1,
  OFF: 2,
} as const
export type MiscStatus = (typeof MiscStatus)[keyof typeof MiscStatus]

export interface BannerInfo {
  id: number
  image_url: string
  sort: number
  status: MiscStatus
  ct: number
  ut: number
}

export interface AnnouncementInfo {
  id: number
  content: string
  sort: number
  status: MiscStatus
  ct: number
  ut: number
}

export interface ListBannersReply {
  list: BannerInfo[]
}

export interface ListAnnouncementsReply {
  list: AnnouncementInfo[]
}

export interface AdminCreateBannerRequest extends WithBaseParam {
  image_url: string
  sort: number
  status: MiscStatus
}

export interface AdminUpdateBannerRequest extends WithBaseParam {
  id: number
  image_url: string
  sort: number
  status: MiscStatus
}

export type AdminListBannersRequest = WithBaseParam

export interface AdminBannerReply {
  banner: BannerInfo
}

export interface AdminCreateAnnouncementRequest extends WithBaseParam {
  content: string
  sort: number
  status: MiscStatus
}

export interface AdminUpdateAnnouncementRequest extends WithBaseParam {
  id: number
  content: string
  sort: number
  status: MiscStatus
}

export type AdminListAnnouncementsRequest = WithBaseParam

export interface AdminAnnouncementReply {
  announcement: AnnouncementInfo
}
