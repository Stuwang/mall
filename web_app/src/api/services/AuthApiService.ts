import type { HttpClient } from '@/api/http/HttpClient'
import type {
  GetUserInfoReply,
  GetUserInfoRequest,
  LoginReply,
  LoginRequest,
  RefreshTokenReply,
  RegisterReply,
  RegisterRequest,
} from '@/api/types'

/** 注册 / 登录 / 当前用户信息 */
export class AuthApiService {
  private readonly http: HttpClient

  constructor(http: HttpClient) {
    this.http = http
  }

  register(request: RegisterRequest): Promise<RegisterReply> {
    return this.http.post<RegisterReply>('/user/api/v1/register', request)
  }

  login(request: LoginRequest): Promise<LoginReply> {
    return this.http.post<LoginReply>('/user/api/v1/login', request)
  }

  refreshToken(): Promise<RefreshTokenReply> {
    return this.http.post<RefreshTokenReply>('/user/api/v1/token/refresh', {})
  }

  getUserInfo(request: GetUserInfoRequest = {}): Promise<GetUserInfoReply> {
    return this.http.post<GetUserInfoReply>('/user/api/v1/info', request)
  }
}
