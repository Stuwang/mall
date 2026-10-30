<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import UserAvatar from '@/components/UserAvatar.vue'
import { appContext } from '@/app/AppContext'
import { routeNames } from '@/router/routes'
import { useAuthStore } from '@/stores/authStore'
import { useImLive } from '@/stores/imStore'
import { notifyError } from '@/utils/notify'
import { useRestoreScroll } from '@/utils/restoreScroll'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const api = appContext.api
useImLive()

const coinBalance = ref<number | null>(null)

async function loadBalance(): Promise<void> {
  try {
    const reply = await api.coin.balance()
    coinBalance.value = reply.balance
  } catch (error) {
    notifyError(error, '加载金币失败')
  }
}

function openProfile(): void {
  const id = authStore.user?.id
  if (!id) {
    notifyError(null, '未登录')
    return
  }
  void router.push({ name: routeNames.profile, params: { userId: String(id) } })
}

function openCoin(): void {
  void router.push({ name: routeNames.coin })
}

onMounted(() => {
  void loadBalance()
})

useRestoreScroll()

watch(
  () => route.fullPath,
  () => {
    if (route.name === routeNames.mine) {
      void loadBalance()
    }
  },
)

function openSettings(): void {
  void router.push({ name: routeNames.settings })
}

function openAbout(): void {
  void router.push({ name: routeNames.about })
}
</script>

<template>
  <div class="page">
    <van-nav-bar title="我的" />

    <div class="profile-card im-fade-up">
      <div class="glow" aria-hidden="true" />
      <button type="button" class="profile-hit" @click="openProfile">
        <UserAvatar
          :user-id="authStore.user?.id ?? 0"
          :nickname="authStore.displayName"
          :src="authStore.user?.avatar"
          :size="72"
        />
        <div class="profile-meta">
          <div class="name">{{ authStore.displayName }}</div>
          <div class="sub">用户 ID {{ authStore.user?.id ?? '-' }}</div>
          <div class="hint">查看主页 · 相册与帖子</div>
        </div>
        <van-icon name="arrow" class="chevron" />
      </button>
      <button type="button" class="coin-chip" @click="openCoin">
        <van-icon name="gold-coin-o" />
        <span>金币 {{ coinBalance == null ? '…' : coinBalance }}</span>
      </button>
    </div>

    <van-cell-group inset class="menu im-fade-up">
      <van-cell title="设置" is-link @click="openSettings">
        <template #icon>
          <span class="cell-icon settings"><van-icon name="setting-o" /></span>
        </template>
      </van-cell>
      <van-cell title="关于" is-link @click="openAbout">
        <template #icon>
          <span class="cell-icon about"><van-icon name="info-o" /></span>
        </template>
      </van-cell>
    </van-cell-group>
  </div>
</template>

<style scoped>
.profile-card {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  width: calc(100% - 32px);
  margin: 18px 16px 14px;
  padding: 20px 18px 16px;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.55);
  border-radius: 20px;
  background: linear-gradient(
    135deg,
    rgba(15, 143, 134, 0.92),
    rgba(10, 107, 101, 0.88) 48%,
    rgba(45, 90, 120, 0.9)
  );
  color: #fff;
  text-align: left;
  box-shadow: 0 16px 36px rgba(15, 143, 134, 0.28);
}

.profile-hit {
  position: relative;
  display: flex;
  align-items: center;
  gap: 16px;
  width: 100%;
  padding: 0;
  border: none;
  background: transparent;
  color: inherit;
  text-align: left;
  cursor: pointer;
}

.glow {
  position: absolute;
  inset: auto -20% -40% auto;
  width: 180px;
  height: 180px;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(255, 255, 255, 0.28), transparent 68%);
  pointer-events: none;
}

.profile-meta {
  position: relative;
  flex: 1;
  min-width: 0;
}

.name {
  font-size: 20px;
  font-weight: 700;
  letter-spacing: 0.01em;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.sub {
  margin-top: 4px;
  font-size: 12px;
  opacity: 0.82;
}

.hint {
  margin-top: 10px;
  font-size: 12px;
  opacity: 0.72;
}

.chevron {
  position: relative;
  color: rgba(255, 255, 255, 0.75);
  font-size: 16px;
}

.coin-chip {
  position: relative;
  z-index: 1;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  margin: 12px 0 0 88px;
  padding: 6px 12px;
  border: 1px solid rgba(255, 255, 255, 0.28);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.16);
  color: #fff;
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
}

.menu {
  animation-delay: 0.06s;
}

.cell-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  margin-right: 10px;
  border-radius: 8px;
  font-size: 16px;
}

.cell-icon.settings {
  color: var(--im-primary);
  background: var(--im-primary-soft);
}

.cell-icon.about {
  color: var(--im-accent);
  background: rgba(232, 135, 58, 0.14);
}
</style>
