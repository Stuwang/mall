import type { HttpClient } from '@/api/http/HttpClient'
import type {
  AddAlbumPhotoReply,
  AddAlbumPhotoRequest,
  DeleteAlbumPhotoReply,
  DeleteAlbumPhotoRequest,
  GetProfileReply,
  GetProfileRequest,
  ListAlbumReply,
  ListAlbumRequest,
} from '@/api/types'

export class ProfileApiService {
  private readonly http: HttpClient

  constructor(http: HttpClient) {
    this.http = http
  }

  getProfile(request: GetProfileRequest = {}): Promise<GetProfileReply> {
    return this.http.post<GetProfileReply>('/im/api/v1/profile/get', request)
  }

  listAlbum(request: ListAlbumRequest = {}): Promise<ListAlbumReply> {
    return this.http.post<ListAlbumReply>('/im/api/v1/album/list', request)
  }

  addAlbumPhoto(request: AddAlbumPhotoRequest): Promise<AddAlbumPhotoReply> {
    return this.http.post<AddAlbumPhotoReply>('/im/api/v1/album/add', request)
  }

  deleteAlbumPhoto(request: DeleteAlbumPhotoRequest): Promise<DeleteAlbumPhotoReply> {
    return this.http.post<DeleteAlbumPhotoReply>('/im/api/v1/album/delete', request)
  }
}
