<script setup lang="ts">
import { useAuthStore } from '@/stores/authStore'

const auth = useAuthStore()
</script>

<template>
  <div class="page-card">
    <h2>欢迎回来，{{ auth.displayName }}</h2>
    <p class="desc">左侧菜单可管理用户、分类、商品、Banner 与订单；商品/Banner 支持图片上传。</p>
    <el-alert
      type="info"
      :closable="false"
      title="联调提示"
      description="请确保 server 的 env 不为 test（否则走 UserId Header 而非 JWT），并已将账号 is_admin 设为 true。"
      show-icon
    />
  </div>
</template>

<style scoped>
h2 {
  margin: 0 0 8px;
}

.desc {
  color: #6b7280;
  margin: 0 0 16px;
}
</style>
