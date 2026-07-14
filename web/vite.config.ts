import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import path from 'path'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    proxy: {
      '/telemetry': 'http://localhost:8081',
      '/plcs': 'http://localhost:8081',
      '/tags': 'http://localhost:8081',
      '/health': 'http://localhost:8081',
      '/api': 'http://localhost:8081',
    },
  },
})
