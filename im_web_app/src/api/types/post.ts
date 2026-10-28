import type { WithBaseParam } from './common'

export interface PostItem {
  post_id: number
  user_id: number
  nickname: string
  avatar?: string
  text: string
  like_count: number
  comment_count: number
  liked: boolean
  ct: number
  recommend_score?: number
}

export interface CommentItem {
  comment_id: number
  post_id: number
  user_id: number
  nickname: string
  avatar?: string
  text: string
  ct: number
}

export interface RecommendRequest extends WithBaseParam {
  has_cursor?: boolean
  cursor_score?: number
  cursor_id?: number
  limit?: number
}

export interface RecommendReply {
  list: PostItem[]
  has_more: boolean
}

export interface CreatePostRequest extends WithBaseParam {
  client_post_id: string
  text: string
}

export interface CreatePostReply {
  post: PostItem
}

export interface DeletePostRequest extends WithBaseParam {
  post_id: number
}

export interface GetPostRequest extends WithBaseParam {
  post_id: number
}

export interface GetPostReply {
  post: PostItem
}

export interface LikePostRequest extends WithBaseParam {
  post_id: number
}

export interface LikePostReply {
  post: PostItem
}

export interface UnlikePostRequest extends WithBaseParam {
  post_id: number
}

export interface UnlikePostReply {
  post: PostItem
}

export interface ListCommentRequest extends WithBaseParam {
  post_id: number
  after_id?: number
  limit?: number
}

export interface ListCommentReply {
  list: CommentItem[]
  has_more: boolean
}

export interface CreateCommentRequest extends WithBaseParam {
  post_id: number
  client_comment_id: string
  text: string
}

export interface CreateCommentReply {
  comment: CommentItem
  post: PostItem
}

export interface DeleteCommentRequest extends WithBaseParam {
  comment_id: number
}

export interface DeleteCommentReply {
  post: PostItem
}
