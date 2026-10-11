<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { appContext } from '@/app/AppContext'
import type { ProductInfo } from '@/api/types'
import { routeNames } from '@/router/routes'
import { useAuthStore } from '@/stores/authStore'
import { useCartStore } from '@/stores/cartStore'
import { formatFenToYuan } from '@/utils/format'
import { notifyError, notifySuccess } from '@/utils/notify'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const cartStore = useCartStore()
const api = appContext.api.products

const product = ref<ProductInfo | null>(null)
const loading = ref(true)
const submitting = ref(false)

const productId = computed(() => Number(route.params.id))
const images = computed(() => product.value?.images?.filter(Boolean) ?? [])

async function loadDetail(): Promise<void> {
  loading.value = true
  try {
    const reply = await api.getProduct({ id: productId.value })
    product.value = reply.product
  } catch (error) {
    notifyError(error, '商品不存在或已下架')
    product.value = null
  } finally {
    loading.value = false
  }
}

async function addToCart(): Promise<void> {
  if (!product.value) return
  if (!authStore.isLoggedIn) {
    await router.push({
      name: routeNames.login,
      query: { redirect: route.fullPath },
    })
    return
  }
  submitting.value = true
  try {
    await cartStore.add(product.value.id, 1)
    notifySuccess('已加入购物车')
  } catch (error) {
    notifyError(error, '加入购物车失败')
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  void loadDetail()
})
</script>

<template>
  <div class="page detail-page">
    <van-nav-bar title="商品详情" left-arrow fixed placeholder @click-left="router.back()" />

    <van-skeleton title :row="6" :loading="loading">
      <template v-if="product">
        <van-swipe v-if="images.length" class="banner" :autoplay="3000" indicator-color="white">
          <van-swipe-item v-for="(img, index) in images" :key="index">
            <van-image width="100%" height="100%" fit="cover" :src="img" />
          </van-swipe-item>
        </van-swipe>
        <div v-else class="banner banner-empty">
          <van-icon name="photo-o" size="48" />
        </div>

        <div class="panel">
          <div class="price">{{ formatFenToYuan(product.price) }}</div>
          <h1 class="title">{{ product.name }}</h1>
          <p class="muted">库存 {{ product.stock }}</p>
        </div>

        <div class="panel">
          <h2>商品介绍</h2>
          <p class="desc">{{ product.description || '暂无介绍' }}</p>
        </div>
      </template>
      <van-empty v-else description="商品不可用" />
    </van-skeleton>

    <van-action-bar safe-area-inset-bottom>
      <van-action-bar-icon icon="shop-o" text="首页" @click="router.push({ name: routeNames.home })" />
      <van-action-bar-icon
        icon="shopping-cart-o"
        text="购物车"
        :badge="cartStore.totalQuantity || ''"
        @click="router.push({ name: routeNames.cart })"
      />
      <van-action-bar-button
        type="danger"
        text="加入购物车"
        :loading="submitting"
        :disabled="!product || product.stock <= 0"
        @click="addToCart"
      />
    </van-action-bar>
  </div>
</template>

<style scoped>
.detail-page {
  padding-bottom: 60px;
}

.banner {
  height: 320px;
  background: #f2f3f5;
}

.banner-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  color: #c8c9cc;
}

.panel {
  margin: 10px 0;
  padding: 16px;
  background: #fff;
}

.title {
  margin: 8px 0;
  font-size: 18px;
  font-weight: 600;
}

.panel h2 {
  margin: 0 0 8px;
  font-size: 15px;
}

.desc {
  margin: 0;
  color: #646566;
  white-space: pre-wrap;
  word-break: break-word;
}
</style>
