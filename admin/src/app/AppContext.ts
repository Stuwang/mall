import { ApiHub } from '@/api/ApiHub'
import { AuthSession } from '@/auth/AuthSession'

/**
 * 应用级单例容器：Session + ApiHub
 */
export class AppContext {
  readonly session: AuthSession
  readonly api: ApiHub

  constructor() {
    this.session = new AuthSession()
    this.api = new ApiHub(this.session)
  }
}

export const appContext = new AppContext()
