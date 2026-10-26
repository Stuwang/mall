<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import UserAvatar from '@/components/UserAvatar.vue'
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

function startChat(userId: number, nickname: string, avatar?: string): void {
  imStore.openLocalChat(userId, nickname || `用户${userId}`, avatar || '')
  void router.push({ name: routeNames.chat, params: { peerId: String(userId) } })
}
</script>

<template>
  <div class="page">
    <van-nav-bar title="在线用户" />
    <p class="hint muted">无需好友，点头像即可开聊。仅显示当前 WebSocket 在线的用户。</p>
    <van-pull-refresh v-model="refreshing" @refresh="onRefresh">
      <van-empty v-if="imStore.online.length === 0" description="暂时没有其他在线用户" />
      <div v-else class="user-list">
        <button
          v-for="u in imStore.online"
          :key="u.user_id"
          type="button"
          class="user-item"
          @click="startChat(u.user_id, u.nickname, u.avatar)"
        >
          <UserAvatar
            :user-id="u.user_id"
            :nickname="u.nickname"
            :src="u.avatar"
            :size="48"
          />
          <div class="meta">
            <div class="name">{{ u.nickname || `用户${u.user_id}` }}</div>
            <div class="muted">id {{ u.user_id }} · {{ u.conn_count }} 连接</div>
          </div>
          <van-icon name="chat-o" class="go" />
        </button>
      </div>
    </van-pull-refresh>
  </div>
</template>

<style scoped>
.hint {
  margin: 12px 16px 0;
}

.user-list {
  padding: 8px 12px 16px;
}

.user-item {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  margin: 0 0 8px;
  padding: 12px;
  border: none;
  border-radius: 12px;
  background: #fff;
  text-align: left;
  cursor: pointer;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.04);
}

.user-item:active {
  background: #f2f3f5;
}

.meta {
  flex: 1;
  min-width: 0;
}

.name {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.go {
  color: var(--im-primary);
  font-size: 20px;
}
</style>
