import { ApiCode } from '@/api/types'

/** 业务接口错误（HTTP 常为 200，以 code 判断） */
export class ApiError extends Error {
  readonly code: number
  readonly msg: string

  constructor(code: number, msg: string) {
    super(msg || `api error ${code}`)
    this.name = 'ApiError'
    this.code = code
    this.msg = msg
  }

  get isUnauthorized(): boolean {
    return this.code === ApiCode.Unauthorized
  }

  get isForbidden(): boolean {
    return this.code === ApiCode.Forbidden
  }

  get isBanned(): boolean {
    return this.code === ApiCode.UserBanned
  }
}
