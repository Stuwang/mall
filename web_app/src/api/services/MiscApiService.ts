import type { HttpClient } from '@/api/http/HttpClient'
import type {
  ListAnnouncementsReply,
  ListAnnouncementsRequest,
  ListBannersReply,
  ListBannersRequest,
} from '@/api/types'

/** 用户侧 Banner / 公告（白名单，可不登录） */
export class MiscApiService {
  private readonly http: HttpClient

  constructor(http: HttpClient) {
    this.http = http
  }

  listBanners(request: ListBannersRequest = {}): Promise<ListBannersReply> {
    return this.http.post<ListBannersReply>('/misc/api/v1/banner/list', request)
  }

  listAnnouncements(request: ListAnnouncementsRequest = {}): Promise<ListAnnouncementsReply> {
    return this.http.post<ListAnnouncementsReply>('/misc/api/v1/announcement/list', request)
  }
}
