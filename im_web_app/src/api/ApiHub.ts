import { HttpClient } from '@/api/http/HttpClient'
import { AuthApiService } from '@/api/services/AuthApiService'
import { CoinApiService } from '@/api/services/CoinApiService'
import { ImApiService } from '@/api/services/ImApiService'
import { PostApiService } from '@/api/services/PostApiService'
import { ProfileApiService } from '@/api/services/ProfileApiService'
import type { AuthSession } from '@/auth/AuthSession'

export class ApiHub {
  readonly http: HttpClient
  readonly auth: AuthApiService
  readonly im: ImApiService
  readonly profile: ProfileApiService
  readonly post: PostApiService
  readonly coin: CoinApiService

  constructor(session: AuthSession) {
    this.http = new HttpClient({
      getAccessToken: () => session.getAccessToken(),
    })
    this.auth = new AuthApiService(this.http)
    this.im = new ImApiService(this.http)
    this.profile = new ProfileApiService(this.http)
    this.post = new PostApiService(this.http)
    this.coin = new CoinApiService(this.http)
  }
}
