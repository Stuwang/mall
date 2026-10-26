export interface BaseParam {
  app?: string
  ver?: string
  ip?: string
}

export interface WithBaseParam {
  BaseParam?: BaseParam
}

export interface ApiResponse<T> {
  code: number
  msg: string
  msg_trans?: Record<string, string> | null
  data: T
}
