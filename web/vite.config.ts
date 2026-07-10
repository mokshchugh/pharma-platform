import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/telemetry': 'http://localhost:8081',
      '/plcs': 'http://localhost:8081',
      '/tags': 'http://localhost:8081',
      '/health': 'http://localhost:8081',
    },
  },
})
