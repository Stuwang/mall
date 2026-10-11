import { fileURLToPath, URL } from 'node:url'
import vue from '@vitejs/plugin-vue'
import { defineConfig } from 'vite'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    port: 5173,
    // 前缀必须落到 /xxx/api，避免 /users、/products、/orders 刷新被代理成后端 404
    proxy: {
      '/user/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/admin': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/product/api': {
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
    },
  },
})
