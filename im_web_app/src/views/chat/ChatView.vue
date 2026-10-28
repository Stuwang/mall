<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ContentTypeGift, ContentTypeImage, ContentTypeText, type Gift } from '@/api/types'
import { appContext } from '@/app/AppContext'
import ImageViewer from '@/components/ImageViewer.vue'
import UserAvatar from '@/components/UserAvatar.vue'
import { routeNames } from '@/router/routes'
import { useAuthStore } from '@/stores/authStore'
import { useImLive, useImStore } from '@/stores/imStore'
import { isGiftIconUrl } from '@/utils/coin'
import { formatTime, newClientMsgId } from '@/utils/conv'
import { notifyError, notifyFail } from '@/utils/notify'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const imStore = useImStore()
const api = appContext.api
useImLive()

const peerId = computed(() => Number(route.params.peerId))
const draft = ref('')
const sending = ref(false)
const listRef = ref<HTMLElement | null>(null)
const listInnerRef = ref<HTMLElement | null>(null)
const fileInput = ref<HTMLInputElement | null>(null)
const previewUrl = ref('')
const previewVisible = ref(false)
const giftOpen = ref(false)
const gifts = ref<Gift[]>([])
const giftsLoading = ref(false)
const giftBalance = ref(0)
const giftBalanceReady = ref(false)
const sendingGift = ref(false)
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

const selfNickname = computed(() => authStore.user?.nickname || authStore.user?.username || '我')

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

function isGift(m: { content_type?: number }): boolean {
  return m.content_type === ContentTypeGift
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

async function openGifts(): Promise<void> {
  if (sending.value || sendingGift.value) {
    return
  }
  giftOpen.value = true
  giftsLoading.value = true
  try {
    const [giftReply, balReply] = await Promise.all([api.coin.listGifts(), api.coin.balance()])
    gifts.value = giftReply.list || []
    giftBalance.value = balReply.balance
    giftBalanceReady.value = true
  } catch (error) {
    giftBalanceReady.value = false
    notifyError(error, '加载礼物失败')
  } finally {
    giftsLoading.value = false
  }
}

async function onSendGift(gift: Gift): Promise<void> {
  if (sendingGift.value || giftsLoading.value) {
    return
  }
  if (!peerId.value) {
    notifyFail('无法送礼')
    return
  }
  if (!giftBalanceReady.value) {
    notifyFail('金币余额未加载')
    return
  }
  if (gift.price > giftBalance.value) {
    notifyFail('金币不足')
    return
  }
  sendingGift.value = true
  try {
    const reply = await api.coin.sendGift({
      peer_user_id: peerId.value,
      gift_id: gift.id,
      client_msg_id: newClientMsgId(),
    })
    giftBalance.value = reply.balance
    imStore.deliverServerMsg(reply.msg)
    giftOpen.value = false
    await scrollBottom()
  } catch (error) {
    notifyError(error, '送礼失败')
  } finally {
    sendingGift.value = false
  }
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
          <div
            class="bubble"
            :class="{ pending: m.pending, failed: m.failed, image: isImage(m), gift: isGift(m) }"
          >
            <img
              v-if="isImage(m) && m.content?.image_url"
              class="msg-image"
              :src="m.content.image_url"
              alt="图片"
              @load="onImageLoad"
              @click="openPreview(m.content.image_url)"
            />
            <div v-else-if="isGift(m)" class="gift-msg">
              <img
                v-if="isGiftIconUrl(m.content?.gift?.icon)"
                class="gift-msg-icon"
                :src="m.content?.gift?.icon"
                alt=""
              />
              <span v-else class="gift-msg-emoji">{{ m.content?.gift?.icon || '🎁' }}</span>
              <div class="gift-msg-body">
                <div class="gift-msg-name">{{ m.content?.gift?.name || '礼物' }}</div>
                <div class="gift-msg-label">送出了礼物</div>
              </div>
            </div>
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
      <button type="button" class="img-btn" :disabled="sending || sendingGift" @click="openGifts">
        <van-icon name="gift-o" />
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
    <van-popup v-model:show="giftOpen" position="bottom" round safe-area-inset-bottom>
      <div class="gift-sheet">
        <div class="gift-head">
          <span>送礼物</span>
          <span class="gift-balance">金币 {{ giftBalanceReady ? giftBalance : '…' }}</span>
        </div>
        <div v-if="giftsLoading && gifts.length === 0" class="gift-loading muted">加载中…</div>
        <van-empty v-else-if="giftBalanceReady && gifts.length === 0" description="暂无礼物" />
        <div v-else class="gift-grid">
          <button
            v-for="gift in gifts"
            :key="gift.id"
            type="button"
            class="gift-cell"
            :disabled="sendingGift || giftsLoading || !giftBalanceReady"
            @click="onSendGift(gift)"
          >
            <img v-if="isGiftIconUrl(gift.icon)" class="gift-icon" :src="gift.icon" alt="" />
            <span v-else class="gift-emoji">{{ gift.icon || '🎁' }}</span>
            <span class="gift-name">{{ gift.name }}</span>
            <span class="gift-price">{{ gift.price }} 金币</span>
          </button>
        </div>
      </div>
    </van-popup>
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
  background:
    radial-gradient(700px 280px at 10% 0%, rgba(15, 143, 134, 0.1), transparent 60%),
    linear-gradient(180deg, rgba(245, 247, 250, 0.9) 0%, var(--im-bg) 160px);
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
  border-radius: 16px;
  background: var(--im-bubble-peer);
  border: 1px solid var(--im-border);
  box-shadow: var(--im-shadow-sm);
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
  padding: 10px 12px calc(10px + var(--im-safe-bottom));
  background: rgba(255, 255, 255, 0.92);
  backdrop-filter: blur(14px);
  -webkit-backdrop-filter: blur(14px);
  border-top: 1px solid var(--im-border);
}

.img-btn {
  flex-shrink: 0;
  width: 36px;
  height: 36px;
  border: none;
  border-radius: 10px;
  background: var(--im-primary-soft);
  color: var(--im-primary-deep);
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
  background: #f3f5f8;
  border-radius: 12px;
  padding: 6px 10px;
}

.gift-msg {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 148px;
}

.gift-msg-emoji {
  font-size: 32px;
  line-height: 1;
}

.gift-msg-icon {
  width: 40px;
  height: 40px;
  object-fit: contain;
}

.gift-msg-name {
  font-size: 15px;
  font-weight: 700;
}

.gift-msg-label {
  margin-top: 2px;
  font-size: 12px;
  color: var(--im-text-secondary);
}

.gift-sheet {
  padding: 16px 16px calc(16px + var(--im-safe-bottom));
  max-height: 70vh;
  overflow: auto;
}

.gift-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
  font-size: 16px;
  font-weight: 700;
}

.gift-balance {
  font-size: 13px;
  font-weight: 700;
  color: var(--im-accent);
}

.gift-loading {
  padding: 28px 0;
  text-align: center;
}

.gift-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 8px;
}

.gift-cell {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  min-width: 0;
  padding: 10px 4px;
  border: 1px solid var(--im-border);
  border-radius: 12px;
  background: #fff;
  cursor: pointer;
}

.gift-cell:disabled {
  opacity: 0.55;
}

.gift-emoji {
  font-size: 28px;
  line-height: 1.1;
}

.gift-icon {
  width: 36px;
  height: 36px;
  object-fit: contain;
}

.gift-name {
  max-width: 100%;
  overflow: hidden;
  font-size: 12px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.gift-price {
  font-size: 11px;
  font-weight: 700;
  color: var(--im-accent);
}
</style>
