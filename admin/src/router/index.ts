import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/authStore'
import { routeNames, routes } from './routes'

export const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  const isPublic = Boolean(to.meta.public)

  if (!isPublic && !auth.isLoggedIn) {
    return { name: routeNames.login, query: { redirect: to.fullPath } }
  }

  if (!isPublic && auth.isLoggedIn && !auth.isAdmin) {
    auth.logout()
    return { name: routeNames.login }
  }

  if (to.name === routeNames.login && auth.isLoggedIn && auth.isAdmin) {
    return { name: routeNames.dashboard }
  }

  return true
})

router.afterEach((to) => {
  const title = typeof to.meta.title === 'string' ? to.meta.title : 'Admin'
  document.title = `${title} · Mall Admin`
})
