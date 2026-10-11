import type { HttpClient } from '@/api/http/HttpClient'
import type { GetMallInfoReply, GetMallInfoRequest } from '@/api/types'

/** 商城公共信息（免登录） */
export class MallApiService {
  private readonly http: HttpClient

  constructor(http: HttpClient) {
    this.http = http
  }

  getMallInfo(request: GetMallInfoRequest = {}): Promise<GetMallInfoReply> {
    return this.http.post<GetMallInfoReply>('/mall/api/v1/info', request)
  }
}
