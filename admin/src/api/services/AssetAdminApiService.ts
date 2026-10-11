import type { HttpClient } from '@/api/http/HttpClient'
import type {
  AdminGetAssetHistoryRequest,
  AdminGetUserAssetsRequest,
  AdminGrantPointsReply,
  AdminGrantPointsRequest,
  GetAssetHistoryReply,
  GetUserAssetsReply,
} from '@/api/types'

/** /admin/api/v1/asset/* */
export class AssetAdminApiService {
  private readonly http: HttpClient

  constructor(http: HttpClient) {
    this.http = http
  }

  getUserAssets(request: AdminGetUserAssetsRequest): Promise<GetUserAssetsReply> {
    return this.http.post<GetUserAssetsReply>('/admin/api/v1/asset/list', request)
  }

  getHistory(request: AdminGetAssetHistoryRequest): Promise<GetAssetHistoryReply> {
    return this.http.post<GetAssetHistoryReply>('/admin/api/v1/asset/history', request)
  }

  grantPoints(request: AdminGrantPointsRequest): Promise<AdminGrantPointsReply> {
    return this.http.post<AdminGrantPointsReply>('/admin/api/v1/asset/grant_points', request)
  }
}
