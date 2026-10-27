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
    // 监听所有网卡，局域网用 http://192.168.96.132:5175 打开；localhost 仍可用。
    host: true,
    // 默认 CORS 只放行 localhost。用 IP 打开时，浏览器拉 ES module 会带上该 Origin。
    cors: true,
    port: 5175,
    proxy: {
      '/user/api': {
        target: 'http://localhost:8090',
        changeOrigin: true,
      },
      '/im/api': {
        target: 'http://localhost:8090',
        changeOrigin: true,
      },
      '/im/ws': {
        target: 'ws://localhost:8090',
        ws: true,
        changeOrigin: true,
      },
      '/files': {
        target: 'http://localhost:8090',
        changeOrigin: true,
      },
    },
  },
})
