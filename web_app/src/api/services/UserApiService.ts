import type { HttpClient } from '@/api/http/HttpClient'
import type { UpdateUserInfoReply, UpdateUserInfoRequest } from '@/api/types'

/** 用户资料修改 */
export class UserApiService {
  private readonly http: HttpClient

  constructor(http: HttpClient) {
    this.http = http
  }

  updateUserInfo(request: UpdateUserInfoRequest): Promise<UpdateUserInfoReply> {
    return this.http.post<UpdateUserInfoReply>('/user/api/v1/update', request)
  }
}
