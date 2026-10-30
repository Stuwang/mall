<script setup lang="ts">
import { computed } from 'vue'
import { RouterView, useRoute } from 'vue-router'
import GiftEffect from '@/components/GiftEffect.vue'
import { useAuthStore } from '@/stores/authStore'

const route = useRoute()
const auth = useAuthStore()

/** 退出登录后丢掉上一账号的页面缓存。 */
const cacheKey = computed(() => auth.user?.id ?? 0)
const isTab = computed(() => route.matched.some((record) => record.meta.tab))
</script>

<template>
  <!-- 底栏壳单独缓存，避免被详情页挤出。 -->
  <RouterView v-slot="{ Component }">
    <KeepAlive :key="cacheKey">
      <component :is="Component" v-if="isTab" />
    </KeepAlive>
  </RouterView>
  <!-- 主页、聊天、帖子详情等按完整路径各留一份。 -->
  <RouterView v-slot="{ Component, route: child }">
    <KeepAlive :key="cacheKey" :max="12">
      <component :is="Component" v-if="child.meta.keepAlive" :key="child.fullPath" />
    </KeepAlive>
  </RouterView>
  <RouterView v-slot="{ Component, route: child }">
    <component :is="Component" v-if="!isTab && !child.meta.keepAlive" :key="child.fullPath" />
  </RouterView>
  <GiftEffect />
</template>
