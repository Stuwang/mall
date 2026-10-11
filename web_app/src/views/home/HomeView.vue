<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import type { AnnouncementInfo, BannerInfo, CategoryInfo, ProductInfo } from '@/api/types'
import { appContext } from '@/app/AppContext'
import { routeNames } from '@/router/routes'
import { formatFenToYuan } from '@/utils/format'
import { notifyError } from '@/utils/notify'

const router = useRouter()
const productApi = appContext.api.products
const miscApi = appContext.api.misc

const refreshing = ref(false)
const banners = ref<BannerInfo[]>([])
const categories = ref<CategoryInfo[]>([])
const announcements = ref<AnnouncementInfo[]>([])
const hotProducts = ref<ProductInfo[]>([])

const noticeText = computed(() =>
  announcements.value
    .map((item) => item.content)
    .filter(Boolean)
    .join('　　'),
)

function coverOf(product: ProductInfo): string {
  return product.images?.[0] || ''
}

function openDetail(id: number): void {
  void router.push({ name: routeNames.productDetail, params: { id: String(id) } })
}

function openCatalog(categoryId?: number): void {
  void router.push({
    name: routeNames.catalog,
    query: categoryId ? { category_id: String(categoryId) } : undefined,
  })
}

async function loadHome(): Promise<void> {
  try {
    const [bannerReply, categoryReply, announcementReply, hotReply] = await Promise.all([
      miscApi.listBanners({}),
      productApi.listCategories({}),
      miscApi.listAnnouncements({}),
      productApi.listHotProducts({ page: 1, page_size: 20 }),
    ])
    banners.value = bannerReply.list ?? []
    categories.value = categoryReply.list ?? []
    announcements.value = announcementReply.list ?? []
    hotProducts.value = hotReply.list ?? []
  } catch (error) {
    notifyError(error, '首页加载失败')
  } finally {
    refreshing.value = false
  }
}

async function onRefresh(): Promise<void> {
  refreshing.value = true
  await loadHome()
}

onMounted(() => {
  void loadHome()
})
</script>

<template>
  <div class="home">
    <van-pull-refresh v-model="refreshing" @refresh="onRefresh">
      <section v-if="banners.length > 0" class="banner-wrap">
        <van-swipe class="banner-swipe" :autoplay="4000" indicator-color="#fff">
          <van-swipe-item v-for="item in banners" :key="item.id">
            <van-image class="banner-img" fit="cover" :src="item.image_url" />
          </van-swipe-item>
        </van-swipe>
      </section>

      <section class="section category-section">
        <div class="section-head">
          <h2 class="section-title">商品分类</h2>
          <button type="button" class="link-btn" @click="openCatalog()">全部</button>
        </div>
        <div v-if="categories.length > 0" class="category-grid">
          <button
            v-for="item in categories"
            :key="item.id"
            type="button"
            class="category-item"
            @click="openCatalog(item.id)"
          >
            <span class="category-icon">{{ item.name.slice(0, 1) }}</span>
            <span class="category-name">{{ item.name }}</span>
          </button>
        </div>
        <van-empty v-else description="暂无分类" image-size="80" />
      </section>

      <section v-if="noticeText" class="notice-section">
        <van-notice-bar left-icon="volume-o" :text="noticeText" />
      </section>

      <section class="section hot-section">
        <div class="section-head">
          <h2 class="section-title">爆款推荐</h2>
        </div>
        <div v-if="hotProducts.length > 0" class="product-grid">
          <div
            v-for="item in hotProducts"
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
        <van-empty v-else description="暂无爆款" image-size="80" />
      </section>
    </van-pull-refresh>
  </div>
</template>

<style scoped>
.home {
  min-height: 100%;
  background: #f7f8fa;
}

.banner-wrap {
  background: #fff;
}

.banner-swipe {
  height: 160px;
}

.banner-img {
  display: block;
  width: 100%;
  height: 160px;
  background: #f2f3f5;
}

.section {
  margin-top: 10px;
  padding: 12px 12px 4px;
  background: #fff;
}

.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
}

.section-title {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #323233;
}

.link-btn {
  padding: 0;
  border: 0;
  background: transparent;
  color: #969799;
  font-size: 13px;
}

.category-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px 8px;
  padding-bottom: 8px;
}

.category-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 0;
  border: 0;
  background: transparent;
  color: inherit;
}

.category-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  border-radius: 12px;
  background: #fff1f0;
  color: #ee0a24;
  font-size: 16px;
  font-weight: 600;
}

.category-name {
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 12px;
  color: #323233;
}

.notice-section {
  margin-top: 10px;
}

.hot-section {
  padding-bottom: 16px;
  margin-bottom: 12px;
}

.product-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}

.product-card {
  overflow: hidden;
  background: #fff;
  border-radius: 10px;
  border: 1px solid #f2f3f5;
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
