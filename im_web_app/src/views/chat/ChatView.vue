<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import UserAvatar from '@/components/UserAvatar.vue'
import { useAuthStore } from '@/stores/authStore'
import { useImLive, useImStore } from '@/stores/imStore'
import { formatTime } from '@/utils/conv'
import { notifyError } from '@/utils/notify'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const imStore = useImStore()
useImLive()

const peerId = computed(() => Number(route.params.peerId))
const draft = ref('')
const sending = ref(false)
const listRef = ref<HTMLElement | null>(null)

const peerNickname = computed(() => {
  const conv = imStore.convs.find((c) => c.peer?.user_id === peerId.value)
  if (conv?.peer?.nickname) {
    return conv.peer.nickname
  }
  const online = imStore.online.find((u) => u.user_id === peerId.value)
  return online?.nickname || `用户${peerId.value}`
})

const peerAvatar = computed(() => {
  const conv = imStore.convs.find((c) => c.peer?.user_id === peerId.value)
  if (conv?.peer?.avatar) {
    return conv.peer.avatar
  }
  return imStore.online.find((u) => u.user_id === peerId.value)?.avatar || ''
})

const selfNickname = computed(
  () => authStore.user?.nickname || authStore.user?.username || '我',
)

const selfAvatar = computed(() => authStore.user?.avatar || '')

const selfId = computed(() => authStore.user?.id ?? 0)

const messages = computed(() => imStore.messagesOf(peerId.value))

async function scrollBottom(): Promise<void> {
  await nextTick()
  const el = listRef.value
  if (el) {
    el.scrollTop = el.scrollHeight
  }
}

onMounted(() => {
  imStore.setActivePeer(peerId.value)
  void imStore.syncPeer(peerId.value).then(() => scrollBottom())
})

onUnmounted(() => {
  imStore.setActivePeer(0)
})

watch(peerId, (id) => {
  imStore.setActivePeer(id)
  void imStore.syncPeer(id).then(() => scrollBottom())
})

watch(
  () => messages.value.length,
  () => {
    void scrollBottom()
  },
)

async function onSend(): Promise<void> {
  const text = draft.value
  if (!text.trim() || sending.value) {
    return
  }
  sending.value = true
  try {
    await imStore.sendText(peerId.value, text, peerNickname.value, peerAvatar.value)
    draft.value = ''
    await scrollBottom()
  } catch (error) {
    notifyError(error, '发送失败')
  } finally {
    sending.value = false
  }
}

function isMine(senderId: number): boolean {
  return senderId === selfId.value
}

function senderName(senderId: number): string {
  if (isMine(senderId)) {
    return selfNickname.value
  }
  return peerNickname.value
}
</script>

<template>
  <div class="page chat-page">
    <van-nav-bar :title="peerNickname" left-arrow fixed placeholder @click-left="router.back()" />
    <div ref="listRef" class="msg-list">
      <div
        v-for="(m, idx) in messages"
        :key="m.msg_id || m.client_msg_id || idx"
        class="msg-row"
        :class="{ mine: isMine(m.sender_id) }"
      >
        <UserAvatar
          v-if="!isMine(m.sender_id)"
          class="avatar"
          :user-id="m.sender_id || peerId"
          :nickname="senderName(m.sender_id)"
          :src="peerAvatar"
          :size="40"
        />
        <div class="bubble" :class="{ pending: m.pending, failed: m.failed }">
          <div class="text">{{ m.content?.text }}</div>
          <div class="meta">
            <span v-if="m.pending">发送中</span>
            <span v-else-if="m.failed">失败</span>
            <span v-else>{{ formatTime(m.ct) }}</span>
          </div>
        </div>
        <UserAvatar
          v-if="isMine(m.sender_id)"
          class="avatar"
          :user-id="selfId"
          :nickname="selfNickname"
          :src="selfAvatar"
          :size="40"
        />
      </div>
    </div>
    <div class="composer">
      <van-field
        v-model="draft"
        rows="1"
        autosize
        type="textarea"
        placeholder="输入消息"
        @keydown.enter.exact.prevent="onSend"
      />
      <van-button type="primary" size="small" :loading="sending" @click="onSend">发送</van-button>
    </div>
  </div>
</template>

<style scoped>
.chat-page {
  display: flex;
  flex-direction: column;
  height: 100vh;
  height: 100dvh;
}

.msg-list {
  flex: 1;
  overflow-y: auto;
  padding: 12px 12px 8px;
  background: linear-gradient(180deg, #eef3f8 0%, var(--im-bg) 120px);
}

.msg-row {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin-bottom: 14px;
}

.msg-row.mine {
  flex-direction: row;
  justify-content: flex-end;
}

.avatar {
  margin-top: 2px;
}

.bubble {
  max-width: calc(100% - 96px);
  padding: 10px 12px;
  border-radius: 12px;
  background: var(--im-bubble-peer);
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.06);
}

.msg-row:not(.mine) .bubble {
  border-top-left-radius: 4px;
}

.msg-row.mine .bubble {
  background: var(--im-bubble-me);
  border-top-right-radius: 4px;
}

.bubble.pending {
  opacity: 0.7;
}

.bubble.failed {
  outline: 1px solid #ee0a24;
}

.text {
  white-space: pre-wrap;
  word-break: break-word;
  font-size: 15px;
}

.meta {
  margin-top: 4px;
  font-size: 11px;
  color: var(--im-muted);
  text-align: right;
}

.composer {
  display: flex;
  gap: 8px;
  align-items: flex-end;
  padding: 8px 12px calc(8px + var(--im-safe-bottom));
  background: #fff;
  border-top: 1px solid #ebedf0;
}

.composer :deep(.van-field) {
  flex: 1;
  background: #f7f8fa;
  border-radius: 8px;
  padding: 6px 10px;
}
</style>
