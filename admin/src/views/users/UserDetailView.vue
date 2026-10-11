<script setup lang="ts">
import { ElMessage, ElMessageBox } from 'element-plus'
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { AssetType, OrderStatus, type AssetHistory, type OrderInfo, type UserInfo } from '@/api/types'
import { appContext } from '@/app/AppContext'
import { routeNames } from '@/router/routes'
import { formatFenToYuan, formatUnix } from '@/utils/format'
import {
  assetBizTypeLabel,
  orderStatusLabel,
  orderTypeLabel,
} from '@/utils/labels'
import { showError } from '@/utils/notify'

const route = useRoute()
const router = useRouter()
const userApi = appContext.api.users
const assetApi = appContext.api.assets
const orderApi = appContext.api.orders

const userId = computed(() => Number(route.params.id))

const loading = ref(false)
const user = ref<UserInfo | null>(null)
const pointBalance = ref(0)

const ordersLoading = ref(false)
const orders = ref<OrderInfo[]>([])
const ordersTotal = ref(0)
const orderQuery = reactive({
  page: 1,
  page_size: 10,
})

const historyLoading = ref(false)
const history = ref<AssetHistory[]>([])
const historyTotal = ref(0)
const historyQuery = reactive({
  page: 1,
  page_size: 10,
})

const grantVisible = ref(false)
const grantSubmitting = ref(false)
const grantForm = reactive({
  points: 100,
  remark: '',
})

async function loadUser(): Promise<void> {
  if (!Number.isFinite(userId.value) || userId.value <= 0) {
    ElMessage.error('无效的用户 ID')
    void router.replace({ name: routeNames.users })
    return
  }
  loading.value = true
  try {
    const [userReply, assetReply] = await Promise.all([
      userApi.detail({ user_id: userId.value }),
      assetApi.getUserAssets({ user_id: userId.value, type: AssetType.POINT }),
    ])
    user.value = userReply.user
    const point = (assetReply.assets ?? []).find((item) => item.type === AssetType.POINT)
    pointBalance.value = point?.balance ?? 0
  } catch (error) {
    showError(error, '加载用户失败')
    void router.replace({ name: routeNames.users })
  } finally {
    loading.value = false
  }
}

async function loadOrders(): Promise<void> {
  if (!userId.value) return
  ordersLoading.value = true
  try {
    const reply = await orderApi.list({
      user_id: userId.value,
      product_id: 0,
      status: OrderStatus.UNSPECIFIED,
      page: orderQuery.page,
      page_size: orderQuery.page_size,
    })
    orders.value = reply.list ?? []
    ordersTotal.value = Number(reply.total ?? 0)
  } catch (error) {
    showError(error, '加载订单失败')
  } finally {
    ordersLoading.value = false
  }
}

async function loadHistory(): Promise<void> {
  if (!userId.value) return
  historyLoading.value = true
  try {
    const reply = await assetApi.getHistory({
      user_id: userId.value,
      type: AssetType.POINT,
      page: historyQuery.page,
      page_size: historyQuery.page_size,
    })
    history.value = reply.list ?? []
    historyTotal.value = Number(reply.total ?? 0)
  } catch (error) {
    showError(error, '加载积分明细失败')
  } finally {
    historyLoading.value = false
  }
}

async function reloadAll(): Promise<void> {
  await loadUser()
  await Promise.all([loadOrders(), loadHistory()])
}

async function handleBan(): Promise<void> {
  if (!user.value) return
  try {
    const { value } = await ElMessageBox.prompt(
      '请输入封禁截止时间（本地时间，格式 YYYY-MM-DD HH:mm:ss）。留空并确认将解封。',
      user.value.banned ? '调整封禁 / 解封' : '封禁用户',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        inputPlaceholder: '例如 2026-12-31 23:59:59，空=解封',
        inputValue: '',
      },
    )
    const trimmed = value.trim()
    let banUntil = 0
    if (trimmed) {
      const ts = Math.floor(new Date(trimmed.replace(/-/g, '/')).getTime() / 1000)
      if (!Number.isFinite(ts) || ts <= 0) {
        ElMessage.error('时间格式无效')
        return
      }
      banUntil = ts
    }
    await userApi.ban({ user_id: user.value.id, ban_until: banUntil })
    ElMessage.success(banUntil === 0 ? '已解封' : '已封禁')
    await loadUser()
  } catch (error) {
    if (error === 'cancel' || error === 'close') {
      return
    }
    showError(error, '操作失败')
  }
}

function openGrant(): void {
  grantForm.points = 100
  grantForm.remark = ''
  grantVisible.value = true
}

async function submitGrant(): Promise<void> {
  if (!user.value) return
  if (!grantForm.points || grantForm.points <= 0) {
    ElMessage.error('积分数必须大于 0')
    return
  }
  grantSubmitting.value = true
  try {
    const reply = await assetApi.grantPoints({
      user_id: user.value.id,
      points: grantForm.points,
      remark: grantForm.remark.trim() || undefined,
    })
    pointBalance.value = reply.asset?.balance ?? pointBalance.value + grantForm.points
    ElMessage.success('加积分成功')
    grantVisible.value = false
    historyQuery.page = 1
    await loadHistory()
  } catch (error) {
    showError(error, '加积分失败')
  } finally {
    grantSubmitting.value = false
  }
}

