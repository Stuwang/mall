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
    <van-nav-bar title="在线" />
    <div class="banner im-fade-up">
      <div class="banner-title">此刻在线</div>
      <div class="banner-desc">点头像看主页，点消息开聊</div>
    </div>
    <van-pull-refresh v-model="refreshing" @refresh="onRefresh">
      <van-empty v-if="imStore.online.length === 0" description="暂时没有其他在线用户" />
      <div v-else class="user-list">
        <div
          v-for="u in imStore.online"
          :key="u.user_id"
          class="user-item"
        >
          <button type="button" class="avatar-hit" @click="openProfile(u.user_id)">
            <span class="online-dot" aria-hidden="true" />
            <UserAvatar
              :user-id="u.user_id"
              :nickname="u.nickname"
              :src="u.avatar"
              :size="52"
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
.banner {
  margin: 12px 16px 4px;
  padding: 16px 18px;
  border-radius: 18px;
  background:
    linear-gradient(135deg, rgba(255, 255, 255, 0.92), rgba(215, 243, 239, 0.75));
  border: 1px solid var(--im-border);
  box-shadow: var(--im-shadow-sm);
}

.banner-title {
  font-size: 16px;
  font-weight: 700;
}

.banner-desc {
  margin-top: 4px;
  font-size: 13px;
  color: var(--im-text-secondary);
}

.user-list {
  padding: 10px 12px 16px;
}

.user-item {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  margin: 0 0 10px;
  padding: 14px;
  border-radius: var(--im-radius);
  background: var(--im-card);
  border: 1px solid var(--im-border);
}

.avatar-hit {
  position: relative;
  display: inline-flex;
  flex-shrink: 0;
  padding: 0;
  border: none;
  border-radius: 50%;
  background: transparent;
  cursor: pointer;
}

.online-dot {
  position: absolute;
  right: 1px;
  bottom: 1px;
  z-index: 1;
  width: 12px;
  height: 12px;
  border: 2px solid #fff;
  border-radius: 50%;
  background: var(--im-success);
  box-shadow: 0 0 0 2px rgba(31, 157, 106, 0.18);
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
  border-radius: 999px;
  padding: 8px 16px;
  background: var(--im-primary-soft);
  color: var(--im-primary-deep);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.15s ease, color 0.15s ease;
}

.msg-btn:active {
  background: var(--im-primary);
  color: #fff;
}
</style>
