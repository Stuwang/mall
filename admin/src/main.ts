import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import zhCn from 'element-plus/es/locale/lang/zh-cn'

import App from './App.vue'
import { appContext } from './app/AppContext'
import { router } from './router'
import { routeNames } from './router/routes'
import { useAuthStore } from './stores/authStore'
import './styles/index.css'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)
app.use(ElementPlus, { locale: zhCn })

const authStore = useAuthStore()
appContext.api.http.setUnauthorizedHandler(() => {
  authStore.forceLogout()
  if (router.currentRoute.value.name !== routeNames.login) {
    void router.replace({ name: routeNames.login })
  }
})

app.mount('#app')
