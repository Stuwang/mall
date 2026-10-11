import type { HttpClient } from '@/api/http/HttpClient'
import type {
  AdminCategoryReply,
  AdminCreateCategoryRequest,
  AdminListCategoriesRequest,
  AdminUpdateCategoryRequest,
  ListCategoriesReply,
} from '@/api/types'

/** /admin/api/v1/category/* */
export class CategoryAdminApiService {
  private readonly http: HttpClient

  constructor(http: HttpClient) {
    this.http = http
  }

  list(request: AdminListCategoriesRequest = {}): Promise<ListCategoriesReply> {
    return this.http.post<ListCategoriesReply>('/admin/api/v1/category/list', request)
  }

  create(request: AdminCreateCategoryRequest): Promise<AdminCategoryReply> {
    return this.http.post<AdminCategoryReply>('/admin/api/v1/category/create', request)
  }

  update(request: AdminUpdateCategoryRequest): Promise<AdminCategoryReply> {
    return this.http.post<AdminCategoryReply>('/admin/api/v1/category/update', request)
  }
}
