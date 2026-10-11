<script setup lang="ts">
import { nextTick, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { appContext } from '@/app/AppContext'
import type { CategoryInfo, ProductInfo } from '@/api/types'
import { routeNames } from '@/router/routes'
import { formatFenToYuan } from '@/utils/format'
import { notifyError } from '@/utils/notify'

const router = useRouter()
const api = appContext.api.products

const keyword = ref('')
const categories = ref<CategoryInfo[]>([])
const activeCategoryId = ref(0)
const products = ref<ProductInfo[]>([])
const page = ref(1)
const total = ref(0)

/**
 * 仅绑定给 van-list。
 * 注意：List 触发 @load 前会先把该值设为 true，因此绝不能用它做早退条件。
 */
const listLoading = ref(false)
const refreshing = ref(false)
const finished = ref(false)

/** 真实请求互斥（不用 ref，避免与 List 的 loading 耦合） */
let inFlight = false

async function loadCategories(): Promise<void> {
  try {
    const reply = await api.listCategories({})
    categories.value = reply.list ?? []
  } catch (error) {
    notifyError(error, '分类加载失败')
  }
}

async function fetchProducts(reset: boolean): Promise<void> {
  if (inFlight) {
    // 若是 List 触发的多余 load，必须把 loading 拨回 false，否则会一直转圈
    listLoading.value = false
    return
  }

  inFlight = true
  listLoading.value = true

  if (reset) {
    page.value = 1
    finished.value = false
    products.value = []
  }

  try {
    const pageSize = 10
    const reply =
      activeCategoryId.value > 0
        ? await api.listProductsByCategory({
            category_id: activeCategoryId.value,
            page: page.value,
            page_size: pageSize,
          })
        : await api.listProducts({
            keyword: keyword.value.trim() || undefined,
            page: page.value,
            page_size: pageSize,
          })

    const list = reply.list ?? []
    products.value = reset ? list : products.value.concat(list)
    total.value = Number(reply.total ?? 0)

    const noMore =
      list.length === 0 ||
      list.length < pageSize ||
      (total.value > 0 && products.value.length >= total.value)

    finished.value = noMore
    if (!noMore) {
      page.value += 1
    }
  } catch (error) {
    finished.value = true
    notifyError(error, '商品加载失败')
  } finally {
    inFlight = false
    listLoading.value = false
    refreshing.value = false
  }
}

/** van-list 触底 / 首次 check */
function onListLoad(): void {
  if (finished.value) {
    listLoading.value = false
    return
  }
  void fetchProducts(false)
}

async function reload(): Promise<void> {
  finished.value = false
  await fetchProducts(true)
}

async function onRefresh(): Promise<void> {
  refreshing.value = true
  await reload()
}

function onSearch(): void {
  activeCategoryId.value = 0
  void reload()
}

function onCategoryChange(name: number | string): void {
  activeCategoryId.value = Number(name)
  keyword.value = ''
  void reload()
}

function openDetail(id: number): void {
  void router.push({ name: routeNames.productDetail, params: { id: String(id) } })
}

function coverOf(product: ProductInfo): string {
  return product.images?.[0] || ''
}

onMounted(async () => {
  await loadCategories()
  // 关闭 List 的 immediate-check，改由这里主动拉第一页，避免与 tabs 初始化抢 loading
  await nextTick()
  await reload()
})
</script>

<template>
  <div class="home">
    <van-sticky>
      <div class="search-bar">
        <van-search
          v-model="keyword"
          shape="round"
          placeholder="搜索商品"
          show-action
          @search="onSearch"
        >
          <template #action>
            <div @click="onSearch">搜索</div>
          </template>
        </van-search>
      </div>
      <van-tabs
        :model-value="activeCategoryId"
        shrink
        color="#ee0a24"
        @change="onCategoryChange"
      >
        <van-tab title="全部" :name="0" />
        <van-tab
          v-for="item in categories"
          :key="item.id"
          :title="item.name"
          :name="item.id"
        />
      </van-tabs>
    </van-sticky>

    <van-pull-refresh v-model="refreshing" @refresh="onRefresh">
      <van-list
        v-model:loading="listLoading"
        :finished="finished"
        finished-text="没有更多了"
        :immediate-check="false"
        @load="onListLoad"
      >
        <div v-if="products.length > 0" class="product-grid">
          <div
            v-for="item in products"
            :key="item.id"
            class="product-card"
            @click="openDetail(item.id)"
          >
            <van-image class="cover" fit="cover" :src="coverOf(item)" />
            <div class="info">
              <div class="name">{{ item.name }}</div>
              <div class="row">
                <span class="price">{{ formatFenToYuan(item.price) }}</span>
                <span class="muted">库存 {{ item.stock }}</span>
              </div>
            </div>
          </div>
        </div>
        <van-empty
          v-else-if="finished && !listLoading"
          description="暂无商品"
        />
      </van-list>
    </van-pull-refresh>
  </div>
</template>

<style scoped>
.home {
  min-height: 100%;
}

.home :deep(.van-pull-refresh) {
  min-height: 60vh;
}

.search-bar {
  background: #fff;
}

.product-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
  padding: 12px;
}

.product-card {
  overflow: hidden;
  background: #fff;
  border-radius: 10px;
}

.cover {
  display: block;
  width: 100%;
  height: 140px;
  background: #f2f3f5;
}

.info {
  padding: 8px 10px 12px;
}

.name {
  height: 40px;
  margin-bottom: 6px;
  overflow: hidden;
  font-size: 14px;
  line-height: 20px;
}

.row {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 8px;
}
</style>
