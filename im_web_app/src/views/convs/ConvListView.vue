<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import UserAvatar from '@/components/UserAvatar.vue'
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

function unreadBadge(peerId: number): string | undefined {
  const n = imStore.unreadOf(peerId)
  if (n <= 0) {
    return undefined
  }
  return n > 99 ? '99+' : String(n)
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
      <div v-else class="conv-list">
        <button
          v-for="c in imStore.convs"
          :key="c.conv_id"
          type="button"
          class="conv-item"
          @click="openChat(c.peer.user_id)"
        >
          <UserAvatar
            :user-id="c.peer.user_id"
            :nickname="c.peer?.nickname"
            :size="48"
          />
          <div class="body">
            <div class="top">
              <span class="name">{{ c.peer?.nickname || `用户${c.peer?.user_id}` }}</span>
              <span class="time muted">{{ formatTime(c.sort_ut) }}</span>
            </div>
            <div class="bottom">
              <span class="preview muted">
                {{ c.last_msg?.preview || (c.local_only ? '草稿会话' : '') }}
              </span>
              <van-badge v-if="unreadBadge(c.peer.user_id)" :content="unreadBadge(c.peer.user_id)" />
            </div>
          </div>
        </button>
      </div>
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

.conv-list {
  padding: 8px 12px 16px;
}

.conv-item {
  display: flex;
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

.conv-item:active {
  background: #f2f3f5;
}

.body {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 6px;
}

.top,
.bottom {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.name {
  font-size: 16px;
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.preview {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.time {
  flex-shrink: 0;
}
</style>
