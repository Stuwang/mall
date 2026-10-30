<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import UserAvatar from '@/components/UserAvatar.vue'
import { routeNames } from '@/router/routes'
import { useImLive, useImStore } from '@/stores/imStore'
import { formatTime } from '@/utils/conv'
import { notifyError } from '@/utils/notify'
import { useRestoreScroll } from '@/utils/restoreScroll'

const router = useRouter()
const imStore = useImStore()
useImLive()

const refreshing = ref(false)

async function onRefresh(): Promise<void> {
  refreshing.value = true
  try {
    await Promise.all([imStore.syncAll(true), imStore.refreshOnline()])
  } catch (error) {
    notifyError(error, '同步失败')
  } finally {
    refreshing.value = false
  }
}

function previewText(c: (typeof imStore.convs)[number]): string {
  if (c.last_msg?.preview) {
    return c.last_msg.preview
  }
  if (c.local_only) {
    return '草稿会话'
  }
  return ''
}

function openChat(peerId: number): void {
  void router.push({ name: routeNames.chat, params: { peerId: String(peerId) } })
}

useRestoreScroll()

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
        <span class="ws-pill" :class="{ on: imStore.wsConnected }">
          <span class="ws-dot" />
          {{ imStore.wsConnected ? '已连接' : '未连接' }}
        </span>
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
            :src="c.peer?.avatar"
            :size="52"
          />
          <div class="body">
            <div class="top">
              <span class="name">{{ c.peer?.nickname || `用户${c.peer?.user_id}` }}</span>
              <span class="time muted">{{ formatTime(c.sort_ut) }}</span>
            </div>
            <div class="bottom">
              <span class="preview muted">{{ previewText(c) }}</span>
              <van-badge v-if="unreadBadge(c.peer.user_id)" :content="unreadBadge(c.peer.user_id)" />
            </div>
          </div>
        </button>
      </div>
    </van-pull-refresh>
  </div>
</template>

<style scoped>
.ws-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  margin-right: 4px;
  padding: 4px 10px;
  border-radius: 999px;
  background: rgba(139, 149, 165, 0.12);
  color: var(--im-muted);
  font-size: 11px;
  font-weight: 600;
}

.ws-pill.on {
  background: rgba(31, 157, 106, 0.12);
  color: var(--im-success);
}

.ws-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: currentColor;
}

.conv-list {
  padding: 10px 12px 16px;
}

.conv-item {
  display: flex;
  gap: 12px;
  width: 100%;
  margin: 0 0 10px;
  padding: 14px;
  border: 1px solid var(--im-border);
  border-radius: var(--im-radius);
  background: var(--im-card);
  text-align: left;
  cursor: pointer;
}

.conv-item:active {
  background: #f6f8fa;
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
