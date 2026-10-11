import type { HttpClient } from '@/api/http/HttpClient'
import type {
  GetAssetHistoryReply,
  GetAssetHistoryRequest,
  GetUserAssetsReply,
  GetUserAssetsRequest,
} from '@/api/types'

/** 用户资产（积分等） */
export class AssetApiService {
  private readonly http: HttpClient

  constructor(http: HttpClient) {
    this.http = http
  }

  getUserAssets(request: GetUserAssetsRequest = {}): Promise<GetUserAssetsReply> {
    return this.http.post<GetUserAssetsReply>('/asset/api/v1/list', request)
  }

  getAssetHistory(request: GetAssetHistoryRequest = {}): Promise<GetAssetHistoryReply> {
    return this.http.post<GetAssetHistoryReply>('/asset/api/v1/history', request)
  }
}
