<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { showConfirmDialog } from 'vant'
import { useRouter } from 'vue-router'
import { AssetType } from '@/api/types'
import { appContext } from '@/app/AppContext'
import { routeNames } from '@/router/routes'
import { useAuthStore } from '@/stores/authStore'
import { useCartStore } from '@/stores/cartStore'
import { notifyError, notifySuccess } from '@/utils/notify'

const router = useRouter()
const authStore = useAuthStore()
const cartStore = useCartStore()
const assetApi = appContext.api.assets

const pointBalance = ref(0)
const loadingAssets = ref(false)

const subtitle = computed(() =>
  authStore.isLoggedIn ? `@${authStore.user?.username ?? ''}` : '登录后查看订单与积分',
)

async function loadAssets(): Promise<void> {
  if (!authStore.isLoggedIn) {
    pointBalance.value = 0
    return
  }
  loadingAssets.value = true
  try {
    const reply = await assetApi.getUserAssets({ type: AssetType.POINT })
    const point = (reply.assets ?? []).find((item) => item.type === AssetType.POINT)
    pointBalance.value = point?.balance ?? 0
  } catch (error) {
    notifyError(error, '积分加载失败')
  } finally {
    loadingAssets.value = false
  }
}

async function ensureLogin(): Promise<boolean> {
  if (authStore.isLoggedIn) return true
  await router.push({ name: routeNames.login, query: { redirect: '/mine' } })
  return false
}

async function goOrders(): Promise<void> {
  if (!(await ensureLogin())) return
  await router.push({ name: routeNames.orders })
}

async function onLogout(): Promise<void> {
  try {
    await showConfirmDialog({ title: '退出登录', message: '确定退出当前账号？' })
    authStore.logout()
    cartStore.clearLocal()
    pointBalance.value = 0
    notifySuccess('已退出')
  } catch {
    /* cancel */
  }
}

onMounted(async () => {
  if (authStore.isLoggedIn) {
    try {
      await authStore.refreshProfile()
    } catch (error) {
      notifyError(error, '刷新用户信息失败')
    }
    await loadAssets()
  }
})
</script>

<template>
  <div class="mine-page">
    <div class="profile">
      <div class="avatar">{{ authStore.displayName.slice(0, 1) }}</div>
      <div class="info">
        <div class="name">{{ authStore.isLoggedIn ? authStore.displayName : '未登录' }}</div>
        <div class="muted">{{ subtitle }}</div>
      </div>
      <van-button
        v-if="!authStore.isLoggedIn"
        size="small"
        type="primary"
        round
        @click="router.push({ name: routeNames.login })"
      >
        登录
      </van-button>
    </div>

    <van-cell-group inset class="block">
      <van-cell title="我的积分" :value="loadingAssets ? '...' : String(pointBalance)" is-link @click="ensureLogin" />
      <van-cell title="我的订单" is-link @click="goOrders" />
      <van-cell title="购物车" :value="String(cartStore.totalQuantity)" is-link @click="router.push({ name: routeNames.cart })" />
    </van-cell-group>

    <van-cell-group v-if="authStore.isLoggedIn" inset class="block">
      <van-cell title="退出登录" is-link @click="onLogout" />
    </van-cell-group>

    <van-cell-group v-else inset class="block">
      <van-cell title="注册账号" is-link @click="router.push({ name: routeNames.register })" />
    </van-cell-group>
  </div>
</template>

<style scoped>
.mine-page {
  padding-top: 12px;
}

.profile {
  display: flex;
  gap: 12px;
  align-items: center;
  margin: 0 12px 16px;
  padding: 20px 16px;
  background: linear-gradient(135deg, #ff6034, #ee0a24);
  border-radius: 12px;
  color: #fff;
}

.avatar {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 52px;
  height: 52px;
  font-size: 22px;
  font-weight: 600;
  background: rgba(255, 255, 255, 0.22);
  border-radius: 50%;
}

.info {
  flex: 1;
  min-width: 0;
}

.name {
  margin-bottom: 4px;
  font-size: 18px;
  font-weight: 600;
}

.profile .muted {
  color: rgba(255, 255, 255, 0.85);
}

.block {
  margin-bottom: 12px;
}
</style>
