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
const mallApi = appContext.api.mall

const pointBalance = ref(0)
const pointsPerYuan = ref(0)
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

async function loadMallInfo(): Promise<void> {
  try {
    const reply = await mallApi.getMallInfo({})
    pointsPerYuan.value = reply.points_per_yuan || 0
  } catch {
    /* 非关键 */
  }
}

async function ensureLogin(redirect = '/mine'): Promise<boolean> {
  if (authStore.isLoggedIn) return true
  await router.push({ name: routeNames.login, query: { redirect } })
  return false
}

async function goPoints(): Promise<void> {
  if (!(await ensureLogin('/points'))) return
  await router.push({ name: routeNames.pointsPurchase })
}

async function goOrders(): Promise<void> {
  if (!(await ensureLogin('/orders'))) return
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
  void loadMallInfo()
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
    <!-- 1. 用户信息 -->
    <section class="profile-card">
      <div class="profile-bg" />
      <div class="profile-body">
        <div class="avatar">{{ authStore.displayName.slice(0, 1) }}</div>
        <div class="info">
          <div class="name">{{ authStore.isLoggedIn ? authStore.displayName : '未登录' }}</div>
          <div class="sub">{{ subtitle }}</div>
        </div>
        <van-button
          v-if="!authStore.isLoggedIn"
          size="small"
          round
          color="#fff"
          style="color: #ee0a24"
          @click="router.push({ name: routeNames.login })"
        >
          登录
        </van-button>
        <van-button
          v-else
          size="small"
          round
          plain
          hairline
          color="#fff"
          @click="onLogout"
        >
          退出
        </van-button>
      </div>
    </section>

    <!-- 2. 积分 -->
    <section class="section-card points-card" @click="goPoints">
      <div class="section-head">
        <div>
          <div class="section-title">我的积分</div>
          <div class="section-desc">
            {{ pointsPerYuan > 0 ? `1 元 = ${pointsPerYuan} 积分 · 点击购买` : '点击购买积分' }}
          </div>
        </div>
        <van-icon name="arrow" class="arrow" />
      </div>
      <div class="points-value">
        <span class="num">{{ loadingAssets ? '...' : pointBalance }}</span>
        <span class="unit">积分</span>
      </div>
      <div class="points-action">去充值</div>
    </section>

    <!-- 3. 订单 -->
    <section class="section-card">
      <div class="section-head clickable" @click="goOrders">
        <div>
          <div class="section-title">我的订单</div>
          <div class="section-desc">查看全部订单状态</div>
        </div>
        <van-icon name="arrow" class="arrow" />
      </div>
      <div class="order-grid">
        <button type="button" class="order-item" @click="goOrders">
          <van-icon name="pending-payment" size="22" />
          <span>全部订单</span>
        </button>
        <button type="button" class="order-item" @click="router.push({ name: routeNames.cart })">
          <van-icon name="shopping-cart-o" size="22" />
          <span>购物车</span>
          <van-badge v-if="cartStore.totalQuantity > 0" :content="cartStore.totalQuantity" />
        </button>
      </div>
    </section>

    <!-- 4. 其他功能 -->
    <section class="section-card">
      <div class="section-title padded">其他功能</div>
      <van-cell title="帮助中心" is-link :border="false" @click="router.push({ name: routeNames.help })" />
      <van-cell title="关于我们" is-link :border="false" @click="router.push({ name: routeNames.about })" />
      <van-cell title="联系我们" is-link :border="false" @click="router.push({ name: routeNames.contact })" />
      <van-cell
        v-if="!authStore.isLoggedIn"
        title="注册账号"
        is-link
        :border="false"
        @click="router.push({ name: routeNames.register })"
      />
    </section>
  </div>
</template>

<style scoped>
.mine-page {
  min-height: 100%;
  padding: 12px 12px 24px;
  background:
    radial-gradient(120% 60% at 0% 0%, rgba(255, 120, 60, 0.18), transparent 55%),
    radial-gradient(90% 50% at 100% 10%, rgba(238, 10, 36, 0.12), transparent 50%),
    #f7f8fa;
}

.profile-card {
  position: relative;
  overflow: hidden;
  margin-bottom: 12px;
  border-radius: 16px;
}

.profile-bg {
  position: absolute;
  inset: 0;
  background: linear-gradient(135deg, #ff7a45 0%, #ee0a24 55%, #c2081c 100%);
}

.profile-body {
  position: relative;
  display: flex;
  gap: 12px;
  align-items: center;
  padding: 22px 16px;
  color: #fff;
}

.avatar {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 56px;
  height: 56px;
  font-size: 24px;
  font-weight: 700;
  background: rgba(255, 255, 255, 0.22);
  border: 2px solid rgba(255, 255, 255, 0.35);
  border-radius: 50%;
}

.info {
  flex: 1;
  min-width: 0;
}

.name {
  margin-bottom: 4px;
  overflow: hidden;
  font-size: 20px;
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.sub {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.88);
}

.section-card {
  margin-bottom: 12px;
  overflow: hidden;
  background: #fff;
  border-radius: 14px;
  box-shadow: 0 8px 24px rgba(50, 50, 51, 0.04);
}

.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 14px 0;
}

.section-head.clickable,
.points-card {
  cursor: pointer;
}

.section-title {
  font-size: 15px;
  font-weight: 650;
}

.section-title.padded {
  padding: 14px 16px 4px;
}

.section-desc {
  margin-top: 2px;
  font-size: 12px;
  color: #969799;
}

.arrow {
  color: #c8c9cc;
}

.points-card {
  padding-bottom: 14px;
  background: linear-gradient(180deg, #fff7f2 0%, #fff 55%);
}

.points-value {
  display: flex;
  gap: 6px;
  align-items: baseline;
  padding: 10px 14px 0;
}

.points-value .num {
  font-size: 34px;
  font-weight: 750;
  color: #ee0a24;
  line-height: 1;
}

.points-value .unit {
  font-size: 13px;
  color: #969799;
}

.points-action {
  display: inline-block;
  margin: 12px 14px 0;
  padding: 4px 12px;
  font-size: 12px;
  color: #ee0a24;
  background: rgba(238, 10, 36, 0.08);
  border-radius: 999px;
}

.order-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
  padding: 12px 14px 14px;
}

.order-item {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 6px;
  align-items: center;
  justify-content: center;
  padding: 14px 8px;
  color: #323233;
  cursor: pointer;
  background: #f7f8fa;
  border: 0;
  border-radius: 12px;
}

.order-item span {
  font-size: 12px;
}
</style>
