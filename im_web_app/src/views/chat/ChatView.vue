<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ContentTypeImage, ContentTypeText } from '@/api/types'
import ImageViewer from '@/components/ImageViewer.vue'
import UserAvatar from '@/components/UserAvatar.vue'
import { routeNames } from '@/router/routes'
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
const listInnerRef = ref<HTMLElement | null>(null)
const fileInput = ref<HTMLInputElement | null>(null)
const previewUrl = ref('')
const previewVisible = ref(false)
/** 用户是否贴在底部；图片加载撑高后据此继续贴底 */
const stickToBottom = ref(true)
let listResizeObserver: ResizeObserver | null = null

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

function isNearBottom(threshold = 80): boolean {
  const el = listRef.value
  if (!el) {
    return true
  }
  return el.scrollHeight - el.scrollTop - el.clientHeight <= threshold
}

function applyScrollBottom(): void {
  const el = listRef.value
  if (!el) {
    return
  }
  el.scrollTop = el.scrollHeight
  stickToBottom.value = true
}

async function scrollBottom(): Promise<void> {
  stickToBottom.value = true
  await nextTick()
  applyScrollBottom()
  // 再等一帧，避免布局未完成
  requestAnimationFrame(() => applyScrollBottom())
}

function onListScroll(): void {
  stickToBottom.value = isNearBottom()
}

function onImageLoad(): void {
  if (stickToBottom.value || isNearBottom()) {
    applyScrollBottom()
  }
}

function bindListResize(): void {
  listResizeObserver?.disconnect()
  listResizeObserver = null
  const inner = listInnerRef.value
  if (!inner || typeof ResizeObserver === 'undefined') {
    return
  }
  listResizeObserver = new ResizeObserver(() => {
    if (stickToBottom.value) {
      applyScrollBottom()
    }
  })
  listResizeObserver.observe(inner)
}

onMounted(() => {
  imStore.setActivePeer(peerId.value)
  bindListResize()
  void imStore.syncPeer(peerId.value).then(() => scrollBottom())
})

onUnmounted(() => {
  listResizeObserver?.disconnect()
  listResizeObserver = null
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

watch(listInnerRef, () => {
  bindListResize()
})

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

function pickImage(): void {
  if (sending.value) {
    return
  }
  fileInput.value?.click()
}

async function onImageSelected(ev: Event): Promise<void> {
  const input = ev.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file || sending.value) {
    return
  }
  if (!file.type.startsWith('image/')) {
    notifyError(null, '请选择图片文件')
    return
  }
  sending.value = true
  try {
    await imStore.sendImage(peerId.value, file, peerNickname.value, peerAvatar.value)
    await scrollBottom()
  } catch (error) {
    notifyError(error, '发送图片失败')
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

function isImage(m: { content_type?: number }): boolean {
  return m.content_type === ContentTypeImage
}

function openPreview(url: string): void {
  if (!url) {
    return
  }
  previewUrl.value = url
  previewVisible.value = true
}

function isText(m: { content_type?: number }): boolean {
  return !m.content_type || m.content_type === ContentTypeText
}

function openPeerProfile(): void {
  if (!peerId.value) {
    return
  }
  void router.push({ name: routeNames.profile, params: { userId: String(peerId.value) } })
}
</script>

<template>
  <div class="page chat-page">
    <van-nav-bar
      :title="peerNickname"
      left-arrow
      fixed
      placeholder
      @click-left="router.back()"
      @click-title="openPeerProfile"
    />
    <div ref="listRef" class="msg-list" @scroll.passive="onListScroll">
      <div ref="listInnerRef" class="msg-list-inner">
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
          <div class="bubble" :class="{ pending: m.pending, failed: m.failed, image: isImage(m) }">
            <img
              v-if="isImage(m) && m.content?.image_url"
              class="msg-image"
              :src="m.content.image_url"
              alt="图片"
              @load="onImageLoad"
              @click="openPreview(m.content.image_url)"
            />
            <div v-else-if="isText(m)" class="text">{{ m.content?.text }}</div>
            <div v-else class="text muted">[不支持的消息]</div>
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
    </div>
    <div class="composer">
      <button type="button" class="img-btn" :disabled="sending" @click="pickImage">
        <van-icon name="photo-o" />
      </button>
      <input
        ref="fileInput"
        class="hidden-input"
        type="file"
        accept="image/jpeg,image/png,image/gif,image/webp"
        @change="onImageSelected"
      />
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
    <ImageViewer v-model:show="previewVisible" :src="previewUrl" />
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

.bubble.image {
  padding: 6px;
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

.msg-image {
  display: block;
  max-width: 220px;
  max-height: 280px;
  border-radius: 8px;
  object-fit: cover;
  cursor: zoom-in;
}

.meta {
  margin-top: 4px;
  font-size: 11px;
  color: var(--im-muted);
  text-align: right;
}

.bubble.image .meta {
  padding: 0 6px 2px;
}

.composer {
  display: flex;
  gap: 8px;
  align-items: flex-end;
  padding: 8px 12px calc(8px + var(--im-safe-bottom));
  background: #fff;
  border-top: 1px solid #ebedf0;
}

.img-btn {
  flex-shrink: 0;
  width: 36px;
  height: 36px;
  border: none;
  border-radius: 8px;
  background: #f7f8fa;
  color: #323233;
  font-size: 20px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

.img-btn:disabled {
  opacity: 0.5;
}

.hidden-input {
  display: none;
}

.composer :deep(.van-field) {
  flex: 1;
  background: #f7f8fa;
  border-radius: 8px;
  padding: 6px 10px;
}
</style>
