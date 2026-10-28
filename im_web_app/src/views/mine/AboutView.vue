<script setup lang="ts">
import { useRouter } from 'vue-router'

const router = useRouter()
</script>

<template>
  <div class="page">
    <van-nav-bar title="关于" left-arrow fixed placeholder @click-left="router.back()" />
    <div class="body im-fade-up">
      <div class="logo">IM</div>
      <div class="title">即时通讯与社区</div>
      <div class="desc">聊天、帖子、相册，一处相聚。</div>
      <div class="version">版本 1.0.0</div>
    </div>
  </div>
</template>

<style scoped>
.body {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  padding: 72px 28px;
  text-align: center;
}

.logo {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 76px;
  height: 76px;
  margin-bottom: 8px;
  border-radius: 22px;
  background: linear-gradient(145deg, var(--im-primary), var(--im-primary-deep));
  color: #fff;
  font-family: Outfit, 'Noto Sans SC', sans-serif;
  font-size: 28px;
  font-weight: 700;
  letter-spacing: 0.06em;
  box-shadow: 0 18px 36px rgba(15, 143, 134, 0.28);
}

.title {
  font-size: 20px;
  font-weight: 700;
}

.desc {
  color: var(--im-text-secondary);
  font-size: 14px;
}

.version {
  margin-top: 18px;
  padding: 6px 12px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.7);
  border: 1px solid var(--im-border);
  color: var(--im-muted);
  font-size: 12px;
  font-weight: 600;
}
</style>
