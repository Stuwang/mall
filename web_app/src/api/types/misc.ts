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

export type ListBannersRequest = WithBaseParam

export interface ListBannersReply {
  list: BannerInfo[]
}

export type ListAnnouncementsRequest = WithBaseParam

export interface ListAnnouncementsReply {
  list: AnnouncementInfo[]
}
