import { defineStore } from 'pinia'
import { computed, onMounted, onUnmounted, ref } from 'vue'
import type { OnlineUser } from '@/api/types'
import { imClient } from '@/im/ImClient'
import type { LocalConv, LocalMsg } from '@/im/types'

export const useImStore = defineStore('im', () => {
  const convs = ref<LocalConv[]>([])
  const online = ref<OnlineUser[]>([])
  const wsConnected = ref(false)
  const totalUnread = ref(0)
  const tick = ref(0)

  function refresh(): void {
    convs.value = imClient.listConvs()
    online.value = imClient.getOnline()
    wsConnected.value = imClient.isWsConnected()
    totalUnread.value = imClient.totalUnread()
    tick.value += 1
  }

  function bind(): () => void {
    refresh()
    return imClient.subscribe(refresh)
  }

  function messagesOf(peerUserId: number): LocalMsg[] {
    void tick.value
    return imClient.getMessages(peerUserId)
  }

  function unreadOf(peerUserId: number): number {
    void tick.value
    return imClient.unreadOf(peerUserId)
  }

  const unreadBadge = computed(() =>
    totalUnread.value > 0 ? String(totalUnread.value > 99 ? '99+' : totalUnread.value) : '',
  )

  function setActivePeer(peerUserId: number): void {
    imClient.setActivePeer(peerUserId)
    refresh()
  }

  async function refreshOnline(): Promise<void> {
    await imClient.refreshOnline()
    refresh()
  }

  async function syncAll(fullList = false): Promise<void> {
    await imClient.syncAll(fullList)
    refresh()
  }

  async function syncPeer(peerUserId: number): Promise<void> {
    await imClient.syncPeer(peerUserId)
    refresh()
  }

  function openLocalChat(peerUserId: number, nickname: string, avatar = ''): void {
    imClient.openLocalChat(peerUserId, nickname, avatar)
    refresh()
  }

  async function sendText(
    peerUserId: number,
    text: string,
    nickname = '',
    avatar = '',
  ): Promise<void> {
    await imClient.sendText(peerUserId, text, nickname, avatar)
    refresh()
  }

  async function sendImage(
    peerUserId: number,
    file: File,
    nickname = '',
    avatar = '',
  ): Promise<void> {
    await imClient.sendImage(peerUserId, file, nickname, avatar)
    refresh()
  }

  return {
    convs,
    online,
    wsConnected,
    totalUnread,
    unreadBadge,
    tick,
    refresh,
    bind,
    messagesOf,
    unreadOf,
    setActivePeer,
    refreshOnline,
    syncAll,
    syncPeer,
    openLocalChat,
    sendText,
    sendImage,
  }
})

/** 在组件里订阅 ImClient 变更 */
export function useImLive(): void {
  const store = useImStore()
  let unsub: (() => void) | undefined
  onMounted(() => {
    unsub = store.bind()
  })
  onUnmounted(() => {
    unsub?.()
  })
}
