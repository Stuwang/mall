<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { routeNames } from '@/router/routes'
import { useImLive, useImStore } from '@/stores/imStore'
import { notifyError } from '@/utils/notify'

const router = useRouter()
const imStore = useImStore()
useImLive()

const refreshing = ref(false)

async function onRefresh(): Promise<void> {
  refreshing.value = true
  try {
    await imStore.refreshOnline()
  } catch (error) {
    notifyError(error, '刷新失败')
  } finally {
    refreshing.value = false
  }
}

function startChat(userId: number, nickname: string): void {
  imStore.openLocalChat(userId, nickname || `用户${userId}`)
  void router.push({ name: routeNames.chat, params: { peerId: String(userId) } })
}
</script>

<template>
  <div class="page">
    <van-nav-bar title="在线用户" />
    <p class="hint muted">无需好友，点头像即可开聊。仅显示当前 WebSocket 在线的用户。</p>
    <van-pull-refresh v-model="refreshing" @refresh="onRefresh">
      <van-empty v-if="imStore.online.length === 0" description="暂时没有其他在线用户" />
      <van-cell-group v-else inset>
        <van-cell
          v-for="u in imStore.online"
          :key="u.user_id"
          is-link
          :title="u.nickname || `用户${u.user_id}`"
          :label="`id ${u.user_id} · ${u.conn_count} 连接`"
          @click="startChat(u.user_id, u.nickname)"
        >
          <template #icon>
            <van-icon name="user-circle-o" size="28" style="margin-right: 10px" />
          </template>
        </van-cell>
      </van-cell-group>
    </van-pull-refresh>
  </div>
</template>

<style scoped>
.hint {
  margin: 12px 16px 0;
}
</style>
