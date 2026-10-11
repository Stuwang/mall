<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { OrderStatus, type OrderInfo } from '@/api/types'
import { appContext } from '@/app/AppContext'
import { formatFenToYuan, formatUnix } from '@/utils/format'
import {
  cancelReasonLabel,
  orderStatusLabel,
  orderTypeLabel,
  payTypeLabel,
} from '@/utils/labels'
import { showError } from '@/utils/notify'

const api = appContext.api.orders
const loading = ref(false)
const detailLoading = ref(false)
const list = ref<OrderInfo[]>([])
const total = ref(0)
const detailVisible = ref(false)
const current = ref<OrderInfo | null>(null)

const query = reactive({
  user_id: undefined as number | undefined,
  product_id: undefined as number | undefined,
  status: 0 as OrderStatus,
  page: 1,
  page_size: 20,
})

async function loadList(): Promise<void> {
  loading.value = true
  try {
    const reply = await api.list({
      user_id: query.user_id || 0,
      product_id: query.product_id || 0,
      status: query.status || OrderStatus.UNSPECIFIED,
      page: query.page,
      page_size: query.page_size,
    })
    list.value = reply.list ?? []
    total.value = Number(reply.total ?? 0)
  } catch (error) {
    showError(error, '加载订单失败')
  } finally {
    loading.value = false
  }
}

function handleSearch(): void {
  query.page = 1
  void loadList()
}

async function openDetail(row: OrderInfo): Promise<void> {
  detailVisible.value = true
  detailLoading.value = true
  current.value = null
  try {
    const reply = await api.detail({ order_id: row.id })
    current.value = reply.order
  } catch (error) {
    showError(error, '加载订单详情失败')
    detailVisible.value = false
  } finally {
    detailLoading.value = false
  }
}

onMounted(() => {
  void loadList()
})
</script>

<template>
  <div class="page-card">
    <div class="page-toolbar">
      <el-input-number
        v-model="query.user_id"
        :min="0"
        :controls="false"
        placeholder="用户 ID"
      />
      <el-input-number
        v-model="query.product_id"
        :min="0"
        :controls="false"
        placeholder="商品 ID"
      />
      <el-select v-model="query.status" style="width: 140px">
        <el-option :value="0" label="全部状态" />
        <el-option :value="OrderStatus.PENDING_PAY" label="待支付" />
        <el-option :value="OrderStatus.PAID" label="已支付" />
        <el-option :value="OrderStatus.CANCELLED" label="已取消" />
      </el-select>
      <el-button type="primary" @click="handleSearch">查询</el-button>
    </div>

    <el-table v-loading="loading" :data="list" stripe border>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="order_no" label="订单号" min-width="180" />
      <el-table-column prop="user_id" label="用户" width="90" />
      <el-table-column label="类型" width="90">
        <template #default="{ row }">{{ orderTypeLabel(row.order_type) }}</template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag size="small">{{ orderStatusLabel(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="总价(元)" width="100">
        <template #default="{ row }">{{ formatFenToYuan(row.total_amount) }}</template>
      </el-table-column>
      <el-table-column label="现金(元)" width="100">
        <template #default="{ row }">{{ formatFenToYuan(row.cash_amount) }}</template>
      </el-table-column>
      <el-table-column prop="point_amount" label="积分" width="90" />
      <el-table-column label="下单时间" min-width="160">
        <template #default="{ row }">{{ formatUnix(row.ct) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="90" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openDetail(row)">详情</el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pager">
      <el-pagination
        v-model:current-page="query.page"
        v-model:page-size="query.page_size"
        background
        layout="total, prev, pager, next, sizes"
        :total="total"
        :page-sizes="[10, 20, 50]"
        @current-change="loadList"
        @size-change="
          () => {
            query.page = 1
            loadList()
          }
        "
      />
    </div>

    <el-drawer v-model="detailVisible" title="订单详情" size="480px">
      <div v-loading="detailLoading">
        <template v-if="current">
          <el-descriptions :column="1" border>
            <el-descriptions-item label="订单号">{{ current.order_no }}</el-descriptions-item>
            <el-descriptions-item label="用户 ID">{{ current.user_id }}</el-descriptions-item>
            <el-descriptions-item label="类型">
              {{ orderTypeLabel(current.order_type) }}
            </el-descriptions-item>
            <el-descriptions-item label="状态">
              {{ orderStatusLabel(current.status) }}
            </el-descriptions-item>
            <el-descriptions-item label="支付方式">
              {{ payTypeLabel(current.pay_type) }}
            </el-descriptions-item>
            <el-descriptions-item label="总价">
              ¥{{ formatFenToYuan(current.total_amount) }}
            </el-descriptions-item>
            <el-descriptions-item label="现金">
              ¥{{ formatFenToYuan(current.cash_amount) }}
            </el-descriptions-item>
            <el-descriptions-item label="积分">{{ current.point_amount }}</el-descriptions-item>
            <el-descriptions-item label="汇率快照">
              1 元 = {{ current.points_per_yuan }} 积分
            </el-descriptions-item>
            <el-descriptions-item label="支付截止">
              {{ formatUnix(current.pay_expire_at) }}
            </el-descriptions-item>
            <el-descriptions-item label="支付时间">
              {{ formatUnix(current.paid_at) }}
            </el-descriptions-item>
            <el-descriptions-item label="取消原因">
              {{ cancelReasonLabel(current.cancel_reason) }}
            </el-descriptions-item>
            <el-descriptions-item label="创建时间">
              {{ formatUnix(current.ct) }}
            </el-descriptions-item>
          </el-descriptions>

          <h4 class="items-title">明细</h4>
          <el-table :data="current.items || []" size="small" border>
            <el-table-column prop="product_id" label="商品ID" width="80" />
            <el-table-column prop="product_name" label="名称" min-width="120" />
            <el-table-column label="单价" width="90">
              <template #default="{ row }">{{ formatFenToYuan(row.product_price) }}</template>
            </el-table-column>
            <el-table-column prop="quantity" label="数量" width="70" />
            <el-table-column label="小计" width="90">
              <template #default="{ row }">{{ formatFenToYuan(row.amount) }}</template>
            </el-table-column>
          </el-table>
        </template>
      </div>
    </el-drawer>
  </div>
</template>

<style scoped>
.pager {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

.items-title {
  margin: 20px 0 10px;
}
</style>
