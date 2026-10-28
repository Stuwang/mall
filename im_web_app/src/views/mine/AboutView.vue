<script setup lang="ts">
import { useRouter } from 'vue-router'

const router = useRouter()
</script>

<template>
  <div class="page">
    <van-nav-bar title="关于" left-arrow fixed placeholder @click-left="router.back()" />
    <div class="body">
      <div class="title">IM</div>
      <div class="muted">即时通讯与社区</div>
      <div class="muted version">版本 1.0.0</div>
    </div>
  </div>
</template>

<style scoped>
.body {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 64px 24px;
  text-align: center;
}

.title {
  font-size: 28px;
  font-weight: 700;
}

.version {
  margin-top: 16px;
}
</style>
