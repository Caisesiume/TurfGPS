import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

// Nothing below ships. The build output is the deployed artefact — a directory
// of static files — and `DEPLOYMENT.md § Deployment architecture` puts it behind
// one reverse proxy that serves those files and forwards /api to the service on
// the same origin. This proxy exists only so the dev and preview servers
// reproduce that single origin locally; the prefix is stripped because the
// service's only route today is its root.
//
// The key is a regex and not the string '/api' because vite matches a string
// key with url.startsWith — so '/api' also matches /apifoo, and the strip below
// would hand the service `foo`. That is worth a regex in a file that ships
// nothing, because this block is what a deployed proxy gets modelled on, and
// `DEPLOYMENT.md § Where the deployment configuration lives` makes that proxy a
// candidate enforcement point for the plan-lookup throttle. nginx's
// `location /api` is unanchored in exactly the same way, so a throttle scoped
// to the lookup path is bypassed by asking for the same upstream route without
// the boundary. Anything copied from here anchors on the segment boundary too.
const apiProxy = {
  '^/api($|[/?])': {
    target: 'http://127.0.0.1:8080',
    // The match makes what is stripped a whole segment, leaving '', '/…' or
    // '?…' behind; the leading slash is restored because a proxied request
    // still needs a path. On the path this is the mapping the NFR-005 harness
    // gives its origin — /api and /api/ both reach the root — which is what
    // lets that harness stand in for this rule. A query is carried through
    // here; that harness forwards a pathname only, so it exercises none.
    rewrite: (path: string) => `/${path.replace(/^\/api\/?/, '')}`,
  },
}

export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: { proxy: apiProxy },
  preview: { proxy: apiProxy },
  test: { environment: 'jsdom' },
})
