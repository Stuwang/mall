<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { showConfirmDialog } from 'vant'
import { useRoute, useRouter } from 'vue-router'
import { AssetType, OrderStatus, OrderType, PayType, type OrderInfo } from '@/api/types'
import { appContext } from '@/app/AppContext'
import { routeNames } from '@/router/routes'
import { formatFenToYuan } from '@/utils/format'
import { payTypeLabel } from '@/utils/labels'
import { notifyError, notifySuccess } from '@/utils/notify'
import { maxUsablePoints, pointsToCents } from '@/utils/points'

const route = useRoute()
const router = useRouter()
const orderApi = appContext.api.orders
const assetApi = appContext.api.assets

const order = ref<OrderInfo | null>(null)
const pointBalance = ref(0)
const usePointsEnabled = ref(false)
const usePoints = ref(0)
const payType = ref<PayType>(PayType.MOCK)
const loading = ref(true)
const submitting = ref(false)

const orderId = computed(() => Number(route.params.id))

const maxPoints = computed(() => {
  if (!order.value) return 0
  return maxUsablePoints(
    order.value.total_amount,
    order.value.points_per_yuan,
    pointBalance.value,
  )
})

const deductCents = computed(() => {
  if (!order.value || !usePointsEnabled.value) return 0
  return pointsToCents(usePoints.value, order.value.points_per_yuan)
})

const isPointOrder = computed(() => order.value?.order_type === OrderType.POINT)

const cashCents = computed(() => {
  if (!order.value) return 0
  if (isPointOrder.value) return order.value.total_amount
  return Math.max(0, order.value.total_amount - deductCents.value)
})

async function load(): Promise<void> {
  loading.value = true
  try {
    const orderReply = await orderApi.getOrder({ order_id: orderId.value })
    order.value = orderReply.order

    if (order.value.status !== OrderStatus.PENDING_PAY) {
      await router.replace({
        name: routeNames.orderDetail,
        params: { id: String(order.value.id) },
      })
      return
    }

    // 积分充值单不能再用积分抵扣
    if (order.value.order_type !== OrderType.POINT) {
      const assetReply = await assetApi.getUserAssets({ type: AssetType.POINT })
      const point = (assetReply.assets ?? []).find((item) => item.type === AssetType.POINT)
      pointBalance.value = point?.balance ?? 0
    } else {
      usePointsEnabled.value = false
      usePoints.value = 0
      pointBalance.value = 0
    }
  } catch (error) {
    notifyError(error, '加载订单失败')
    order.value = null
  } finally {
    loading.value = false
  }
}

async function onTogglePoints(checked: boolean): Promise<void> {
  if (!checked) {
    usePointsEnabled.value = false
    usePoints.value = 0
    return
  }
  if (maxPoints.value <= 0) {
    notifyError(new Error('暂无可用积分'))
    usePointsEnabled.value = false
    return
  }
  try {
    await showConfirmDialog({
      title: '使用积分抵扣',
      message: `当前可用 ${maxPoints.value} 积分（余额 ${pointBalance.value}）。确认开启积分抵扣？`,
    })
    usePointsEnabled.value = true
    usePoints.value = maxPoints.value
  } catch {
    usePointsEnabled.value = false
    usePoints.value = 0
  }
}

async function onSubmit(): Promise<void> {
  if (!order.value) return
  const points = isPointOrder.value ? 0 : usePointsEnabled.value ? usePoints.value : 0
  if (points > 0) {
    try {
      await showConfirmDialog({
        title: '确认支付方案',
        message: `使用 ${points} 积分抵扣 ¥${formatFenToYuan(deductCents.value)}，现金应付 ¥${formatFenToYuan(cashCents.value)}。确认继续？`,
      })
    } catch {
      return
    }
  }

  submitting.value = true
  try {
    const reply = await orderApi.pay({
      order_id: order.value.id,
      pay_type: payType.value,
      use_points: points,
    })
    const paid = reply.order
    if (paid.status === OrderStatus.PAID) {
      notifySuccess('支付成功')
      await router.replace({
        name: routeNames.orderDetail,
        params: { id: String(paid.id) },
      })
      return
    }
    if (paid.pay_type === PayType.MOCK) {
      await router.replace({
        name: routeNames.mockPay,
        params: { id: String(paid.id) },
      })
      return
    }
    notifyError(new Error(`${payTypeLabel(paid.pay_type)} 暂未开通，请选择 Mock 支付`))
  } catch (error) {
    notifyError(error, '发起支付失败')
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  void load()
})
</script>

