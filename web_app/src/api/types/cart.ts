import type { WithBaseParam } from './common'

/** 与 cart.v1.CartItem 对齐 */
export interface CartItem {
  id: number
  user_id: number
  product_id: number
  quantity: number
  product_name: string
  /** 价格，单位分 */
  product_price: number
  product_stock: number
  product_cover: string
  product_status: number
  ct: number
  ut: number
}

export interface AddToCartRequest extends WithBaseParam {
  product_id: number
  /** 增加数量，默认 1 */
  quantity?: number
}

export interface RemoveFromCartRequest extends WithBaseParam {
  /** 购物车项 ID；与 product_id 二选一，优先 id */
  id?: number
  product_id?: number
}

export type RemoveFromCartReply = Record<string, never>

export interface UpdateCartQuantityRequest extends WithBaseParam {
  id: number
  /** 目标数量；<=0 视为移除 */
  quantity: number
}

export interface CartItemReply {
  item: CartItem
}

export type ListCartRequest = WithBaseParam

export interface ListCartReply {
  list: CartItem[]
  total_quantity: number
}
