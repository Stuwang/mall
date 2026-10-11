import type { HttpClient } from '@/api/http/HttpClient'
import type {
  GetProductReply,
  GetProductRequest,
  ListCategoriesReply,
  ListCategoriesRequest,
  ListProductsByCategoryRequest,
  ListProductsReply,
  ListProductsRequest,
} from '@/api/types'

/** 用户侧商品浏览（白名单，可不登录） */
export class ProductApiService {
  private readonly http: HttpClient

  constructor(http: HttpClient) {
    this.http = http
  }

  listProducts(request: ListProductsRequest = {}): Promise<ListProductsReply> {
    return this.http.post<ListProductsReply>('/product/api/v1/list', request)
  }

  listProductsByCategory(request: ListProductsByCategoryRequest): Promise<ListProductsReply> {
    return this.http.post<ListProductsReply>('/product/api/v1/list_by_category', request)
  }

  getProduct(request: GetProductRequest): Promise<GetProductReply> {
    return this.http.post<GetProductReply>('/product/api/v1/detail', request)
  }

  listCategories(request: ListCategoriesRequest = {}): Promise<ListCategoriesReply> {
    return this.http.post<ListCategoriesReply>('/product/api/v1/category/list', request)
  }
}
