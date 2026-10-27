import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { ApiError } from '@/api/errors/ApiError'
import type { UserInfo } from '@/api/types'
import { appContext } from '@/app/AppContext'
import { imClient } from '@/im/ImClient'

const TOKEN_REFRESH_INTERVAL_MS = 60 * 60 * 1000

export const useAuthStore = defineStore('auth', () => {
  const session = appContext.session
  const api = appContext.api.auth

  const accessToken = ref<string | null>(session.getAccessToken())
  const user = ref<UserInfo | null>(session.getUser())
  let refreshTimer: number | null = null
  let refreshing = false

  const isLoggedIn = computed(() => Boolean(accessToken.value))
  const displayName = computed(
    () => user.value?.nickname || user.value?.username || '用户',
  )

  function syncFromSession(): void {
    accessToken.value = session.getAccessToken()
    user.value = session.getUser()
  }

  function startIm(): void {
    const u = session.getUser()
    const token = session.getAccessToken()
    if (u && token) {
      imClient.start(u.id, token)
    }
  }

  async function login(username: string, password: string): Promise<void> {
    const reply = await api.login({ username, password })
    if (!reply.access_token) {
      throw new ApiError(1, '登录未返回 access_token')
    }
    if (!reply.user) {
      throw new ApiError(1, '登录未返回用户信息')
    }
    session.applyLogin(reply.access_token, reply.user)
    syncFromSession()
    startIm()
    startTokenRefresh(false)
  }

  async function register(
    username: string,
    password: string,
    nickname?: string,
  ): Promise<void> {
    await api.register({ username, password, nickname })
    await login(username, password)
  }

  async function refreshProfile(): Promise<void> {
    if (!session.isLoggedIn()) {
      return
    }
    const reply = await api.getUserInfo({})
    if (!reply.user) {
      logout()
      throw new ApiError(1, '用户信息为空')
    }
    session.updateUser(reply.user)
    syncFromSession()
  }

  async function updateProfile(payload: {
    nickname?: string
    sex?: number
    avatar?: string
  }): Promise<void> {
    const reply = await api.updateUserInfo(payload)
    if (!reply.user) {
      throw new ApiError(1, '更新未返回用户信息')
    }
    session.updateUser(reply.user)
    syncFromSession()
  }

  async function uploadAvatar(file: File): Promise<void> {
    const uploaded = await api.uploadMedia(file)
    if (!uploaded?.uri) {
      throw new ApiError(1, '上传未返回文件地址')
    }
    await updateProfile({
      nickname: user.value?.nickname || user.value?.username || '',
      sex: user.value?.sex ?? 0,
      avatar: uploaded.uri,
    })
  }

  function logout(): void {
    stopTokenRefresh()
    imClient.stop()
    session.clear()
    syncFromSession()
  }

  function forceLogout(): void {
    logout()
  }

  function stopTokenRefresh(): void {
    if (refreshTimer != null) {
      window.clearInterval(refreshTimer)
      refreshTimer = null
    }
  }

  async function refreshAccessToken(): Promise<void> {
    if (!session.isLoggedIn() || refreshing) {
      return
    }
    refreshing = true
    try {
      const reply = await api.refreshToken()
      if (!session.isLoggedIn()) {
        return
      }
      if (!reply.access_token) {
        return
      }
      session.replaceAccessToken(reply.access_token)
      syncFromSession()
      imClient.replaceToken(reply.access_token)
    } catch (err) {
      if (err instanceof ApiError && (err.isUnauthorized || err.isForbidden || err.isBanned)) {
        return
      }
    } finally {
      refreshing = false
    }
  }

  /** immediate：恢复本地登录态时立刻续一次，避免打开页面时 token 已临近过期。 */
  function startTokenRefresh(immediate: boolean): void {
    stopTokenRefresh()
    if (!session.isLoggedIn()) {
      return
    }
    refreshTimer = window.setInterval(() => {
      void refreshAccessToken()
    }, TOKEN_REFRESH_INTERVAL_MS)
    if (immediate) {
      void refreshAccessToken()
    }
  }

  function bootstrap(): void {
    if (session.isLoggedIn()) {
      startIm()
      startTokenRefresh(true)
    }
  }

  return {
    accessToken,
    user,
    isLoggedIn,
    displayName,
    login,
    register,
    logout,
    forceLogout,
    refreshProfile,
    updateProfile,
    uploadAvatar,
    syncFromSession,
    bootstrap,
  }
})
