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

function openProfile(userId: number): void {
  void router.push({ name: routeNames.profile, params: { userId: String(userId) } })
}
</script>

<template>
  <div class="page">
    <van-nav-bar title="在线用户" />
    <p class="hint muted">点头像看主页，点消息开聊。仅显示当前 WebSocket 在线的用户。</p>
    <van-pull-refresh v-model="refreshing" @refresh="onRefresh">
      <van-empty v-if="imStore.online.length === 0" description="暂时没有其他在线用户" />
      <div v-else class="user-list">
        <div v-for="u in imStore.online" :key="u.user_id" class="user-item">
          <button type="button" class="avatar-hit" @click="openProfile(u.user_id)">
            <UserAvatar
              :user-id="u.user_id"
              :nickname="u.nickname"
              :src="u.avatar"
              :size="48"
            />
          </button>
          <div class="meta">
            <div class="name">{{ u.nickname || `用户${u.user_id}` }}</div>
            <div class="muted">id {{ u.user_id }} · {{ u.conn_count }} 连接</div>
          </div>
          <button
            type="button"
            class="msg-btn"
            @click="startChat(u.user_id, u.nickname, u.avatar)"
          >
            消息
          </button>
        </div>
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
  border-radius: 12px;
  background: #fff;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.04);
}

.avatar-hit {
  display: inline-flex;
  flex-shrink: 0;
  padding: 0;
  border: none;
  border-radius: 50%;
  background: transparent;
  cursor: pointer;
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

.msg-btn {
  flex-shrink: 0;
  border: none;
  border-radius: 16px;
  padding: 6px 14px;
  background: var(--im-primary);
  color: #fff;
  font-size: 14px;
  cursor: pointer;
}

.msg-btn:active {
  opacity: 0.85;
}
</style>
