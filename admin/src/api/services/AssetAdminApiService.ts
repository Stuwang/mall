import type { HttpClient } from '@/api/http/HttpClient'
import type {
  AdminGetUserAssetsRequest,
  AdminGrantPointsReply,
  AdminGrantPointsRequest,
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

  grantPoints(request: AdminGrantPointsRequest): Promise<AdminGrantPointsReply> {
    return this.http.post<AdminGrantPointsReply>('/admin/api/v1/asset/grant_points', request)
  }
}
