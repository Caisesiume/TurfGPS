import { defineConfig } from 'vitest/config'
import { loadEnv } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import { LOCAL_SERVICE, parseApiBaseUrl } from './src/apiBaseUrl.ts'

// No dev proxy: under ADR-0006 § D4 the browser calls the service cross-origin
// in every mode, so the dev server models the deployment rather than a
// same-origin arrangement nothing deploys.
export default defineConfig(({ command, mode }) => {
  const configured = loadEnv(mode, process.cwd(), 'VITE_').VITE_API_BASE_URL?.trim()
  // A malformed value fails here, in every mode. An unset one falls back to the
  // local service, because the gate build runs with no environment; a deploy
  // build must set it, so the fallback is announced rather than silent, and
  // the card renders the base it was built with.
  if (command === 'build' && !configured) {
    console.warn(`VITE_API_BASE_URL is unset: this build calls ${LOCAL_SERVICE}. Set it for a deploy.`)
  }
  const apiBaseUrl = parseApiBaseUrl(configured || LOCAL_SERVICE)

  return {
    plugins: [react(), tailwindcss()],
    define: { 'import.meta.env.VITE_API_BASE_URL': JSON.stringify(apiBaseUrl) },
    test: { environment: 'jsdom', include: ['src/**/*.test.{ts,tsx}', 'test/**/*.test.ts'] },
  }
})
