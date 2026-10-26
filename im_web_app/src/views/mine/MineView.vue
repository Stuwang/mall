<script setup lang="ts">
import { useRouter } from 'vue-router'
import { routeNames } from '@/router/routes'
import { useAuthStore } from '@/stores/authStore'
import { useImLive, useImStore } from '@/stores/imStore'
import { notifySuccess } from '@/utils/notify'

const router = useRouter()
const authStore = useAuthStore()
const imStore = useImStore()
useImLive()

function logout(): void {
  authStore.logout()
  notifySuccess('已退出')
  void router.replace({ name: routeNames.login })
}
</script>

<template>
  <div class="page">
    <van-nav-bar title="我的" />
    <van-cell-group inset>
      <van-cell title="昵称" :value="authStore.displayName" />
      <van-cell title="用户 ID" :value="String(authStore.user?.id ?? '')" />
      <van-cell title="账号" :value="authStore.user?.username || '-'" />
      <van-cell title="长连接" :value="imStore.wsConnected ? '已连接' : '未连接'" />
    </van-cell-group>
    <div class="actions">
      <van-button round block type="danger" plain @click="logout">退出登录</van-button>
    </div>
  </div>
</template>

<style scoped>
.actions {
  margin: 24px 16px;
}
</style>
