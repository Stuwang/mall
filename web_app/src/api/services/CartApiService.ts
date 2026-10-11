import type { HttpClient } from '@/api/http/HttpClient'
import type {
  AddToCartRequest,
  CartItemReply,
  ListCartReply,
  ListCartRequest,
  RemoveFromCartReply,
  RemoveFromCartRequest,
  UpdateCartQuantityRequest,
} from '@/api/types'

/** 用户购物车 */
export class CartApiService {
  private readonly http: HttpClient

  constructor(http: HttpClient) {
    this.http = http
  }

  add(request: AddToCartRequest): Promise<CartItemReply> {
    return this.http.post<CartItemReply>('/cart/api/v1/add', request)
  }

  remove(request: RemoveFromCartRequest): Promise<RemoveFromCartReply> {
    return this.http.post<RemoveFromCartReply>('/cart/api/v1/remove', request)
  }

  updateQuantity(request: UpdateCartQuantityRequest): Promise<CartItemReply> {
    return this.http.post<CartItemReply>('/cart/api/v1/update_quantity', request)
  }

  list(request: ListCartRequest = {}): Promise<ListCartReply> {
    return this.http.post<ListCartReply>('/cart/api/v1/list', request)
  }
}
