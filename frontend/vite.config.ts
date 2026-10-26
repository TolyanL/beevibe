import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import vueDevTools from 'vite-plugin-vue-devtools'

const backend = process.env.VITE_BACKEND_URL ?? 'http://127.0.0.1:8080'
const proxy = {
  '/api': {
    target: backend,
    changeOrigin: false,
  },
  '^/g/': {
    target: backend,
    changeOrigin: false,
  },
}

export default defineConfig({
  plugins: [vue(), tailwindcss(), vueDevTools()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    port: 5173,
    proxy,
  },
  preview: {
    port: 4173,
    proxy,
  },
})
