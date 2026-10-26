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
