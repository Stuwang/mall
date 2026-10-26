import type { UserInfo } from '@/api/types'

const STORAGE_KEY = 'im.web.auth'

interface PersistedAuth {
  accessToken: string
  user: UserInfo
}

export class AuthSession {
  private accessToken: string | null = null
  private user: UserInfo | null = null

  constructor() {
    this.restore()
  }

  getAccessToken(): string | null {
    return this.accessToken
  }

  getUser(): UserInfo | null {
    return this.user
  }

  isLoggedIn(): boolean {
    return Boolean(this.accessToken)
  }

  applyLogin(accessToken: string, user: UserInfo): void {
    this.accessToken = accessToken
    this.user = user
    this.persist()
  }

  updateUser(user: UserInfo): void {
    this.user = user
    this.persist()
  }

  clear(): void {
    this.accessToken = null
    this.user = null
    localStorage.removeItem(STORAGE_KEY)
  }

  private persist(): void {
    if (!this.accessToken || !this.user) {
      return
    }
    const payload: PersistedAuth = {
      accessToken: this.accessToken,
      user: this.user,
    }
    localStorage.setItem(STORAGE_KEY, JSON.stringify(payload))
  }

  private restore(): void {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) {
      return
    }
    try {
      const parsed = JSON.parse(raw) as PersistedAuth
      if (parsed.accessToken && parsed.user) {
        this.accessToken = parsed.accessToken
        this.user = parsed.user
      }
    } catch {
      localStorage.removeItem(STORAGE_KEY)
    }
  }
}