<template>
  <div class="page">
    <van-nav-bar title="确认支付" left-arrow fixed placeholder @click-left="router.back()" />

    <van-skeleton title :row="8" :loading="loading">
      <template v-if="order">
        <van-cell-group inset class="block">
          <van-cell title="订单号" :value="order.order_no" />
          <van-cell
            v-if="isPointOrder"
            title="购买积分"
            :value="`${order.point_amount} 积分`"
          />
          <van-cell title="订单总额">
            <template #value>
              <span class="price">{{ formatFenToYuan(order.total_amount) }}</span>
            </template>
          </van-cell>
          <van-cell title="汇率快照" :value="`1元 = ${order.points_per_yuan} 积分`" />
        </van-cell-group>

        <van-cell-group v-if="!isPointOrder" inset class="block">
          <van-cell title="积分余额" :value="String(pointBalance)" />
          <van-cell title="积分抵扣" center>
            <template #right-icon>
              <van-switch
                :model-value="usePointsEnabled"
                size="20px"
                @update:model-value="onTogglePoints"
              />
            </template>
          </van-cell>
          <van-cell v-if="usePointsEnabled" title="使用积分">
            <template #value>
              <van-stepper
                v-model="usePoints"
                integer
                :min="0"
                :max="maxPoints"
                theme="round"
                button-size="22px"
              />
            </template>
          </van-cell>
          <van-cell title="积分抵扣金额">
            <template #value>
              <span class="price">{{ formatFenToYuan(deductCents) }}</span>
            </template>
          </van-cell>
          <van-cell title="现金应付">
            <template #value>
              <span class="price">{{ formatFenToYuan(cashCents) }}</span>
            </template>
          </van-cell>
        </van-cell-group>

        <van-cell-group v-else inset class="block">
          <van-cell title="现金应付">
            <template #value>
              <span class="price">{{ formatFenToYuan(cashCents) }}</span>
            </template>
          </van-cell>
          <van-cell title="说明" value="积分充值不可再用积分支付" />
        </van-cell-group>

        <van-cell-group inset class="block" title="支付方式">
          <van-radio-group v-model="payType">
            <van-cell title="Mock 支付（测试）" clickable @click="payType = PayType.MOCK">
              <template #right-icon>
                <van-radio :name="PayType.MOCK" />
              </template>
            </van-cell>
            <van-cell title="微信支付" label="暂未开通" clickable>
              <template #right-icon>
                <van-radio :name="PayType.WECHAT" disabled />
              </template>
            </van-cell>
            <van-cell title="支付宝" label="暂未开通" clickable>
              <template #right-icon>
                <van-radio :name="PayType.ALIPAY" disabled />
              </template>
            </van-cell>
          </van-radio-group>
        </van-cell-group>

        <div class="actions">
          <van-button
            round
            block
            type="danger"
            :loading="submitting"
            :disabled="cashCents > 0 && payType !== PayType.MOCK"
            @click="onSubmit"
          >
            {{ cashCents === 0 ? '确认积分支付' : '去支付' }}
          </van-button>
        </div>
      </template>
      <van-empty v-else description="订单不可用" />
    </van-skeleton>
  </div>
</template>

<style scoped>
.block {
  margin-top: 12px;
}

.actions {
  margin: 24px 16px;
}
</style>
