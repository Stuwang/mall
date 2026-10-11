import { fileURLToPath, URL } from 'node:url'
import vue from '@vitejs/plugin-vue'
import { VantResolver } from '@vant/auto-import-resolver'
import Components from 'unplugin-vue-components/vite'
import { defineConfig } from 'vite'

export default defineConfig({
  plugins: [
    vue(),
    Components({
      resolvers: [VantResolver()],
    }),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    port: 5174,
    // 仅代理后端 API 前缀，避免与前端页面路由（如 /cart、/product/:id）冲突；
    // 刷新页面时否则会被代理到后端，导致 SPA 无法回落 index.html。
    proxy: {
      '/user/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/product/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/cart/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/order/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/asset/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/mall/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
