<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  PayTypeAlipay,
  PayTypeMock,
  PayTypeWechat,
  RechargeStatusCancelled,
  RechargeStatusPaid,
  RechargeStatusPending,
  type RechargeOrder,
} from '@/api/types'
import { appContext } from '@/app/AppContext'
import { routeNames } from '@/router/routes'
import { formatYuanFromFen, parseRouteId } from '@/utils/coin'
import { notifyError, notifyFail } from '@/utils/notify'

const route = useRoute()
const router = useRouter()
const api = appContext.api

const orderId = computed(() => parseRouteId(route.params.orderId))
const order = ref<RechargeOrder | null>(null)
const loading = ref(true)
const paying = ref(false)
const payType = ref<number>(PayTypeMock)

const yuan = computed(() => formatYuanFromFen(order.value?.cash_amount ?? 0))

async function loadOrder(): Promise<void> {
  const id = orderId.value
  paying.value = false
  if (!id) {
    order.value = null
    loading.value = false
    notifyFail('订单不存在')
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

async function onConfirm(): Promise<void> {
  if (!order.value || paying.value || order.value.status !== RechargeStatusPending) {
    return
  }
  const selected = payType.value
  paying.value = true
  try {
    const reply = await api.coin.payRecharge({
      order_id: order.value.id,
      pay_type: selected,
    })
    order.value = reply
    if (selected === PayTypeMock) {
      await router.push({
        name: routeNames.coinMockPay,
        params: { orderId: String(reply.id || order.value.id) },
      })
      return
    }
  } catch (error) {
    notifyError(error, '支付失败')
  } finally {
    paying.value = false
  }
}

function goCoin(): void {
  void router.replace({ name: routeNames.coin })
}

watch(
  orderId,
  () => {
    payType.value = PayTypeMock
    void loadOrder()
  },
  { immediate: true },
)
</script>

<template>
  <div class="page pay-page">
    <van-nav-bar title="支付" left-arrow fixed placeholder @click-left="router.back()" />

    <div v-if="loading" class="state muted">加载中…</div>
    <template v-else-if="order">
      <section class="summary im-card im-fade-up">
        <div class="row">
          <span>订单号</span>
          <span class="value">{{ order.order_no }}</span>
        </div>
        <div class="row">
          <span>金币</span>
          <span class="value">{{ order.coins }}</span>
        </div>
        <div class="row">
          <span>金额</span>
          <span class="yuan">¥{{ yuan }}</span>
        </div>
      </section>

      <div v-if="order.status === RechargeStatusPaid" class="result im-fade-up">
        <div class="status ok">已支付</div>
        <van-button type="primary" block @click="goCoin">返回我的金币</van-button>
      </div>
      <div v-else-if="order.status === RechargeStatusCancelled" class="result im-fade-up">
        <div class="status">订单已关闭</div>
        <van-button block @click="goCoin">返回我的金币</van-button>
      </div>
      <div v-else-if="order.status === RechargeStatusPending" class="pending im-fade-up">
        <van-radio-group v-model="payType">
          <van-cell-group inset>
            <van-cell title="微信支付" clickable @click="payType = PayTypeWechat">
              <template #icon>
                <span class="pay-icon wechat"><van-icon name="wechat" /></span>
              </template>
              <template #right-icon>
                <van-radio :name="PayTypeWechat" />
              </template>
            </van-cell>
            <van-cell title="支付宝" clickable @click="payType = PayTypeAlipay">
              <template #icon>
                <span class="pay-icon alipay"><van-icon name="alipay" /></span>
              </template>
              <template #right-icon>
                <van-radio :name="PayTypeAlipay" />
              </template>
            </van-cell>
            <van-cell title="模拟支付" clickable @click="payType = PayTypeMock">
              <template #icon>
                <span class="pay-icon mock"><van-icon name="gold-coin-o" /></span>
              </template>
              <template #right-icon>
                <van-radio :name="PayTypeMock" />
              </template>
            </van-cell>
          </van-cell-group>
        </van-radio-group>
        <div class="confirm">
          <van-button type="primary" block :loading="paying" @click="onConfirm"
            >确认支付</van-button
          >
        </div>
      </div>
      <div v-else class="result">
        <div class="status">订单状态不允许</div>
        <van-button block @click="goCoin">返回我的金币</van-button>
      </div>
    </template>
    <div v-else class="result">
      <div class="status">订单加载失败</div>
      <van-button block @click="goCoin">返回我的金币</van-button>
    </div>
  </div>
</template>

<style scoped>
.pay-page {
  padding-bottom: calc(24px + var(--im-safe-bottom));
}

.summary {
  margin: 16px;
  padding: 8px 16px;
}

.row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 0;
  border-bottom: 1px solid var(--im-border);
  color: var(--im-text-secondary);
  font-size: 14px;
}

.row:last-child {
  border-bottom: none;
}

.value {
  color: var(--im-text);
  font-weight: 600;
  text-align: right;
  word-break: break-all;
}

.yuan {
  color: var(--im-accent);
  font-family: Outfit, 'Noto Sans SC', sans-serif;
  font-size: 22px;
  font-weight: 700;
}

.pending :deep(.van-cell-group) {
  margin-top: 4px;
}

.pay-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  margin-right: 10px;
  border-radius: 8px;
  font-size: 16px;
}

.pay-icon.wechat {
  color: #1aad19;
  background: rgba(26, 173, 25, 0.12);
}

.pay-icon.alipay {
  color: #1677ff;
  background: rgba(22, 119, 255, 0.12);
}

.pay-icon.mock {
  color: var(--im-accent);
  background: rgba(232, 135, 58, 0.14);
}

.confirm,
.result {
  margin: 20px 16px 0;
}

.status {
  margin-bottom: 16px;
  text-align: center;
  font-size: 18px;
  font-weight: 700;
}

.status.ok {
  color: var(--im-success);
}

.state {
  padding: 48px 16px;
  text-align: center;
}
</style>
