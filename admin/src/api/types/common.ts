/** 与后端 helloworld.v1.BaseParam 对齐 */
export interface BaseParam {
  app?: string
  ver?: string
  ip?: string
}

/** 可携带 BaseParam 的请求基类字段 */
export interface WithBaseParam {
  BaseParam?: BaseParam
}

/**
 * 后端统一 HTTP 响应包装（pkg/server/encoder.go CommonResp）
 * data 按接口不同而变化
 */
export interface ApiResponse<T> {
  code: number
  msg: string
  msg_trans?: Record<string, string> | null
  data: T
}

export interface PageQuery {
  page?: number
  page_size?: number
}
