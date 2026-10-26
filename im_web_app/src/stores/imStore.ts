import { defineStore } from 'pinia'
import { onMounted, onUnmounted, ref } from 'vue'
import type { OnlineUser } from '@/api/types'
import { imClient } from '@/im/ImClient'
import type { LocalConv, LocalMsg } from '@/im/types'

export const useImStore = defineStore('im', () => {
  const convs = ref<LocalConv[]>([])
  const online = ref<OnlineUser[]>([])
  const wsConnected = ref(false)
  const tick = ref(0)

  function refresh(): void {
    convs.value = imClient.listConvs()
    online.value = imClient.getOnline()
    wsConnected.value = imClient.isWsConnected()
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

  async function refreshOnline(): Promise<void> {
    await imClient.refreshOnline()
    refresh()
  }

  async function syncAll(): Promise<void> {
    await imClient.syncAll()
    refresh()
  }

  async function syncPeer(peerUserId: number): Promise<void> {
    await imClient.syncPeer(peerUserId)
    refresh()
  }

  function openLocalChat(peerUserId: number, nickname: string): void {
    imClient.openLocalChat(peerUserId, nickname)
    refresh()
  }

  async function sendText(peerUserId: number, text: string, nickname = ''): Promise<void> {
    await imClient.sendText(peerUserId, text, nickname)
    refresh()
  }

  return {
    convs,
    online,
    wsConnected,
    tick,
    refresh,
    bind,
    messagesOf,
    refreshOnline,
    syncAll,
    syncPeer,
    openLocalChat,
    sendText,
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
