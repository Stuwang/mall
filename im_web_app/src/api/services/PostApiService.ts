import type { HttpClient } from '@/api/http/HttpClient'
import type {
  CreateCommentReply,
  CreateCommentRequest,
  CreatePostReply,
  CreatePostRequest,
  DeleteCommentReply,
  DeleteCommentRequest,
  DeletePostRequest,
  GetPostReply,
  GetPostRequest,
  LikePostReply,
  LikePostRequest,
  ListCommentReply,
  ListCommentRequest,
  ListUserPostReply,
  ListUserPostRequest,
  RecommendReply,
  RecommendRequest,
  UnlikePostReply,
  UnlikePostRequest,
} from '@/api/types'

export class PostApiService {
  private readonly http: HttpClient

  constructor(http: HttpClient) {
    this.http = http
  }

  recommend(request: RecommendRequest = {}): Promise<RecommendReply> {
    return this.http.post<RecommendReply>('/im/api/v1/post/recommend', request)
  }

  listUser(request: ListUserPostRequest = {}): Promise<ListUserPostReply> {
    return this.http.post<ListUserPostReply>('/im/api/v1/post/user', request)
  }

  create(request: CreatePostRequest): Promise<CreatePostReply> {
    return this.http.post<CreatePostReply>('/im/api/v1/post/create', request)
  }

  delete(request: DeletePostRequest): Promise<Record<string, never>> {
    return this.http.post('/im/api/v1/post/delete', request)
  }

  detail(request: GetPostRequest): Promise<GetPostReply> {
    return this.http.post<GetPostReply>('/im/api/v1/post/detail', request)
  }

  like(request: LikePostRequest): Promise<LikePostReply> {
    return this.http.post<LikePostReply>('/im/api/v1/post/like', request)
  }

  unlike(request: UnlikePostRequest): Promise<UnlikePostReply> {
    return this.http.post<UnlikePostReply>('/im/api/v1/post/unlike', request)
  }

  listComment(request: ListCommentRequest): Promise<ListCommentReply> {
    return this.http.post<ListCommentReply>('/im/api/v1/post/comment/list', request)
  }

  createComment(request: CreateCommentRequest): Promise<CreateCommentReply> {
    return this.http.post<CreateCommentReply>('/im/api/v1/post/comment/create', request)
  }

  deleteComment(request: DeleteCommentRequest): Promise<DeleteCommentReply> {
    return this.http.post<DeleteCommentReply>('/im/api/v1/post/comment/delete', request)
  }
}
