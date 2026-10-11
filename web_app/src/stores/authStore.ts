import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { ApiError } from '@/api/errors/ApiError'
import type { UserInfo } from '@/api/types'
import { appContext } from '@/app/AppContext'

/**
 * Pinia 适配层：对外暴露响应式登录态，内部委托 AuthSession + AuthApiService
 */
export const useAuthStore = defineStore('auth', () => {
  const session = appContext.session
  const api = appContext.api.auth

  const accessToken = ref<string | null>(session.getAccessToken())
  const user = ref<UserInfo | null>(session.getUser())

  const isLoggedIn = computed(() => Boolean(accessToken.value))
  const displayName = computed(
    () => user.value?.nickname || user.value?.username || '用户',
  )

  function syncFromSession(): void {
    accessToken.value = session.getAccessToken()
    user.value = session.getUser()
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
  }

  async function register(
    username: string,
    password: string,
    nickname?: string,
  ): Promise<void> {
    await api.register({ username, password, nickname })
    // 注册不返回 token，注册成功后自动登录
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

  function logout(): void {
    session.clear()
    syncFromSession()
  }

  /** 供 HttpClient 未授权回调使用 */
  function forceLogout(): void {
    logout()
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
    syncFromSession,
  }
})
