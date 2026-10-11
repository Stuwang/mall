<script setup lang="ts">
import { ElMessage, ElMessageBox } from 'element-plus'
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  OrderStatus,
  ProductStatus,
  type CategoryInfo,
  type OrderInfo,
  type ProductInfo,
} from '@/api/types'
import { appContext } from '@/app/AppContext'
import { routeNames } from '@/router/routes'
import { formatFenToYuan, formatUnix } from '@/utils/format'
import { orderStatusLabel, orderTypeLabel, productStatusLabel } from '@/utils/labels'
import { showError } from '@/utils/notify'

const route = useRoute()
const router = useRouter()
const productApi = appContext.api.products
const categoryApi = appContext.api.categories
const orderApi = appContext.api.orders

const productId = computed(() => Number(route.params.id))

const loading = ref(false)
const product = ref<ProductInfo | null>(null)
const categories = ref<CategoryInfo[]>([])

const ordersLoading = ref(false)
const orders = ref<OrderInfo[]>([])
const ordersTotal = ref(0)
const orderQuery = reactive({
  page: 1,
  page_size: 10,
})

function categoryName(id: number): string {
  return categories.value.find((c) => c.id === id)?.name ?? String(id || '-')
}

async function loadCategories(): Promise<void> {
  const reply = await categoryApi.list({})
  categories.value = reply.list ?? []
}

async function loadProduct(): Promise<void> {
  if (!Number.isFinite(productId.value) || productId.value <= 0) {
    ElMessage.error('无效的商品 ID')
    void router.replace({ name: routeNames.products })
    return
  }
  loading.value = true
  try {
    const reply = await productApi.detail({ id: productId.value })
    product.value = reply.product
  } catch (error) {
    showError(error, '加载商品失败')
    void router.replace({ name: routeNames.products })
  } finally {
    loading.value = false
  }
}

async function loadOrders(): Promise<void> {
  if (!productId.value) return
  ordersLoading.value = true
  try {
    const reply = await orderApi.list({
      user_id: 0,
      product_id: productId.value,
      status: OrderStatus.UNSPECIFIED,
      page: orderQuery.page,
      page_size: orderQuery.page_size,
    })
    orders.value = reply.list ?? []
    ordersTotal.value = Number(reply.total ?? 0)
  } catch (error) {
    showError(error, '加载相关订单失败')
  } finally {
    ordersLoading.value = false
  }
}

async function reloadAll(): Promise<void> {
  try {
    await loadCategories()
  } catch (error) {
    showError(error, '加载分类失败')
  }
  await loadProduct()
  await loadOrders()
}

async function toggleStatus(): Promise<void> {
  if (!product.value) return
  const next = product.value.status === ProductStatus.ON ? ProductStatus.OFF : ProductStatus.ON
  try {
    await productApi.setStatus({ id: product.value.id, status: next })
    ElMessage.success(next === ProductStatus.ON ? '已上架' : '已下架')
    await loadProduct()
  } catch (error) {
    showError(error, '更新状态失败')
  }
}

async function quickStock(): Promise<void> {
  if (!product.value) return
  try {
    const { value } = await ElMessageBox.prompt('请输入库存数量', '调整库存', {
      inputValue: String(product.value.stock),
      inputPattern: /^\d+$/,
      inputErrorMessage: '请输入非负整数',
    })
    await productApi.updateStock({ id: product.value.id, stock: Number(value) })
    ElMessage.success('库存已更新')
    await loadProduct()
  } catch (error) {
    if (error === 'cancel' || error === 'close') {
      return
    }
    showError(error, '更新库存失败')
  }
}

function goBack(): void {
  void router.push({ name: routeNames.products })
}

function goUser(userId: number): void {
  if (!userId) return
  void router.push({ name: routeNames.userDetail, params: { id: userId } })
}

function goOrderList(): void {
  void router.push({ name: routeNames.orders, query: { product_id: String(productId.value) } })
}

watch(
  () => route.params.id,
  () => {
    orderQuery.page = 1
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
      <el-button type="warning" @click="quickStock">调整库存</el-button>
      <el-button type="danger" @click="toggleStatus">
        {{ product?.status === ProductStatus.ON ? '下架' : '上架' }}
      </el-button>
    </div>

    <template v-if="product">
      <el-descriptions title="商品信息" :column="3" border>
        <el-descriptions-item label="商品 ID">{{ product.id }}</el-descriptions-item>
        <el-descriptions-item label="名称">{{ product.name }}</el-descriptions-item>
        <el-descriptions-item label="分类">
          {{ categoryName(product.category_id) }}
          <span class="muted">({{ product.category_id }})</span>
        </el-descriptions-item>
        <el-descriptions-item label="价格">¥{{ formatFenToYuan(product.price) }}</el-descriptions-item>
        <el-descriptions-item label="库存">{{ product.stock }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="product.status === ProductStatus.ON ? 'success' : 'info'" size="small">
            {{ productStatusLabel(product.status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatUnix(product.ct) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ formatUnix(product.ut) }}</el-descriptions-item>
        <el-descriptions-item label="描述" :span="3">
          {{ product.description || '-' }}
        </el-descriptions-item>
      </el-descriptions>

      <div v-if="product.images?.length" class="images">
        <div class="images-title">图片</div>
        <div class="images-list">
          <a
            v-for="(url, index) in product.images"
            :key="`${url}-${index}`"
            :href="url"
            target="_blank"
            rel="noopener noreferrer"
            class="image-item"
          >
            <img :src="url" :alt="`image-${index + 1}`" />
          </a>
        </div>
      </div>

      <div class="section-head">
        <h3>相关订单</h3>
        <el-button link type="primary" @click="goOrderList">在订单管理中查看</el-button>
      </div>
      <el-table v-loading="ordersLoading" :data="orders" stripe border>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="order_no" label="订单号" min-width="180" />
        <el-table-column label="用户" width="100">
          <template #default="{ row }">
            <el-button link type="primary" @click="goUser(row.user_id)">{{ row.user_id }}</el-button>
          </template>
        </el-table-column>
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
    </template>
  </div>
</template>

<style scoped>
.muted {
  margin-left: 4px;
  color: #9ca3af;
  font-size: 12px;
}

.images {
  margin-top: 20px;
}

.images-title {
  margin-bottom: 10px;
  font-weight: 600;
}

.images-list {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.image-item {
  display: block;
  width: 96px;
  height: 96px;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  overflow: hidden;
  background: #f9fafb;
}

.image-item img {
  width: 100%;
  height: 100%;
  object-fit: cover;
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
</style>
