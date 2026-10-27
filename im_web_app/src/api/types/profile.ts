import type { WithBaseParam } from './common'

export interface AlbumPhoto {
  id: number
  image_url: string
  caption: string
  ct: number
}

export interface GetProfileRequest extends WithBaseParam {
  user_id?: number
}

export interface GetProfileReply {
  user_id: number
  nickname: string
  avatar?: string
  photo_count: number
}

export interface ListAlbumRequest extends WithBaseParam {
  user_id?: number
  before_id?: number
  limit?: number
}

export interface ListAlbumReply {
  list: AlbumPhoto[]
  has_more: boolean
}

export interface AddAlbumPhotoRequest extends WithBaseParam {
  image_uri: string
  caption?: string
}

export interface AddAlbumPhotoReply {
  photo: AlbumPhoto
}

export interface DeleteAlbumPhotoRequest extends WithBaseParam {
  photo_id: number
}

export interface DeleteAlbumPhotoReply {}
