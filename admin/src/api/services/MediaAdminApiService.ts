import type { HttpClient } from '@/api/http/HttpClient'

export interface MediaUploadReply {
  uri: string
  url: string
}

/**
 * 管理端媒体上传
 */
export class MediaAdminApiService {
  private readonly http: HttpClient

  constructor(http: HttpClient) {
    this.http = http
  }

  upload(file: File): Promise<MediaUploadReply> {
    const form = new FormData()
    form.append('file', file)
    return this.http.uploadForm<MediaUploadReply>('/admin/api/v1/media/upload', form)
  }
}
