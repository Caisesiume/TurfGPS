/// <reference types="vite/client" />

interface ImportMetaEnv {
  // Always a validated absolute URL: `vite.config.ts` defines it.
  readonly VITE_API_BASE_URL: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
