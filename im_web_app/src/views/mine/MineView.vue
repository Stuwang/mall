<script setup lang="ts">
import { useRouter } from 'vue-router'
import UserAvatar from '@/components/UserAvatar.vue'
import { routeNames } from '@/router/routes'
import { useAuthStore } from '@/stores/authStore'
import { useImLive } from '@/stores/imStore'
import { notifyError } from '@/utils/notify'

const router = useRouter()
const authStore = useAuthStore()
useImLive()

function openProfile(): void {
  const id = authStore.user?.id
  if (!id) {
    notifyError(null, '未登录')
    return
  }
  void router.push({ name: routeNames.profile, params: { userId: String(id) } })
}

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

    <button type="button" class="profile-card im-fade-up" @click="openProfile">
      <div class="glow" aria-hidden="true" />
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
  align-items: center;
  gap: 16px;
  width: calc(100% - 32px);
  margin: 18px 16px 14px;
  padding: 20px 18px;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.55);
  border-radius: 20px;
  background:
    linear-gradient(135deg, rgba(15, 143, 134, 0.92), rgba(10, 107, 101, 0.88) 48%, rgba(45, 90, 120, 0.9));
  color: #fff;
  text-align: left;
  cursor: pointer;
  box-shadow: 0 16px 36px rgba(15, 143, 134, 0.28);
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
