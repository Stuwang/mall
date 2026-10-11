import { OrderStatus, OrderType, PayType } from '@/api/types'

export function orderStatusLabel(status: OrderStatus): string {
  switch (status) {
    case OrderStatus.PENDING_PAY:
      return '待支付'
    case OrderStatus.PAID:
      return '已支付'
    case OrderStatus.CANCELLED:
      return '已取消'
    default:
      return '未知'
  }
}

export function orderTypeLabel(orderType: OrderType): string {
  switch (orderType) {
    case OrderType.PRODUCT:
      return '商品订单'
    case OrderType.POINT:
      return '积分充值'
    default:
      return '订单'
  }
}

export function payTypeLabel(payType: PayType): string {
  switch (payType) {
    case PayType.WECHAT:
      return '微信'
    case PayType.ALIPAY:
      return '支付宝'
    case PayType.POINT:
      return '积分'
    case PayType.MOCK:
      return 'Mock'
    default:
      return '-'
  }
}
