<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { routeNames } from '@/router/routes'
import { useImLive, useImStore } from '@/stores/imStore'

const route = useRoute()
const router = useRouter()
const imStore = useImStore()
useImLive()

const tabRoutes = [routeNames.online, routeNames.posts, routeNames.convs, routeNames.mine] as const

const active = computed(() => {
  const idx = tabRoutes.indexOf(route.name as (typeof tabRoutes)[number])
  return idx >= 0 ? idx : 0
})

function onChange(index: number | string): void {
  const target = tabRoutes[Number(index)]
  if (target && route.name !== target) {
    void router.push({ name: target })
  }
}
</script>

<template>
  <div class="page page-with-tab">
    <RouterView v-slot="{ Component }">
      <KeepAlive>
        <component :is="Component" />
      </KeepAlive>
    </RouterView>
    <van-tabbar :model-value="active" fixed placeholder safe-area-inset-bottom @change="onChange">
      <van-tabbar-item icon="friends-o">在线</van-tabbar-item>
      <van-tabbar-item icon="notes-o">帖子</van-tabbar-item>
      <van-tabbar-item icon="chat-o" :badge="imStore.unreadBadge">消息</van-tabbar-item>
      <van-tabbar-item icon="user-o">我的</van-tabbar-item>
    </van-tabbar>
  </div>
</template>
