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

    <button type="button" class="profile-card" @click="openProfile">
      <UserAvatar
        :user-id="authStore.user?.id ?? 0"
        :nickname="authStore.displayName"
        :src="authStore.user?.avatar"
        :size="64"
      />
      <div class="profile-meta">
        <div class="name">{{ authStore.displayName }}</div>
        <div class="muted">用户 ID {{ authStore.user?.id ?? '-' }}</div>
      </div>
      <van-icon name="arrow" class="chevron" />
    </button>

    <van-cell-group inset>
      <van-cell title="设置" is-link @click="openSettings" />
      <van-cell title="关于" is-link @click="openAbout" />
    </van-cell-group>
  </div>
</template>

<style scoped>
.profile-card {
  display: flex;
  align-items: center;
  gap: 14px;
  width: calc(100% - 32px);
  margin: 16px 16px 12px;
  padding: 16px;
  border: none;
  border-radius: 12px;
  background: #fff;
  text-align: left;
  cursor: pointer;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.04);
}

.profile-meta {
  flex: 1;
  min-width: 0;
}

.name {
  font-size: 18px;
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.chevron {
  color: #c8c9cc;
  font-size: 16px;
}
</style>