function goBack(): void {
  void router.push({ name: routeNames.users })
}

function goOrderList(): void {
  void router.push({ name: routeNames.orders, query: { user_id: String(userId.value) } })
}

function goProduct(productId: number): void {
  if (!productId) return
  void router.push({ name: routeNames.productDetail, params: { id: productId } })
}

watch(
  () => route.params.id,
  () => {
    orderQuery.page = 1
    historyQuery.page = 1
    void reloadAll()
  },
)

onMounted(() => {
  void reloadAll()
})
</script>

<template>
  <div v-loading="loading" class="page-card">
    <div class="page-toolbar">
      <el-button @click="goBack">返回列表</el-button>
      <div class="grow" />
      <el-button type="primary" @click="openGrant">加积分</el-button>
      <el-button @click="handleBan">{{ user?.banned ? '解封/改期' : '封禁' }}</el-button>
    </div>

    <template v-if="user">
      <el-descriptions title="基本信息" :column="3" border>
        <el-descriptions-item label="用户 ID">{{ user.id }}</el-descriptions-item>
        <el-descriptions-item label="账号">{{ user.username }}</el-descriptions-item>
        <el-descriptions-item label="昵称">{{ user.nickname || '-' }}</el-descriptions-item>
        <el-descriptions-item label="管理员">
          <el-tag :type="user.is_admin ? 'danger' : 'info'" size="small">
            {{ user.is_admin ? '是' : '否' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="封禁状态">
          <el-tag :type="user.banned ? 'warning' : 'success'" size="small">
            {{ user.banned ? '封禁中' : '正常' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="封禁至">{{ formatUnix(user.ban_until) }}</el-descriptions-item>
        <el-descriptions-item label="当前积分">
          <span class="points">{{ pointBalance }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatUnix(user.ct) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ formatUnix(user.ut) }}</el-descriptions-item>
      </el-descriptions>

      <div class="section-head">
        <h3>订单列表</h3>
        <el-button link type="primary" @click="goOrderList">在订单管理中查看</el-button>
      </div>
      <el-table v-loading="ordersLoading" :data="orders" stripe border>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="order_no" label="订单号" min-width="180" />
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
        <el-table-column prop="point_amount" label="积分" width="90" />
        <el-table-column label="商品" min-width="160">
          <template #default="{ row }">
            <div v-if="row.items?.length" class="item-links">
              <el-button
                v-for="item in row.items"
                :key="item.id"
                link
                type="primary"
                @click="goProduct(item.product_id)"
              >
                {{ item.product_name || `#${item.product_id}` }}
              </el-button>
            </div>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="下单时间" min-width="160">
          <template #default="{ row }">{{ formatUnix(row.ct) }}</template>
        </el-table-column>
      </el-table>
      <div class="pager">
        <el-pagination
          v-model:current-page="orderQuery.page"
          v-model:page-size="orderQuery.page_size"
          background
          layout="total, prev, pager, next, sizes"
          :total="ordersTotal"
          :page-sizes="[10, 20, 50]"
          @current-change="loadOrders"
          @size-change="
            () => {
              orderQuery.page = 1
              loadOrders()
            }
          "
        />
      </div>

      <div class="section-head">
        <h3>积分明细</h3>
      </div>
      <el-table v-loading="historyLoading" :data="history" stripe border>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column label="业务类型" min-width="120">
          <template #default="{ row }">{{ assetBizTypeLabel(row.biz_type) }}</template>
        </el-table-column>
        <el-table-column prop="biz_id" label="业务 ID" min-width="100" />
        <el-table-column label="变更" width="100">
          <template #default="{ row }">
            <span :class="row.change >= 0 ? 'plus' : 'minus'">
              {{ row.change >= 0 ? `+${row.change}` : row.change }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="balance" label="变更后余额" width="110" />
        <el-table-column prop="remark" label="备注" min-width="140" />
        <el-table-column label="时间" min-width="160">
          <template #default="{ row }">{{ formatUnix(row.ct) }}</template>
        </el-table-column>
      </el-table>
      <div class="pager">
        <el-pagination
          v-model:current-page="historyQuery.page"
          v-model:page-size="historyQuery.page_size"
          background
          layout="total, prev, pager, next, sizes"
          :total="historyTotal"
          :page-sizes="[10, 20, 50]"
          @current-change="loadHistory"
          @size-change="
            () => {
              historyQuery.page = 1
              loadHistory()
            }
          "
        />
      </div>
    </template>

    <el-dialog v-model="grantVisible" title="手动加积分" width="480px" destroy-on-close>
      <el-form label-width="80px">
        <el-form-item label="积分数" required>
          <el-input-number v-model="grantForm.points" :min="1" :step="100" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input
            v-model="grantForm.remark"
            maxlength="64"
            show-word-limit
            placeholder="可选，如活动赠送"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="grantVisible = false">取消</el-button>
        <el-button type="primary" :loading="grantSubmitting" @click="submitGrant">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.points {
  font-weight: 700;
  color: #d97706;
}

.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin: 24px 0 12px;
}

.section-head h3 {
  margin: 0;
  font-size: 16px;
}

.pager {
  display: flex;
  justify-content: flex-end;
  margin-top: 12px;
}

.item-links {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.plus {
  color: #16a34a;
}

.minus {
  color: #dc2626;
}
</style>
