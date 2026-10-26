import { HttpClient } from '@/api/http/HttpClient'
import { AuthApiService } from '@/api/services/AuthApiService'
import { ImApiService } from '@/api/services/ImApiService'
import type { AuthSession } from '@/auth/AuthSession'

export class ApiHub {
  readonly http: HttpClient
  readonly auth: AuthApiService
  readonly im: ImApiService

  constructor(session: AuthSession) {
    this.http = new HttpClient({
      getAccessToken: () => session.getAccessToken(),
    })
    this.auth = new AuthApiService(this.http)
    this.im = new ImApiService(this.http)
  }
}
