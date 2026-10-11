import type { PageQuery, WithBaseParam } from './common'

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

export interface AssetHistory {
  id: number
  user_id: number
  type: AssetType
  /** 变更量（可正可负） */
  change: number
  /** 变更后余额 */
  balance: number
  idempotent_id: string
  biz_type: string
  biz_id: string
  remark: string
  ct: number
}

export interface GetUserAssetsRequest extends WithBaseParam {
  /** 可选：按类型筛选，0 表示全部 */
  type?: AssetType
}

export interface GetUserAssetsReply {
  assets: AssetInfo[]
}

export interface GetAssetHistoryRequest extends WithBaseParam, PageQuery {
  type?: AssetType
}

export interface GetAssetHistoryReply {
  list: AssetHistory[]
  total: number
}
