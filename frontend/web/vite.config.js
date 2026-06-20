import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'
import { fileURLToPath } from 'url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      src: path.resolve(__dirname, './src'),
      // 共用核心（純 TS）；Vite 透過 esbuild 直接轉譯。
      '@shared': path.resolve(__dirname, '../../shared/src'),
    },
  },
  server: {
    port: 3000,
    // 允許讀取專案根目錄下的 shared 共用核心。
    fs: {
      allow: [path.resolve(__dirname, '../../')],
    },
  },
})
