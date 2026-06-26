import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:3300',
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: false,
  },
  // 开发模式下 VITE_API_BASE 设为 /api/v1 走代理
  // 生产构建时 .env.production 指向真实后端 https://api.xisu.leisureea.cn/api/v1
})

