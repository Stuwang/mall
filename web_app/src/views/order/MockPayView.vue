<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { showConfirmDialog } from 'vant'
import { useRoute, useRouter } from 'vue-router'
import { OrderStatus, PayType, type OrderInfo } from '@/api/types'
import { appContext } from '@/app/AppContext'
import { routeNames } from '@/router/routes'
import { formatFenToYuan } from '@/utils/format'
import { notifyError, notifySuccess } from '@/utils/notify'

const route = useRoute()
const router = useRouter()
const orderApi = appContext.api.orders

const order = ref<OrderInfo | null>(null)
const loading = ref(true)
const paying = ref(false)

const orderId = computed(() => Number(route.params.id))

async function load(): Promise<void> {
  loading.value = true
  try {
    const reply = await orderApi.getOrder({ order_id: orderId.value })
    order.value = reply.order
    if (reply.order.status === OrderStatus.PAID) {
      await router.replace({
        name: routeNames.orderDetail,
        params: { id: String(reply.order.id) },
      })
      return
    }
    if (
      reply.order.status !== OrderStatus.PENDING_PAY ||
      reply.order.pay_type !== PayType.MOCK
    ) {
      await router.replace({
        name: routeNames.payCheckout,
        params: { id: String(reply.order.id) },
      })
    }
  } catch (error) {
    notifyError(error, '加载订单失败')
    order.value = null
  } finally {
    loading.value = false
  }
}

async function onConfirmPay(): Promise<void> {
  if (!order.value) return
  try {
    await showConfirmDialog({
      title: 'Mock 支付确认',
      message: `确认支付现金 ¥${formatFenToYuan(order.value.cash_amount)}？${
        order.value.point_amount > 0
          ? `\n同时将抵扣 ${order.value.point_amount} 积分。`
          : ''
      }`,
    })
  } catch {
    return
  }

  paying.value = true
  try {
    const reply = await orderApi.mockPaySuccess({ order_id: order.value.id })
    notifySuccess('支付成功')
    await router.replace({
      name: routeNames.orderDetail,
      params: { id: String(reply.order.id) },
    })
  } catch (error) {
    notifyError(error, '支付失败')
  } finally {
    paying.value = false
  }
}

onMounted(() => {
  void load()
})
</script>

<template>
  <div class="page mock-page">
    <van-nav-bar title="Mock 支付" left-arrow fixed placeholder @click-left="router.back()" />

    <van-skeleton title :row="6" :loading="loading">
      <template v-if="order">
        <div class="hero">
          <div class="label">应付现金</div>
          <div class="amount">
            <span class="yen">¥</span>{{ formatFenToYuan(order.cash_amount) }}
          </div>
          <p class="muted">模拟第三方收银台，确认后视为支付成功</p>
        </div>

        <van-cell-group inset>
          <van-cell title="订单号" :value="order.order_no" />
          <van-cell title="订单总额">
            <template #value>
              <span class="price">{{ formatFenToYuan(order.total_amount) }}</span>
            </template>
          </van-cell>
          <van-cell title="积分抵扣" :value="`${order.point_amount} 积分`" />
          <van-cell title="支付方式" value="Mock" />
        </van-cell-group>

        <div class="actions">
          <van-button round block type="primary" :loading="paying" @click="onConfirmPay">
            确认支付
          </van-button>
          <van-button
            round
            block
            plain
            type="primary"
            @click="router.replace({ name: routeNames.payCheckout, params: { id: String(order.id) } })"
          >
            返回修改支付方式
          </van-button>
        </div>
      </template>
      <van-empty v-else description="订单不可用" />
    </van-skeleton>
  </div>
</template>

<style scoped>
.hero {
  margin: 20px 16px;
  padding: 28px 16px;
  text-align: center;
  background: linear-gradient(160deg, #1f6feb, #1158c7);
  border-radius: 14px;
  color: #fff;
}

.label {
  font-size: 14px;
  opacity: 0.9;
}

.amount {
  margin: 8px 0;
  font-size: 40px;
  font-weight: 700;
  letter-spacing: 0.02em;
}

.yen {
  margin-right: 4px;
  font-size: 22px;
}

.hero .muted {
  margin: 0;
  color: rgba(255, 255, 255, 0.85);
}

.actions {
  display: grid;
  gap: 12px;
  margin: 24px 16px;
}
</style>
