import type { WithBaseParam } from './common'

/** asset.v1.AssetType */
export const AssetType = {
  UNSPECIFIED: 0,
  POINT: 1,
} as const
export type AssetType = (typeof AssetType)[keyof typeof AssetType]

export interface AssetInfo {
  user_id: number
  type: AssetType
  balance: number
  ct: number
  ut: number
}

export interface AdminGetUserAssetsRequest extends WithBaseParam {
  user_id: number
  type?: AssetType
}

export interface GetUserAssetsReply {
  assets: AssetInfo[]
}

export interface AdminGrantPointsRequest extends WithBaseParam {
  user_id: number
  /** 增加的积分数，必须 > 0 */
  points: number
  remark?: string
}

export interface AdminGrantPointsReply {
  asset: AssetInfo
}
