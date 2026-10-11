<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { appContext } from '@/app/AppContext'
import { OrderStatus, type OrderInfo } from '@/api/types'
import { formatFenToYuan, formatUnix } from '@/utils/format'
import { orderStatusLabel } from '@/utils/labels'
import { notifyError } from '@/utils/notify'

const router = useRouter()
const api = appContext.api.orders

const status = ref<OrderStatus>(OrderStatus.UNSPECIFIED)
const orders = ref<OrderInfo[]>([])
const page = ref(1)
const total = ref(0)
/** 仅绑定 van-list；不可用它做早退条件 */
const listLoading = ref(false)
const finished = ref(false)
let inFlight = false

async function fetchOrders(reset: boolean): Promise<void> {
  if (inFlight) {
    listLoading.value = false
    return
  }
  inFlight = true
  listLoading.value = true
  if (reset) {
    page.value = 1
    finished.value = false
    orders.value = []
  }
  try {
    const pageSize = 10
    const reply = await api.listOrders({
      status: status.value || undefined,
      page: page.value,
      page_size: pageSize,
    })
    const list = reply.list ?? []
    orders.value = reset ? list : orders.value.concat(list)
    total.value = Number(reply.total ?? 0)
    const noMore =
      list.length === 0 ||
      list.length < pageSize ||
      (total.value > 0 && orders.value.length >= total.value)
    finished.value = noMore
    if (!noMore) {
      page.value += 1
    }
  } catch (error) {
    finished.value = true
    notifyError(error, '订单加载失败')
  } finally {
    inFlight = false
    listLoading.value = false
  }
}

function onListLoad(): void {
  if (finished.value) {
    listLoading.value = false
    return
  }
  void fetchOrders(false)
}

function onTabChange(name: number | string): void {
  status.value = Number(name) as OrderStatus
  void fetchOrders(true)
}

onMounted(() => {
  void fetchOrders(true)
})
</script>

<template>
  <div class="page">
    <van-nav-bar title="我的订单" left-arrow fixed placeholder @click-left="router.back()" />

    <van-tabs :model-value="status" color="#ee0a24" sticky @change="onTabChange">
      <van-tab title="全部" :name="OrderStatus.UNSPECIFIED" />
      <van-tab title="待支付" :name="OrderStatus.PENDING_PAY" />
      <van-tab title="已支付" :name="OrderStatus.PAID" />
      <van-tab title="已取消" :name="OrderStatus.CANCELLED" />
    </van-tabs>

    <van-list
      v-model:loading="listLoading"
      :finished="finished"
      finished-text="没有更多了"
      :immediate-check="false"
      @load="onListLoad"
    >
      <div v-for="order in orders" :key="order.id" class="order-card">
        <div class="head">
          <span>{{ order.order_no }}</span>
          <span class="status">{{ orderStatusLabel(order.status) }}</span>
        </div>
        <div v-for="item in order.items" :key="item.id" class="line">
          <span class="name">{{ item.product_name }} × {{ item.quantity }}</span>
          <span class="price">{{ formatFenToYuan(item.amount) }}</span>
        </div>
        <div class="foot">
          <span class="muted">{{ formatUnix(order.ct) }}</span>
          <span>
            合计
            <span class="price">{{ formatFenToYuan(order.total_amount) }}</span>
          </span>
        </div>
      </div>
      <van-empty v-if="!listLoading && finished && orders.length === 0" description="暂无订单" />
    </van-list>
  </div>
</template>

<style scoped>
.order-card {
  margin: 10px 12px;
  padding: 12px;
  background: #fff;
  border-radius: 10px;
}

.head,
.foot,
.line {
  display: flex;
  justify-content: space-between;
  gap: 8px;
}

.head {
  margin-bottom: 8px;
  font-size: 13px;
}

.status {
  color: #ee0a24;
}

.line {
  margin: 6px 0;
  font-size: 14px;
}

.name {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.foot {
  margin-top: 10px;
  padding-top: 10px;
  border-top: 1px solid #f2f3f5;
  font-size: 13px;
}
</style>
