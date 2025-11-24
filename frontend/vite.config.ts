import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src')
    }
  },
  server: {
    port: 3000,
    host: '0.0.0.0', // 允许外部访问
    allowedHosts: [
      'project.smartxy.com.cn',
      'ungeneralising-harlow-orthogonally.ngrok-free.dev',
      'localhost',
      '127.0.0.1'
    ],
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
        // 不再去掉 /api 前缀，因为后端路由现在统一使用 /api 前缀
      },
      '/uploads': {
        target: 'http://localhost:8080',
        changeOrigin: true
        // 代理上传文件的静态服务
      }
    }
  }
})
