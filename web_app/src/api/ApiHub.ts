import { HttpClient } from '@/api/http/HttpClient'
import { AssetApiService } from '@/api/services/AssetApiService'
import { AuthApiService } from '@/api/services/AuthApiService'
import { CartApiService } from '@/api/services/CartApiService'
import { MallApiService } from '@/api/services/MallApiService'
import { OrderApiService } from '@/api/services/OrderApiService'
import { ProductApiService } from '@/api/services/ProductApiService'
import { UserApiService } from '@/api/services/UserApiService'
import { AuthSession } from '@/auth/AuthSession'

/**
 * API 门面：集中持有 HttpClient 与各业务 Service
 */
export class ApiHub {
  readonly http: HttpClient
  readonly auth: AuthApiService
  readonly users: UserApiService
  readonly products: ProductApiService
  readonly cart: CartApiService
  readonly orders: OrderApiService
  readonly assets: AssetApiService
  readonly mall: MallApiService

  constructor(session: AuthSession) {
    this.http = new HttpClient({
      getAccessToken: () => session.getAccessToken(),
    })
    this.auth = new AuthApiService(this.http)
    this.users = new UserApiService(this.http)
    this.products = new ProductApiService(this.http)
    this.cart = new CartApiService(this.http)
    this.orders = new OrderApiService(this.http)
    this.assets = new AssetApiService(this.http)
    this.mall = new MallApiService(this.http)
  }
}
