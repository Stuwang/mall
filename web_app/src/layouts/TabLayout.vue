<script setup lang="ts">
import { computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { routeNames } from '@/router/routes'
import { useAuthStore } from '@/stores/authStore'
import { useCartStore } from '@/stores/cartStore'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const cartStore = useCartStore()

const tabRoutes = [
  routeNames.home,
  routeNames.catalog,
  routeNames.cart,
  routeNames.mine,
] as const

const active = computed(() => {
  const idx = tabRoutes.indexOf(route.name as (typeof tabRoutes)[number])
  return idx >= 0 ? idx : 0
})

const cartBadge = computed(() =>
  cartStore.totalQuantity > 0 ? String(cartStore.totalQuantity) : '',
)

watch(
  () => authStore.isLoggedIn,
  (loggedIn) => {
    if (loggedIn) {
      void cartStore.refresh()
    } else {
      cartStore.clearLocal()
    }
  },
  { immediate: true },
)

function onChange(index: number | string): void {
  const target = tabRoutes[Number(index)]
  if (target && route.name !== target) {
    void router.push({ name: target })
  }
}
</script>

<template>
  <div class="page page-with-tab">
    <RouterView />
    <van-tabbar :model-value="active" fixed placeholder safe-area-inset-bottom @change="onChange">
      <van-tabbar-item icon="home-o">首页</van-tabbar-item>
      <van-tabbar-item icon="apps-o">分类</van-tabbar-item>
      <van-tabbar-item icon="shopping-cart-o" :badge="cartBadge">购物车</van-tabbar-item>
      <van-tabbar-item icon="user-o">我的</van-tabbar-item>
    </van-tabbar>
  </div>
</template>
