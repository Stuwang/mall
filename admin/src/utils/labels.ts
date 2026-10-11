import {
  CancelReason,
  CategoryStatus,
  MiscStatus,
  OrderStatus,
  OrderType,
  PayType,
  ProductStatus,
} from '@/api/types'

export function productStatusLabel(status: ProductStatus): string {
  switch (status) {
    case ProductStatus.ON:
      return '上架'
    case ProductStatus.OFF:
      return '下架'
    default:
      return '未知'
  }
}

export function categoryStatusLabel(status: CategoryStatus): string {
  switch (status) {
    case CategoryStatus.ON:
      return '启用'
    case CategoryStatus.OFF:
      return '停用'
    default:
      return '未知'
  }
}

export function miscStatusLabel(status: MiscStatus): string {
  switch (status) {
    case MiscStatus.ON:
      return '启用'
    case MiscStatus.OFF:
      return '停用'
    default:
      return '未知'
  }
}

export function orderStatusLabel(status: OrderStatus): string {
  switch (status) {
    case OrderStatus.PENDING_PAY:
      return '待支付'
    case OrderStatus.PAID:
      return '已支付'
    case OrderStatus.CANCELLED:
      return '已取消'
    default:
      return '全部/未知'
  }
}

export function orderTypeLabel(type: OrderType): string {
  switch (type) {
    case OrderType.PRODUCT:
      return '商品单'
    case OrderType.POINT:
      return '积分单'
    default:
      return '未知'
  }
}

export function payTypeLabel(type: PayType): string {
  switch (type) {
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

export function cancelReasonLabel(reason: CancelReason): string {
  switch (reason) {
    case CancelReason.USER:
      return '用户取消'
    case CancelReason.TIMEOUT:
      return '超时取消'
    default:
      return '-'
  }
}

export function assetBizTypeLabel(bizType: string): string {
  switch (bizType) {
    case 'admin_grant_points':
      return '后台加积分'
    case 'order_pay_point':
      return '订单积分抵扣'
    case 'order_point_recharge':
      return '积分充值'
    default:
      return bizType || '-'
  }
}
