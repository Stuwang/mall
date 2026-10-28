export const ApiCode = {
  Success: 0,
  System: 1,
  Param: 1000,
  Unauthorized: 1005,
  UserExists: 1006,
  UserNotFound: 1007,
  Password: 1008,
  Forbidden: 1009,
  UserBanned: 1014,
  Idempotency: 1108,
  CoinNotEnough: 1140,
  PayChannelClosed: 1141,
  RechargeOrderNotFound: 1142,
  RechargeStatusInvalid: 1143,
  RechargeExpired: 1144,
  GiftNotFound: 1145,
} as const

export type ApiCode = (typeof ApiCode)[keyof typeof ApiCode]
