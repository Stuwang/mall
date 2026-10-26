<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { routeNames } from '@/router/routes'
import { useImLive, useImStore } from '@/stores/imStore'
import { formatTime } from '@/utils/conv'
import { notifyError } from '@/utils/notify'

const router = useRouter()
const imStore = useImStore()
useImLive()

const refreshing = ref(false)

async function onRefresh(): Promise<void> {
  refreshing.value = true
  try {
    await imStore.syncAll()
  } catch (error) {
    notifyError(error, '同步失败')
  } finally {
    refreshing.value = false
  }
}

function openChat(peerId: number): void {
  void router.push({ name: routeNames.chat, params: { peerId: String(peerId) } })
}
</script>

<template>
  <div class="page">
    <van-nav-bar title="消息">
      <template #right>
        <span class="ws-dot" :class="{ on: imStore.wsConnected }" />
      </template>
    </van-nav-bar>
    <van-pull-refresh v-model="refreshing" @refresh="onRefresh">
      <van-empty v-if="imStore.convs.length === 0" description="暂无会话，去「在线」找人聊" />
      <van-cell-group v-else inset>
        <van-cell
          v-for="c in imStore.convs"
          :key="c.conv_id"
          is-link
          :title="c.peer?.nickname || `用户${c.peer?.user_id}`"
          :label="c.last_msg?.preview || (c.local_only ? '草稿会话' : '')"
          :value="formatTime(c.sort_ut)"
          @click="openChat(c.peer.user_id)"
        />
      </van-cell-group>
    </van-pull-refresh>
  </div>
</template>

<style scoped>
.ws-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #c8c9cc;
  margin-right: 8px;
}

.ws-dot.on {
  background: #07c160;
}
</style>
