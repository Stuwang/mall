<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  PayTypeMock,
  RechargeStatusCancelled,
  RechargeStatusPaid,
  RechargeStatusPending,
  type RechargeOrder,
} from '@/api/types'
import { appContext } from '@/app/AppContext'
import { routeNames } from '@/router/routes'
import { formatYuanFromFen, parseRouteId } from '@/utils/coin'
import { touchCoin } from '@/utils/feedSync'
import { notifyError, notifySuccess } from '@/utils/notify'

const route = useRoute()
const router = useRouter()
const api = appContext.api

const orderId = computed(() => parseRouteId(route.params.orderId))
const order = ref<RechargeOrder | null>(null)
const loading = ref(true)
const paying = ref(false)

const yuan = computed(() => formatYuanFromFen(order.value?.cash_amount ?? 0))
const canPay = computed(
  () => order.value?.status === RechargeStatusPending && order.value.pay_type === PayTypeMock,
)

const blockedText = computed(() => {
  if (!order.value) {
    return '订单加载失败'
  }
  if (order.value.status === RechargeStatusPaid) {
    return '订单已支付'
  }
  if (order.value.status === RechargeStatusCancelled) {
    return '订单已关闭'
  }
  if (order.value.status !== RechargeStatusPending) {
    return '订单状态不允许支付'
  }
  if (order.value.pay_type !== PayTypeMock) {
    return '当前订单不是模拟支付，请返回支付页重新选择'
  }
  return '暂时无法支付'
})

async function loadOrder(): Promise<void> {
  const id = orderId.value
  paying.value = false
  if (!id) {
    order.value = null
    loading.value = false
    return
  }
  loading.value = true
  try {
    order.value = await api.coin.getRecharge({ order_id: id })
  } catch (error) {
    order.value = null
    notifyError(error, '加载订单失败')
  } finally {
    loading.value = false
  }
}

async function onPaySuccess(): Promise<void> {
  if (!order.value || paying.value || !canPay.value) {
    return
  }
  paying.value = true
  try {
    await api.coin.mockPaySuccess({ order_id: order.value.id })
    touchCoin()
    notifySuccess('支付成功')
    await router.replace({ name: routeNames.coin })
  } catch (error) {
    notifyError(error, '支付失败')
  } finally {
    paying.value = false
  }
}

function goPay(): void {
  const id = order.value?.id || orderId.value
  if (!id) {
    void router.replace({ name: routeNames.coin })
    return
  }
  void router.replace({
    name: routeNames.coinPay,
    params: { orderId: String(id) },
  })
}

function goCoin(): void {
  void router.replace({ name: routeNames.coin })
}

watch(
  orderId,
  () => {
    void loadOrder()
  },
  { immediate: true },
)
</script>

<template>
  <div class="page cashier">
    <van-nav-bar title="收银台" left-arrow fixed placeholder @click-left="router.back()" />

    <div v-if="loading" class="state muted">加载中…</div>
    <template v-else-if="order && canPay">
      <section class="ticket im-card im-fade-up">
        <div class="brand">模拟收银台</div>
        <div class="merchant">金币充值</div>
        <div class="yuan">¥{{ yuan }}</div>
        <div class="coins">{{ order.coins }} 金币</div>
        <div class="split" />
        <div class="meta">
          <span>订单号</span>
          <span>{{ order.order_no }}</span>
        </div>
        <div class="meta">
          <span>支付方式</span>
          <span>模拟支付</span>
        </div>
      </section>
      <div class="action im-fade-up">
        <van-button type="primary" block :loading="paying" @click="onPaySuccess">
          支付成功
        </van-button>
        <p class="hint">确认后将完成本次模拟支付，金币会记入余额。</p>
      </div>
    </template>
    <div v-else class="blocked im-fade-up">
      <div class="status">{{ blockedText }}</div>
      <van-button v-if="order" type="primary" block @click="goPay">返回支付页</van-button>
      <van-button block class="second" @click="goCoin">返回我的金币</van-button>
    </div>
  </div>
</template>

<style scoped>
.cashier {
  min-height: 100vh;
  min-height: 100dvh;
  padding-bottom: calc(24px + var(--im-safe-bottom));
  background:
    radial-gradient(520px 220px at 50% -40px, rgba(15, 143, 134, 0.16), transparent 70%),
    transparent;
}

.ticket {
  margin: 28px 20px 0;
  padding: 28px 20px 18px;
  text-align: center;
}

.brand {
  font-size: 13px;
  font-weight: 700;
  letter-spacing: 0.12em;
  color: var(--im-primary-deep);
}

.merchant {
  margin-top: 14px;
  font-size: 16px;
  color: var(--im-text-secondary);
}

.yuan {
  margin-top: 8px;
  font-family: Outfit, 'Noto Sans SC', sans-serif;
  font-size: 48px;
  font-weight: 700;
  letter-spacing: -0.03em;
  color: var(--im-text);
}

.coins {
  margin-top: 4px;
  color: var(--im-accent);
  font-weight: 700;
}

.split {
  margin: 18px 0 8px;
  border-top: 1px dashed var(--im-border);
}

.meta {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  padding: 8px 0;
  color: var(--im-muted);
  font-size: 13px;
  text-align: left;
}

.meta span:last-child {
  color: var(--im-text);
  text-align: right;
  word-break: break-all;
}

.action {
  margin: 22px 20px 0;
}

.hint {
  margin: 12px 8px 0;
  color: var(--im-muted);
  font-size: 12px;
  line-height: 1.5;
  text-align: center;
}

.blocked {
  margin: 48px 20px 0;
}

.status {
  margin-bottom: 18px;
  text-align: center;
  font-size: 16px;
  font-weight: 700;
  line-height: 1.5;
}

.second {
  margin-top: 10px;
}

.state {
  padding: 48px 16px;
  text-align: center;
}
</style>
