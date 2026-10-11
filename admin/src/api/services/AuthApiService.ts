import type { HttpClient } from '@/api/http/HttpClient'
import type { GetUserInfoReply, GetUserInfoRequest, LoginReply, LoginRequest } from '@/api/types'

/** 用户登录 / 当前用户信息（复用 /user 接口） */
export class AuthApiService {
  private readonly http: HttpClient

  constructor(http: HttpClient) {
    this.http = http
  }

  login(request: LoginRequest): Promise<LoginReply> {
    return this.http.post<LoginReply>('/user/api/v1/login', request)
  }

  getUserInfo(request: GetUserInfoRequest = {}): Promise<GetUserInfoReply> {
    return this.http.post<GetUserInfoReply>('/user/api/v1/info', request)
  }
}
