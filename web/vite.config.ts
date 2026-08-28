import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

// Nothing below ships. The build output is the deployed artefact — a directory
// of static files — and `DEPLOYMENT.md § Deployment architecture` puts it behind
// one reverse proxy that serves those files and forwards /api to the service on
// the same origin. This proxy exists only so the dev and preview servers
// reproduce that single origin locally; the prefix is stripped because the
// service's only route today is its root.
const apiProxy = {
  '/api': {
    target: 'http://127.0.0.1:8080',
    rewrite: (path: string) => path.replace(/^\/api/, ''),
  },
}

export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: { proxy: apiProxy },
  preview: { proxy: apiProxy },
  test: { environment: 'jsdom' },
})
