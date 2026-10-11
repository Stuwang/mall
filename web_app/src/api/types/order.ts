import type { PageQuery, WithBaseParam } from './common'

/** order.v1.OrderType */
export const OrderType = {
  UNSPECIFIED: 0,
  PRODUCT: 1,
  POINT: 2,
} as const
export type OrderType = (typeof OrderType)[keyof typeof OrderType]

/** order.v1.OrderStatus */
export const OrderStatus = {
  UNSPECIFIED: 0,
  PENDING_PAY: 1,
  PAID: 2,
  CANCELLED: 3,
} as const
export type OrderStatus = (typeof OrderStatus)[keyof typeof OrderStatus]

/** order.v1.PayType */
export const PayType = {
  UNSPECIFIED: 0,
  WECHAT: 1,
  ALIPAY: 2,
  POINT: 3,
  MOCK: 4,
} as const
export type PayType = (typeof PayType)[keyof typeof PayType]

/** order.v1.CancelReason */
export const CancelReason = {
  UNSPECIFIED: 0,
  USER: 1,
  TIMEOUT: 2,
} as const
export type CancelReason = (typeof CancelReason)[keyof typeof CancelReason]

export interface OrderItemInfo {
  id: number
  product_id: number
  product_name: string
  product_price: number
  quantity: number
  amount: number
}

export interface OrderInfo {
  id: number
  order_no: string
  user_id: number
  order_type: OrderType
  status: OrderStatus
  pay_type: PayType
  /** 订单总价（分） */
  total_amount: number
  /** 现金应付/已付（分） */
  cash_amount: number
  /** 商品单：抵扣积分数；积分单：购买积分数 */
  point_amount: number
  /** 下单时积分汇率快照：1 元 = N 积分 */
  points_per_yuan: number
  pay_expire_at: number
  paid_at: number
  cancel_reason: CancelReason
  items: OrderItemInfo[]
  ct: number
  ut: number
}

export interface CreateProductItem {
  product_id: number
  quantity: number
}

export interface CreateProductOrderRequest extends WithBaseParam {
  items: CreateProductItem[]
}

export interface CreateCartOrderRequest extends WithBaseParam {
  cart_item_ids: number[]
}

export interface CreatePointOrderRequest extends WithBaseParam {
  points: number
}

export interface PayOrderRequest extends WithBaseParam {
  order_id: number
  pay_type: PayType
  use_points?: number
}

export interface MockPaySuccessRequest extends WithBaseParam {
  order_id: number
}

export interface CancelOrderRequest extends WithBaseParam {
  order_id: number
}

export interface GetOrderRequest extends WithBaseParam {
  order_id: number
}

export interface ListOrdersRequest extends WithBaseParam, PageQuery {
  /** 0=全部 */
  status?: OrderStatus
}

export interface OrderReply {
  order: OrderInfo
}

export interface ListOrdersReply {
  list: OrderInfo[]
  total: number
}
