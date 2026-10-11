<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { OrderStatus, OrderType, type OrderInfo } from '@/api/types'
import { appContext } from '@/app/AppContext'
import { routeNames } from '@/router/routes'
import { formatFenToYuan, formatUnix } from '@/utils/format'
import { orderStatusLabel, orderTypeLabel, payTypeLabel } from '@/utils/labels'
import { notifyError } from '@/utils/notify'

const route = useRoute()
const router = useRouter()
const orderApi = appContext.api.orders

const order = ref<OrderInfo | null>(null)
const loading = ref(true)

const orderId = computed(() => Number(route.params.id))

const canPay = computed(() => order.value?.status === OrderStatus.PENDING_PAY)
const isPointOrder = computed(() => order.value?.order_type === OrderType.POINT)

async function load(): Promise<void> {
  loading.value = true
  try {
    const reply = await orderApi.getOrder({ order_id: orderId.value })
    order.value = reply.order
  } catch (error) {
    notifyError(error, '加载订单失败')
    order.value = null
  } finally {
    loading.value = false
  }
}

function goPay(): void {
  if (!order.value) return
  void router.push({
    name: routeNames.payCheckout,
    params: { id: String(order.value.id) },
  })
}

onMounted(() => {
  void load()
})
</script>

<template>
  <div class="page">
    <van-nav-bar title="订单详情" left-arrow fixed placeholder @click-left="router.back()" />

    <van-skeleton title :row="10" :loading="loading">
      <template v-if="order">
        <van-cell-group inset class="block">
          <van-cell title="订单状态" :value="orderStatusLabel(order.status)" />
          <van-cell title="订单类型" :value="orderTypeLabel(order.order_type)" />
          <van-cell title="订单号" :value="order.order_no" />
          <van-cell title="下单时间" :value="formatUnix(order.ct)" />
          <van-cell v-if="order.paid_at" title="支付时间" :value="formatUnix(order.paid_at)" />
        </van-cell-group>

        <van-cell-group v-if="isPointOrder" inset class="block" title="充值明细">
          <van-cell title="购买积分" :value="`${order.point_amount} 积分`" />
        </van-cell-group>

        <van-cell-group
          v-else-if="order.items?.length"
          inset
          class="block"
          title="商品明细"
        >
          <van-cell
            v-for="item in order.items"
            :key="item.id"
            :title="item.product_name"
            :label="`× ${item.quantity}`"
          >
            <template #value>
              <span class="price">{{ formatFenToYuan(item.amount) }}</span>
            </template>
          </van-cell>
        </van-cell-group>

        <van-cell-group inset class="block" title="支付详情">
          <van-cell title="支付方式" :value="payTypeLabel(order.pay_type)" />
          <van-cell title="订单总额">
            <template #value>
              <span class="price">{{ formatFenToYuan(order.total_amount) }}</span>
            </template>
          </van-cell>
          <van-cell
            :title="isPointOrder ? '到账积分' : '积分抵扣'"
            :value="`${order.point_amount} 积分`"
          />
          <van-cell title="现金金额">
            <template #value>
              <span class="price">{{ formatFenToYuan(order.cash_amount) }}</span>
            </template>
          </van-cell>
          <van-cell
            title="积分汇率"
            :value="order.points_per_yuan ? `1元 = ${order.points_per_yuan} 积分` : '-'"
          />
        </van-cell-group>

        <div v-if="canPay" class="actions">
          <van-button round block type="danger" @click="goPay">去支付</van-button>
        </div>
      </template>
      <van-empty v-else description="订单不存在" />
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
