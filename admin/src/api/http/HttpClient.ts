import axios, {
  type AxiosInstance,
  type AxiosRequestConfig,
  type InternalAxiosRequestConfig,
} from 'axios'
import { ApiError } from '@/api/errors/ApiError'
import { ApiCode, type ApiResponse } from '@/api/types'

export type TokenProvider = () => string | null
export type UnauthorizedHandler = (error: ApiError) => void

export interface HttpClientOptions {
  baseURL?: string
  timeoutMs?: number
  getAccessToken: TokenProvider
  onUnauthorized?: UnauthorizedHandler
}

/**
 * 统一 HTTP 客户端：挂 Bearer、解包 CommonResp、抛出 ApiError
 */
export class HttpClient {
  private readonly axios: AxiosInstance
  private readonly getAccessToken: TokenProvider
  private onUnauthorized: UnauthorizedHandler | undefined

  constructor(options: HttpClientOptions) {
    this.getAccessToken = options.getAccessToken
    this.onUnauthorized = options.onUnauthorized
    this.axios = axios.create({
      baseURL: options.baseURL ?? '',
      timeout: options.timeoutMs ?? 15000,
      headers: {
        'Content-Type': 'application/json',
      },
    })
    this.axios.interceptors.request.use((config) => this.attachAuth(config))
  }

  setUnauthorizedHandler(handler: UnauthorizedHandler | undefined): void {
    this.onUnauthorized = handler
  }

  async post<TResponse>(
    path: string,
    body?: unknown,
    config?: AxiosRequestConfig,
  ): Promise<TResponse> {
    const { data } = await this.axios.post<ApiResponse<TResponse>>(path, body ?? {}, config)
    return this.unwrap(data)
  }

  /** multipart 上传：去掉默认 JSON Content-Type，由浏览器自动带 boundary */
  async uploadForm<TResponse>(path: string, form: FormData): Promise<TResponse> {
    const { data } = await this.axios.post<ApiResponse<TResponse>>(path, form, {
      headers: { 'Content-Type': undefined },
      timeout: 60000,
    })
    return this.unwrap(data)
  }

  private attachAuth(config: InternalAxiosRequestConfig): InternalAxiosRequestConfig {
    const token = this.getAccessToken()
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  }

  private unwrap<T>(payload: ApiResponse<T>): T {
    if (!payload || typeof payload.code !== 'number') {
      throw new ApiError(ApiCode.System, 'invalid response')
    }
    if (payload.code !== ApiCode.Success) {
      const error = new ApiError(payload.code, payload.msg || 'request failed')
      if (error.isUnauthorized || error.isForbidden || error.isBanned) {
        this.onUnauthorized?.(error)
      }
      throw error
    }
    return payload.data
  }
}
