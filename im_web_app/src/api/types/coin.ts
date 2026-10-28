import type { WithBaseParam } from './common'
import type { Msg } from './im'

export const CoinChangeRecharge = 1
export const CoinChangeGiftSend = 2

export const PayTypeWechat = 1
export const PayTypeAlipay = 2
export const PayTypeMock = 3

export const RechargeStatusPending = 1
export const RechargeStatusPaid = 2
export const RechargeStatusCancelled = 3

export interface CoinBalanceReply {
  balance: number
  coins_per_yuan: number
}

export interface LedgerItem {
  id: number
  change_type: number
  change: number
  balance: number
  remark: string
  ext: string
  ct: number
}

export interface LedgerRequest extends WithBaseParam {
  before_id?: number
  limit?: number
}

export interface LedgerReply {
  list: LedgerItem[]
  has_more: boolean
}

export interface Gift {
  id: number
  name: string
  icon: string
  effect: string
  price: number
}

export interface GiftListReply {
  list: Gift[]
}

export interface SendGiftRequest extends WithBaseParam {
  peer_user_id: number
  gift_id: number
  client_msg_id: string
}

export interface SendGiftReply {
  balance: number
  msg: Msg
}

export interface RechargeOrder {
  id: number
  order_no: string
  coins: number
  cash_amount: number
  coins_per_yuan: number
  pay_type: number
  status: number
  expire_at: number
  paid_at: number
  ct: number
}

export interface CreateRechargeRequest extends WithBaseParam {
  coins: number
  client_order_id?: string
}

export interface PayRechargeRequest extends WithBaseParam {
  order_id: number
  pay_type: number
}

export interface RechargeOrderRequest extends WithBaseParam {
  order_id: number
}
