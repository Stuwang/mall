import type { HttpClient } from '@/api/http/HttpClient'
import type {
  AdminCreateProductRequest,
  AdminGetProductRequest,
  AdminListProductsRequest,
  AdminProductReply,
  AdminSetProductStatusRequest,
  AdminUpdateProductCategoryRequest,
  AdminUpdateProductRequest,
  AdminUpdateProductStockRequest,
  GetProductReply,
  ListProductsReply,
} from '@/api/types'

/** /admin/api/v1/product/* */
export class ProductAdminApiService {
  private readonly http: HttpClient

  constructor(http: HttpClient) {
    this.http = http
  }

  list(request: AdminListProductsRequest): Promise<ListProductsReply> {
    return this.http.post<ListProductsReply>('/admin/api/v1/product/list', request)
  }

  detail(request: AdminGetProductRequest): Promise<GetProductReply> {
    return this.http.post<GetProductReply>('/admin/api/v1/product/detail', request)
  }

  create(request: AdminCreateProductRequest): Promise<AdminProductReply> {
    return this.http.post<AdminProductReply>('/admin/api/v1/product/create', request)
  }

  update(request: AdminUpdateProductRequest): Promise<AdminProductReply> {
    return this.http.post<AdminProductReply>('/admin/api/v1/product/update', request)
  }

  setStatus(request: AdminSetProductStatusRequest): Promise<AdminProductReply> {
    return this.http.post<AdminProductReply>('/admin/api/v1/product/set_status', request)
  }

  updateStock(request: AdminUpdateProductStockRequest): Promise<AdminProductReply> {
    return this.http.post<AdminProductReply>('/admin/api/v1/product/update_stock', request)
  }

  updateCategory(request: AdminUpdateProductCategoryRequest): Promise<AdminProductReply> {
    return this.http.post<AdminProductReply>('/admin/api/v1/product/update_category', request)
  }
}
