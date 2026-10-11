<script setup lang="ts">
import { nextTick, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import type { CategoryInfo, ProductInfo } from '@/api/types'
import { appContext } from '@/app/AppContext'
import { routeNames } from '@/router/routes'
import { formatFenToYuan } from '@/utils/format'
import { notifyError } from '@/utils/notify'

interface CategoryGroup {
  category: CategoryInfo
  products: ProductInfo[]
}

const route = useRoute()
const router = useRouter()
const api = appContext.api.products

const loading = ref(true)
const groups = ref<CategoryGroup[]>([])
const activeCategoryId = ref(0)
const rightPanel = ref<HTMLElement | null>(null)

/** 点击左侧时短暂忽略右侧 scroll 联动，避免跳动 */
let lockScrollSync = false
let lockTimer: ReturnType<typeof setTimeout> | null = null

function coverOf(product: ProductInfo): string {
  return product.images?.[0] || ''
}

function openDetail(id: number): void {
  void router.push({ name: routeNames.productDetail, params: { id: String(id) } })
}

function sectionId(categoryId: number): string {
  return `cat-section-${categoryId}`
}

async function loadCatalog(): Promise<void> {
  loading.value = true
  try {
    const categoryReply = await api.listCategories({})
    const categories = categoryReply.list ?? []
    const results = await Promise.all(
      categories.map(async (category) => {
        const reply = await api.listProductsByCategory({
          category_id: category.id,
          page: 1,
          page_size: 100,
        })
        return {
          category,
          products: reply.list ?? [],
        } satisfies CategoryGroup
      }),
    )
    groups.value = results

    const queryId = Number(route.query.category_id || 0)
    if (queryId > 0 && groups.value.some((g) => g.category.id === queryId)) {
      activeCategoryId.value = queryId
    } else {
      activeCategoryId.value = groups.value[0]?.category.id ?? 0
    }

    await nextTick()
    if (queryId > 0) {
      scrollToCategory(queryId, false)
    }
  } catch (error) {
    notifyError(error, '分类商品加载失败')
  } finally {
    loading.value = false
  }
}

function scrollToCategory(categoryId: number, smooth = true): void {
  const panel = rightPanel.value
  const el = document.getElementById(sectionId(categoryId))
  if (!panel || !el) return

  lockScrollSync = true
  if (lockTimer) clearTimeout(lockTimer)
  activeCategoryId.value = categoryId

  panel.scrollTo({
    top: el.offsetTop,
    behavior: smooth ? 'smooth' : 'auto',
  })

  lockTimer = setTimeout(() => {
    lockScrollSync = false
  }, 360)
}

function onCategoryClick(categoryId: number): void {
  scrollToCategory(categoryId, true)
}

function onRightScroll(): void {
  if (lockScrollSync || !rightPanel.value || groups.value.length === 0) return

  const panel = rightPanel.value
  const scrollTop = panel.scrollTop
  let currentId = groups.value[0]?.category.id ?? 0

  for (const group of groups.value) {
    const el = document.getElementById(sectionId(group.category.id))
    if (!el) continue
    if (scrollTop + 24 >= el.offsetTop) {
      currentId = group.category.id
    } else {
      break
    }
  }

  if (activeCategoryId.value !== currentId) {
    activeCategoryId.value = currentId
  }
}

watch(
  () => route.query.category_id,
  (value) => {
    const id = Number(value || 0)
    if (id > 0 && groups.value.some((g) => g.category.id === id)) {
      scrollToCategory(id, true)
    }
  },
)

onMounted(() => {
  void loadCatalog()
})
</script>

<template>
  <div class="catalog">
    <div v-if="loading" class="loading-wrap">
      <van-loading type="spinner" color="#ee0a24" />
    </div>

    <template v-else-if="groups.length > 0">
      <aside class="side">
        <button
          v-for="group in groups"
          :key="group.category.id"
          type="button"
          class="side-item"
          :class="{ active: activeCategoryId === group.category.id }"
          @click="onCategoryClick(group.category.id)"
        >
          {{ group.category.name }}
        </button>
      </aside>

      <main ref="rightPanel" class="right" @scroll="onRightScroll">
        <section
          v-for="group in groups"
          :id="sectionId(group.category.id)"
          :key="group.category.id"
          class="group"
        >
          <h3 class="group-title">{{ group.category.name }}</h3>
          <div v-if="group.products.length > 0" class="product-list">
            <div
              v-for="item in group.products"
              :key="item.id"
              class="product-row"
              @click="openDetail(item.id)"
            >
              <van-image class="thumb" fit="cover" :src="coverOf(item)" radius="8" />
              <div class="meta">
                <div class="name">{{ item.name }}</div>
                <div class="desc muted">{{ item.description || '暂无描述' }}</div>
                <div class="row">
                  <span class="price">{{ formatFenToYuan(item.price) }}</span>
                  <span class="muted">库存 {{ item.stock }}</span>
                </div>
              </div>
            </div>
          </div>
          <van-empty v-else description="该分类暂无商品" image-size="64" />
        </section>
        <div class="bottom-space" />
      </main>
    </template>

    <van-empty v-else description="暂无分类商品" />
  </div>
</template>

<style scoped>
.catalog {
  display: flex;
  height: calc(100vh - 50px);
  height: calc(100dvh - 50px);
  background: #f7f8fa;
  overflow: hidden;
}

.loading-wrap {
  display: flex;
  flex: 1;
  align-items: center;
  justify-content: center;
}

.side {
  flex: 0 0 88px;
  width: 88px;
  overflow-y: auto;
  background: #f2f3f5;
  -webkit-overflow-scrolling: touch;
}

.side-item {
  position: relative;
  display: block;
  width: 100%;
  padding: 16px 8px;
  border: 0;
  background: transparent;
  color: #646566;
  font-size: 13px;
  line-height: 1.3;
  text-align: center;
  word-break: break-all;
}

.side-item.active {
  background: #fff;
  color: #ee0a24;
  font-weight: 600;
}

.side-item.active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  width: 3px;
  height: 16px;
  margin-top: -8px;
  background: #ee0a24;
  border-radius: 0 2px 2px 0;
}

.right {
  position: relative;
  flex: 1;
  overflow-y: auto;
  background: #fff;
  -webkit-overflow-scrolling: touch;
}

.group {
  padding: 12px 12px 4px;
}

.group-title {
  position: sticky;
  top: 0;
  z-index: 1;
  margin: 0 0 10px;
  padding: 8px 0;
  background: #fff;
  font-size: 14px;
  font-weight: 600;
  color: #323233;
}

.product-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.product-row {
  display: flex;
  gap: 10px;
}

.thumb {
  flex: 0 0 72px;
  width: 72px;
  height: 72px;
  background: #f2f3f5;
}

.meta {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.name {
  font-size: 14px;
  font-weight: 500;
  color: #323233;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.desc {
  font-size: 12px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.row {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  margin-top: auto;
  gap: 8px;
}

.bottom-space {
  height: 40vh;
}
</style>
