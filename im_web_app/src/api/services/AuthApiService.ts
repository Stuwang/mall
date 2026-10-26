import type { HttpClient } from '@/api/http/HttpClient'
import type {
  GetUserInfoReply,
  GetUserInfoRequest,
  LoginReply,
  LoginRequest,
  MediaUploadReply,
  RegisterReply,
  RegisterRequest,
  UpdateUserInfoReply,
  UpdateUserInfoRequest,
} from '@/api/types'

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

  getUserInfo(request: GetUserInfoRequest = {}): Promise<GetUserInfoReply> {
    return this.http.post<GetUserInfoReply>('/user/api/v1/info', request)
  }

  updateUserInfo(request: UpdateUserInfoRequest): Promise<UpdateUserInfoReply> {
    return this.http.post<UpdateUserInfoReply>('/user/api/v1/update', request)
  }

  uploadMedia(file: File): Promise<MediaUploadReply> {
    const form = new FormData()
    form.append('file', file)
    return this.http.postForm<MediaUploadReply>('/user/api/v1/media/upload', form)
  }
}
