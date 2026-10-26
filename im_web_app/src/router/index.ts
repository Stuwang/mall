import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/authStore'
import { routeNames, routes } from './routes'

export const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  const requireAuth = Boolean(to.meta.requireAuth)
  const isPublic = Boolean(to.meta.public)

  if ((requireAuth || !isPublic) && !auth.isLoggedIn) {
    return {
      name: routeNames.login,
      query: { redirect: to.fullPath },
    }
  }

  if (
    (to.name === routeNames.login || to.name === routeNames.register) &&
    auth.isLoggedIn
  ) {
    return { name: routeNames.online }
  }

  return true
})

router.afterEach((to) => {
  const title = typeof to.meta.title === 'string' ? to.meta.title : 'IM'
  document.title = `${title} · IM`
})
