import { HttpClient } from '@/api/http/HttpClient'
import { AssetAdminApiService } from '@/api/services/AssetAdminApiService'
import { AuthApiService } from '@/api/services/AuthApiService'
import { CategoryAdminApiService } from '@/api/services/CategoryAdminApiService'
import { OrderAdminApiService } from '@/api/services/OrderAdminApiService'
import { ProductAdminApiService } from '@/api/services/ProductAdminApiService'
import { UserAdminApiService } from '@/api/services/UserAdminApiService'
import { AuthSession } from '@/auth/AuthSession'

/**
 * API 门面：集中持有 HttpClient 与各业务 Service
 */
export class ApiHub {
  readonly http: HttpClient
  readonly auth: AuthApiService
  readonly users: UserAdminApiService
  readonly assets: AssetAdminApiService
  readonly categories: CategoryAdminApiService
  readonly products: ProductAdminApiService
  readonly orders: OrderAdminApiService

  constructor(session: AuthSession) {
    this.http = new HttpClient({
      getAccessToken: () => session.getAccessToken(),
    })
    this.auth = new AuthApiService(this.http)
    this.users = new UserAdminApiService(this.http)
    this.assets = new AssetAdminApiService(this.http)
    this.categories = new CategoryAdminApiService(this.http)
    this.products = new ProductAdminApiService(this.http)
    this.orders = new OrderAdminApiService(this.http)
  }
}
