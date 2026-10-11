import type { HttpClient } from '@/api/http/HttpClient'
import type {
  AdminBannerReply,
  AdminCreateBannerRequest,
  AdminListBannersRequest,
  AdminUpdateBannerRequest,
  ListBannersReply,
} from '@/api/types'

/** /admin/api/v1/banner/* */
export class BannerAdminApiService {
  private readonly http: HttpClient

  constructor(http: HttpClient) {
    this.http = http
  }

  list(request: AdminListBannersRequest = {}): Promise<ListBannersReply> {
    return this.http.post<ListBannersReply>('/admin/api/v1/banner/list', request)
  }

  create(request: AdminCreateBannerRequest): Promise<AdminBannerReply> {
    return this.http.post<AdminBannerReply>('/admin/api/v1/banner/create', request)
  }

  update(request: AdminUpdateBannerRequest): Promise<AdminBannerReply> {
    return this.http.post<AdminBannerReply>('/admin/api/v1/banner/update', request)
  }
}
