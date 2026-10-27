import { createApp } from 'vue'
import { createPinia } from 'pinia'
import 'vant/lib/index.css'

import App from './App.vue'
import { appContext } from './app/AppContext'
import { router } from './router'
import { routeNames } from './router/routes'
import { useAuthStore } from './stores/authStore'
import { useCartStore } from './stores/cartStore'
import './styles/index.css'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)

const authStore = useAuthStore()
const cartStore = useCartStore()

appContext.api.http.setUnauthorizedHandler(() => {
  authStore.forceLogout()
  cartStore.clearLocal()
  if (router.currentRoute.value.name !== routeNames.login) {
    void router.replace({
      name: routeNames.login,
      query: { redirect: router.currentRoute.value.fullPath },
    })
  }
})

authStore.bootstrap()

app.mount('#app')
