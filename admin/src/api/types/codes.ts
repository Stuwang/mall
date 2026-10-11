/** 与后端 base/codeerror 对齐的业务错误码 */
export const ApiCode = {
  Success: 0,
  System: 1,
  Param: 1000,
  Unauthorized: 1005,
  UserExists: 1006,
  UserNotFound: 1007,
  Password: 1008,
  Forbidden: 1009,
  StockInsufficient: 1010,
  AssetInsufficient: 1011,
  OrderStatus: 1012,
  PayExpired: 1013,
  UserBanned: 1014,
} as const

export type ApiCode = (typeof ApiCode)[keyof typeof ApiCode]
