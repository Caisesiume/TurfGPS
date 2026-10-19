// Where the client reaches the service, resolved once, at build time.
//
// `docs/adr/ADR-0006-deployment-topology.md` § D3 and § D4: the client is static
// files on Vercel, and the browser calls the service on the API's own subdomain,
// cross-origin under CORS, never through a Vercel rewrite. So the base must be
// an absolute http(s) URL. A relative one such as `/api` resolves against the
// client's own host, and reaches the service only through the rewrite § D4
// forbids, so it is refused here rather than deployed.

// The service listens on :8080 (`service/cmd/turfgps/main.go`). An unset
// VITE_API_BASE_URL falls back to it, except on Vercel; `resolveApiBaseUrl`
// is the rule.
export const LOCAL_SERVICE = 'http://127.0.0.1:8080'

// resolveApiBaseUrl is the build-time rule. A set value must pass
// `parseApiBaseUrl`. An unset one refuses a Vercel build, which Vercel marks by
// setting VERCEL, so a deploy cannot ship a client pointed at loopback; any
// other build falls back to the local service, because the gate and CI builds
// run with no environment.
export function resolveApiBaseUrl(raw: string | undefined, onVercel: boolean): string {
  if (raw?.trim()) {
    return parseApiBaseUrl(raw)
  }
  if (onVercel) {
    throw new Error('VITE_API_BASE_URL is unset on a Vercel build (VERCEL is set): name the service, e.g. https://api.example.org')
  }
  return LOCAL_SERVICE
}

// parseApiBaseUrl returns the base with no trailing slash, or throws naming
// what is wrong with the configured value.
export function parseApiBaseUrl(raw: string | undefined): string {
  const value = raw?.trim() ?? ''
  if (value === '') {
    throw new Error('VITE_API_BASE_URL is empty: name the service, e.g. https://api.example.org')
  }
  let url: URL
  try {
    url = new URL(value)
  } catch {
    throw new Error(`VITE_API_BASE_URL must be an absolute URL, not "${value}" (ADR-0006 § D4: no rewrite)`)
  }
  if (url.protocol !== 'http:' && url.protocol !== 'https:') {
    throw new Error(`VITE_API_BASE_URL must use http or https, not "${url.protocol}"`)
  }
  if (url.username !== '' || url.password !== '' || url.search !== '' || url.hash !== '') {
    throw new Error('VITE_API_BASE_URL must carry no credentials, query or fragment')
  }
  return `${url.origin}${url.pathname.replace(/\/+$/, '')}`
}
