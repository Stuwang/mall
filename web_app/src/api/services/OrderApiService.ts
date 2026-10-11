import type { HttpClient } from '@/api/http/HttpClient'
import type {
  CancelOrderRequest,
  CreateCartOrderRequest,
  CreatePointOrderRequest,
  CreateProductOrderRequest,
  GetOrderRequest,
  ListOrdersReply,
  ListOrdersRequest,
  MockPaySuccessRequest,
  OrderReply,
  PayOrderRequest,
} from '@/api/types'

/** 用户订单 */
export class OrderApiService {
  private readonly http: HttpClient

  constructor(http: HttpClient) {
    this.http = http
  }

  createProductOrder(request: CreateProductOrderRequest): Promise<OrderReply> {
    return this.http.post<OrderReply>('/order/api/v1/create_product', request)
  }

  createCartOrder(request: CreateCartOrderRequest): Promise<OrderReply> {
    return this.http.post<OrderReply>('/order/api/v1/create_cart', request)
  }

  createPointOrder(request: CreatePointOrderRequest): Promise<OrderReply> {
    return this.http.post<OrderReply>('/order/api/v1/create_point', request)
  }

  pay(request: PayOrderRequest): Promise<OrderReply> {
    return this.http.post<OrderReply>('/order/api/v1/pay', request)
  }

  mockPaySuccess(request: MockPaySuccessRequest): Promise<OrderReply> {
    return this.http.post<OrderReply>('/order/api/v1/mock_pay_success', request)
  }

  cancel(request: CancelOrderRequest): Promise<OrderReply> {
    return this.http.post<OrderReply>('/order/api/v1/cancel', request)
  }

  getOrder(request: GetOrderRequest): Promise<OrderReply> {
    return this.http.post<OrderReply>('/order/api/v1/detail', request)
  }

  listOrders(request: ListOrdersRequest = {}): Promise<ListOrdersReply> {
    return this.http.post<ListOrdersReply>('/order/api/v1/list', request)
  }
}
