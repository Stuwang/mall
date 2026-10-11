import type { HttpClient } from '@/api/http/HttpClient'
import type {
  AdminBanUserRequest,
  AdminGetUserRequest,
  AdminListUsersReply,
  AdminListUsersRequest,
  GetUserInfoReply,
} from '@/api/types'

/** /admin/api/v1/user/* */
export class UserAdminApiService {
  private readonly http: HttpClient

  constructor(http: HttpClient) {
    this.http = http
  }

  list(request: AdminListUsersRequest): Promise<AdminListUsersReply> {
    return this.http.post<AdminListUsersReply>('/admin/api/v1/user/list', request)
  }

  detail(request: AdminGetUserRequest): Promise<GetUserInfoReply> {
    return this.http.post<GetUserInfoReply>('/admin/api/v1/user/detail', request)
  }

  ban(request: AdminBanUserRequest): Promise<GetUserInfoReply> {
    return this.http.post<GetUserInfoReply>('/admin/api/v1/user/ban', request)
  }
}
