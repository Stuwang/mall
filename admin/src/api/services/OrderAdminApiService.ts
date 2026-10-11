import type { HttpClient } from '@/api/http/HttpClient'
import type {
  AdminGetOrderRequest,
  AdminListOrdersRequest,
  ListOrdersReply,
  OrderReply,
} from '@/api/types'

/** /admin/api/v1/order/* */
export class OrderAdminApiService {
  private readonly http: HttpClient

  constructor(http: HttpClient) {
    this.http = http
  }

  list(request: AdminListOrdersRequest): Promise<ListOrdersReply> {
    return this.http.post<ListOrdersReply>('/admin/api/v1/order/list', request)
  }

  detail(request: AdminGetOrderRequest): Promise<OrderReply> {
    return this.http.post<OrderReply>('/admin/api/v1/order/detail', request)
  }
}
