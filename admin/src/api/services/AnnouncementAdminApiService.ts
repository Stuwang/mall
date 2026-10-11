import type { HttpClient } from '@/api/http/HttpClient'
import type {
  AdminAnnouncementReply,
  AdminCreateAnnouncementRequest,
  AdminListAnnouncementsRequest,
  AdminUpdateAnnouncementRequest,
  ListAnnouncementsReply,
} from '@/api/types'

/** /admin/api/v1/announcement/* */
export class AnnouncementAdminApiService {
  private readonly http: HttpClient

  constructor(http: HttpClient) {
    this.http = http
  }

  list(request: AdminListAnnouncementsRequest = {}): Promise<ListAnnouncementsReply> {
    return this.http.post<ListAnnouncementsReply>('/admin/api/v1/announcement/list', request)
  }

  create(request: AdminCreateAnnouncementRequest): Promise<AdminAnnouncementReply> {
    return this.http.post<AdminAnnouncementReply>('/admin/api/v1/announcement/create', request)
  }

  update(request: AdminUpdateAnnouncementRequest): Promise<AdminAnnouncementReply> {
    return this.http.post<AdminAnnouncementReply>('/admin/api/v1/announcement/update', request)
  }
}
